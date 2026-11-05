package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kaptinlin/jsonschema"
	"gopkg.in/yaml.v3"
)

const (
	jsonschemaParseFunctionName    = "jsonschema_parse"
	jsonschemaValidateFunctionName = "jsonschema_validate"
	jsonSchemaEvaluationFailed     = "evaluation failed"
)

type jsonSchemaValidationError struct {
	details string
}

type jsonSchemaSourceRole string

const (
	jsonSchemaSourceRoleSchema jsonSchemaSourceRole = "schema"
	jsonSchemaSourceRoleTarget jsonSchemaSourceRole = "target"
)

func (r jsonSchemaSourceRole) sourceLabel() string {
	return fmt.Sprintf("%s source", r)
}

type jsonSchemaSourceKind string

const (
	jsonSchemaSourceKindURL    jsonSchemaSourceKind = "url"
	jsonSchemaSourceKindFile   jsonSchemaSourceKind = "file"
	jsonSchemaSourceKindInline jsonSchemaSourceKind = "inline"
)

type jsonSchemaParseFormat string

const (
	jsonSchemaParseFormatJSON jsonSchemaParseFormat = "json"
	jsonSchemaParseFormatYAML jsonSchemaParseFormat = "yaml"
)

type jsonSchemaResolvedSource struct {
	data []byte
	kind jsonSchemaSourceKind
}

type jsonSchemaParsedDocument struct {
	value       interface{}
	parseFormat jsonSchemaParseFormat
}

func (e *jsonSchemaValidationError) Error() string {
	return fmt.Sprintf("schema validation failed: %s", e.details)
}

func newJSONSchemaValidationError(validationResult *jsonschema.EvaluationResult) *jsonSchemaValidationError {
	return &jsonSchemaValidationError{details: formatJSONSchemaValidationDetails(validationResult)}
}

func formatJSONSchemaValidationDetails(validationResult *jsonschema.EvaluationResult) string {
	if validationResult == nil {
		return jsonSchemaEvaluationFailed
	}

	detailedErrors := validationResult.DetailedErrors()
	if len(detailedErrors) == 0 {
		return jsonSchemaEvaluationFailed
	}

	formattedErrors := make([]string, 0, len(detailedErrors))
	for path, message := range detailedErrors {
		errorPath := strings.TrimSpace(path)
		if errorPath == "" {
			errorPath = "<root>"
		}

		errorMessage := strings.TrimSpace(message)
		if errorMessage == "" {
			errorMessage = jsonSchemaEvaluationFailed
		}

		formattedErrors = append(formattedErrors, fmt.Sprintf("%s: %s", errorPath, errorMessage))
	}

	sort.Strings(formattedErrors)

	return strings.Join(formattedErrors, "; ")
}

func readJSONSchemaSources(ctx context.Context, request function.RunRequest) (string, string, error) {
	var schemaSource types.String
	var targetSource types.String

	err := request.Arguments.Get(ctx, &schemaSource, &targetSource)
	if err != nil {
		return "", "", fmt.Errorf("error reading function arguments: %w", err)
	}

	return schemaSource.ValueString(), targetSource.ValueString(), nil
}

func jsonSchemaSourceParameters() []function.Parameter {
	return []function.Parameter{
		function.StringParameter{
			Name:               "schema_source",
			Description:        "JSON Schema source: URL, file path, or inline JSON/YAML schema",
			AllowNullValue:     false,
			AllowUnknownValues: false,
		},
		function.StringParameter{
			Name:               "target_source",
			Description:        "Target source: URL, file path, or inline JSON/YAML value",
			AllowNullValue:     false,
			AllowUnknownValues: false,
		},
	}
}

func processJSONSchemaParse(ctx context.Context, schemaSource string, targetSource string) (interface{}, error) {
	schemaResolvedSource, err := resolveSchemaOrTargetSource(ctx, schemaSource, jsonSchemaSourceRoleSchema)
	if err != nil {
		return nil, err
	}

	schemaParsedDocument, err := parseStructuredDocument(schemaResolvedSource.data, jsonSchemaSourceRoleSchema)
	if err != nil {
		return nil, err
	}

	schemaObject, ok := schemaParsedDocument.value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema source must resolve to an object")
	}

	compiledSchema, err := compileJSONSchemaDocument(ctx, schemaObject)
	if err != nil {
		return nil, err
	}

	targetResolvedSource, err := resolveSchemaOrTargetSource(ctx, targetSource, jsonSchemaSourceRoleTarget)
	if err != nil {
		return nil, err
	}

	targetParsedDocument, err := parseStructuredDocument(targetResolvedSource.data, jsonSchemaSourceRoleTarget)
	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "Schema and target sources parsed", map[string]interface{}{
		"stage":               "sources_parsed",
		"schema_source_kind":  string(schemaResolvedSource.kind),
		"target_source_kind":  string(targetResolvedSource.kind),
		"schema_parse_format": string(schemaParsedDocument.parseFormat),
		"target_parse_format": string(targetParsedDocument.parseFormat),
	})

	validationResult := compiledSchema.Validate(targetParsedDocument.value)
	if !validationResult.IsValid() {
		return nil, newJSONSchemaValidationError(validationResult)
	}

	defaultedTarget := applyDefaults(schemaObject, targetParsedDocument.value)
	defaultedValidationResult := compiledSchema.Validate(defaultedTarget)
	if !defaultedValidationResult.IsValid() {
		return nil, newJSONSchemaValidationError(defaultedValidationResult)
	}

	return defaultedTarget, nil
}

func processJSONSchemaValidate(ctx context.Context, schemaSource string, targetSource string) (bool, error) {
	_, err := processJSONSchemaParse(ctx, schemaSource, targetSource)
	if err == nil {
		return true, nil
	}

	if _, ok := errors.AsType[*jsonSchemaValidationError](err); ok {
		return false, nil
	}

	return false, err
}

func resolveSchemaOrTargetSource(ctx context.Context, source string, sourceRole jsonSchemaSourceRole) (jsonSchemaResolvedSource, error) {
	trimmedSource := strings.TrimSpace(source)
	if trimmedSource == "" {
		return jsonSchemaResolvedSource{}, fmt.Errorf("%s cannot be empty", sourceRole.sourceLabel())
	}

	parsedURL, parseURLErr := url.ParseRequestURI(trimmedSource)
	if parseURLErr == nil && (parsedURL.Scheme == "http" || parsedURL.Scheme == "https") {
		urlContent, err := readURLSource(ctx, trimmedSource, sourceRole)
		if err != nil {
			return jsonSchemaResolvedSource{kind: jsonSchemaSourceKindURL}, err
		}

		return jsonSchemaResolvedSource{data: urlContent, kind: jsonSchemaSourceKindURL}, nil
	}

	fileContent, err := readFileSource(trimmedSource)
	if err == nil {
		return jsonSchemaResolvedSource{data: fileContent, kind: jsonSchemaSourceKindFile}, nil
	}

	if strings.ContainsAny(trimmedSource, "\r\n") ||
		strings.HasPrefix(trimmedSource, "{") ||
		strings.HasPrefix(trimmedSource, "[") ||
		strings.HasPrefix(trimmedSource, "-") ||
		strings.Contains(trimmedSource, ":") {
		return jsonSchemaResolvedSource{data: []byte(trimmedSource), kind: jsonSchemaSourceKindInline}, nil
	}

	looksLikeFilePath := filepath.IsAbs(trimmedSource) || strings.HasPrefix(trimmedSource, "./") || strings.HasPrefix(trimmedSource, "../") ||
		strings.Contains(trimmedSource, "/") || strings.Contains(trimmedSource, `\\`)
	if !looksLikeFilePath {
		fileExtension := strings.ToLower(filepath.Ext(trimmedSource))
		looksLikeFilePath = fileExtension == ".json" || fileExtension == ".yaml" || fileExtension == ".yml"
	}

	if looksLikeFilePath {
		return jsonSchemaResolvedSource{kind: jsonSchemaSourceKindFile}, fmt.Errorf("error reading %s '%s': %w", sourceRole.sourceLabel(), trimmedSource, err)
	}

	return jsonSchemaResolvedSource{data: []byte(trimmedSource), kind: jsonSchemaSourceKindInline}, nil
}

func parseStructuredDocument(data []byte, sourceRole jsonSchemaSourceRole) (jsonSchemaParsedDocument, error) {
	var parsed interface{}

	jsonErr := json.Unmarshal(data, &parsed)
	if jsonErr == nil {
		normalizedData := normalizeGenericData(parsed)
		return jsonSchemaParsedDocument{value: normalizedData, parseFormat: jsonSchemaParseFormatJSON}, nil
	}

	yamlErr := yaml.Unmarshal(data, &parsed)
	if yamlErr == nil {
		normalizedData := normalizeGenericData(parsed)
		return jsonSchemaParsedDocument{value: normalizedData, parseFormat: jsonSchemaParseFormatYAML}, nil
	}

	return jsonSchemaParsedDocument{}, fmt.Errorf("%s is not valid JSON or YAML (json: %v, yaml: %v)", sourceRole.sourceLabel(), jsonErr, yamlErr)
}

func compileJSONSchemaDocument(ctx context.Context, schemaObject map[string]interface{}) (*jsonschema.Schema, error) {
	schemaJSON, err := json.Marshal(schemaObject)
	if err != nil {
		tflog.Error(ctx, "Failed to marshal schema document", map[string]interface{}{
			"stage":      "compile_schema",
			"error_kind": "schema_marshal_failed",
			"error":      err.Error(),
		})
		return nil, fmt.Errorf("error marshaling schema document: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	compiledSchema, err := compiler.Compile(schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("error compiling schema: %w", err)
	}

	return compiledSchema, nil
}

func readURLSource(ctx context.Context, sourceURL string, sourceRole jsonSchemaSourceRole) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing %s URL request '%s': %w", sourceRole.sourceLabel(), sourceURL, err)
	}

	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(request)
	if err != nil {
		return nil, fmt.Errorf("error requesting %s URL '%s': %w", sourceRole.sourceLabel(), sourceURL, err)
	}
	// Closing the response body is best-effort; status and read errors are handled separately.
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("error requesting %s URL '%s': unexpected status code %d", sourceRole.sourceLabel(), sourceURL, response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		urlHost := ""
		parsedURL, parseErr := url.Parse(sourceURL)
		if parseErr == nil {
			urlHost = parsedURL.Host
		}

		tflog.Warn(ctx, "Failed to read URL response body", map[string]interface{}{
			"stage":        "request_url",
			"source_label": string(sourceRole),
			"source_kind":  string(jsonSchemaSourceKindURL),
			"url_host":     urlHost,
			"status_code":  response.StatusCode,
			"failure_type": "response_read_failed",
			"error":        err.Error(),
		})
		return nil, fmt.Errorf("error reading %s URL response '%s': %w", sourceRole.sourceLabel(), sourceURL, err)
	}

	return responseBody, nil
}

func readFileSource(path string) ([]byte, error) {
	fileContent, err := os.ReadFile(path)
	if err == nil {
		return fileContent, nil
	}

	if filepath.IsAbs(path) {
		return nil, err
	}

	candidateRoots := []string{
		os.Getenv("PWD"),
		os.Getenv("TF_WORKING_DIR"),
		os.Getenv("INIT_CWD"),
	}

	for _, root := range candidateRoots {
		if root == "" {
			continue
		}

		candidatePath := filepath.Join(root, path)
		candidateContent, candidateErr := os.ReadFile(candidatePath)
		if candidateErr == nil {
			return candidateContent, nil
		}
	}

	return nil, err
}

func applyDefaults(schema interface{}, value interface{}) interface{} {
	return applyDefaultsWithRoot(schema, schema, value, map[string]bool{})
}

func applyDefaultsWithRoot(root interface{}, schema interface{}, value interface{}, materializedRefs map[string]bool) interface{} {
	return applyDefaultsWithRefChain(root, schema, value, materializedRefs, map[string]bool{})
}

func applyDefaultsWithRefChain(root interface{}, schema interface{}, value interface{}, materializedRefs map[string]bool, refChain map[string]bool) interface{} {
	schemaObject, ok := schema.(map[string]interface{})
	if !ok {
		return value
	}

	root = jsonSchemaResourceRoot(root, schemaObject)
	schemaKey := fmt.Sprintf("%p", schemaObject)
	if refChain[schemaKey] {
		return value
	}
	refChain[schemaKey] = true
	defer delete(refChain, schemaKey)

	if reference, hasReference := schemaObject["$ref"].(string); hasReference {
		if referencedSchema, referencedRoot, found := resolveJSONSchemaReference(root, reference); found {
			value = applyDefaultsWithRefChain(
				referencedRoot,
				referencedSchema,
				value,
				materializedRefs,
				refChain,
			)
		}
	}

	if allOfSchemas, hasAllOf := schemaObject["allOf"].([]interface{}); hasAllOf {
		for _, allOfSchema := range allOfSchemas {
			value = applyDefaultsWithRefChain(root, allOfSchema, value, materializedRefs, refChain)
		}
	}

	if isObjectSchema(schemaObject) {
		return applyObjectDefaults(
			root,
			schemaObject,
			value,
			materializedRefs,
		)
	}

	if itemSchema, hasItems := schemaObject["items"]; hasItems {
		return applyArrayDefaults(
			root,
			itemSchema,
			value,
			materializedRefs,
		)
	}

	return value
}

func applyObjectDefaults(
	root interface{},
	schemaObject map[string]interface{},
	value interface{},
	materializedRefs map[string]bool,
) interface{} {
	current, ok := value.(map[string]interface{})
	if !ok {
		return value
	}

	objectValue := make(map[string]interface{}, len(current))
	maps.Copy(objectValue, current)
	properties, _ := schemaObject["properties"].(map[string]interface{})
	for propertyName, propertySchema := range properties {
		currentValue, exists := objectValue[propertyName]
		switch {
		case !exists:
			defaultValue, shouldSet := materializedDefaultForMissingPropertyWithRoot(root, propertySchema, materializedRefs)
			if shouldSet {
				objectValue[propertyName] = defaultValue
			}
		case currentValue != nil:
			objectValue[propertyName] = applyDefaultsWithRoot(
				root,
				propertySchema,
				currentValue,
				materializedRefs,
			)
		}
	}

	additionalPropertiesSchema, hasAdditionalSchema := schemaObject["additionalProperties"].(map[string]interface{})
	if hasAdditionalSchema {
		for key, nestedValue := range objectValue {
			if _, declaredProperty := properties[key]; declaredProperty || nestedValue == nil {
				continue
			}
			objectValue[key] = applyDefaultsWithRoot(
				root,
				additionalPropertiesSchema,
				nestedValue,
				materializedRefs,
			)
		}
	}

	return objectValue
}

func applyArrayDefaults(
	root interface{},
	itemSchema interface{},
	value interface{},
	materializedRefs map[string]bool,
) interface{} {
	arrayValue, isArray := value.([]interface{})
	if !isArray {
		return value
	}

	defaultedArray := make([]interface{}, len(arrayValue))
	for index, item := range arrayValue {
		if item != nil {
			defaultedArray[index] = applyDefaultsWithRoot(
				root,
				itemSchema,
				item,
				materializedRefs,
			)
		}
	}

	return defaultedArray
}

func materializedDefaultForMissingProperty(propertySchema interface{}) (interface{}, bool) {
	return materializedDefaultForMissingPropertyWithRoot(propertySchema, propertySchema, map[string]bool{})
}

func materializedDefaultForMissingPropertyWithRoot(root interface{}, propertySchema interface{}, materializedRefs map[string]bool) (interface{}, bool) {
	propertySchemaObject, ok := propertySchema.(map[string]interface{})
	if !ok {
		return nil, false
	}

	root = jsonSchemaResourceRoot(root, propertySchemaObject)
	// Schema identity also catches cycles through allOf without conflating resource-local references.
	schemaKey := fmt.Sprintf("%p", propertySchemaObject)
	if materializedRefs[schemaKey] {
		return nil, false
	}
	materializedRefs[schemaKey] = true
	defer delete(materializedRefs, schemaKey)

	if propertyDefault, hasDefault := propertySchemaObject["default"]; hasDefault {
		defaultedValue := applyDefaultsWithRoot(root, propertySchemaObject, deepCopyValue(propertyDefault), materializedRefs)
		return defaultedValue, true
	}

	if reference, hasReference := propertySchemaObject["$ref"].(string); hasReference {
		referencedSchema, referencedRoot, found := resolveJSONSchemaReference(root, reference)
		if !found {
			return nil, false
		}
		if defaultValue, shouldSet := materializedDefaultForMissingPropertyWithRoot(referencedRoot, referencedSchema, materializedRefs); shouldSet {
			return applyDefaultsWithRoot(root, propertySchemaObject, defaultValue, materializedRefs), true
		}
	}

	return nil, false
}

func jsonSchemaResourceRoot(root interface{}, schema interface{}) interface{} {
	if schemaObject, ok := schema.(map[string]interface{}); ok {
		if id, ok := schemaObject["$id"].(string); ok && id != "" {
			return schemaObject
		}
	}
	return root
}

func resolveJSONSchemaReference(root interface{}, reference string) (interface{}, interface{}, bool) {
	if !strings.HasPrefix(reference, "#") {
		return nil, nil, false
	}
	fragment, err := url.PathUnescape(strings.TrimPrefix(reference, "#"))
	if err != nil {
		return nil, nil, false
	}
	if fragment == "" {
		return root, root, true
	}
	if !strings.HasPrefix(fragment, "/") {
		return nil, nil, false
	}

	current := root
	for token := range strings.SplitSeq(strings.TrimPrefix(fragment, "/"), "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch value := current.(type) {
		case map[string]interface{}:
			current, _ = value[token]
			if current == nil {
				return nil, nil, false
			}
		case []interface{}:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(value) {
				return nil, nil, false
			}
			current = value[index]
		default:
			return nil, nil, false
		}
		root = jsonSchemaResourceRoot(root, current)
	}

	return current, root, true
}

func isObjectSchema(schemaObject map[string]interface{}) bool {
	if schemaType, hasType := schemaObject["type"]; hasType {
		switch typedValue := schemaType.(type) {
		case string:
			if typedValue == "object" {
				return true
			}
		case []interface{}:
			for _, item := range typedValue {
				if asString, ok := item.(string); ok && asString == "object" {
					return true
				}
			}
		}
	}

	_, hasProperties := schemaObject["properties"]
	return hasProperties
}

func deepCopyValue(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		copyMap := make(map[string]interface{}, len(typedValue))
		for key, nestedValue := range typedValue {
			copyMap[key] = deepCopyValue(nestedValue)
		}
		return copyMap
	case []interface{}:
		copyArray := make([]interface{}, len(typedValue))
		for index, item := range typedValue {
			copyArray[index] = deepCopyValue(item)
		}
		return copyArray
	default:
		return typedValue
	}
}

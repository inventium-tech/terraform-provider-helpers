package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kaptinlin/jsonschema"
	"gopkg.in/yaml.v3"
)

type jsonSchemaValidationError struct {
	details string
}

func (e *jsonSchemaValidationError) Error() string {
	return fmt.Sprintf("schema validation failed: %s", e.details)
}

func readJSONSchemaSources(ctx context.Context, request function.RunRequest) (string, string, error) {
	var schemaSource types.String
	var targetSource types.String

	err := request.Arguments.Get(ctx, &schemaSource, &targetSource)
	if err != nil {
		return "", "", fmt.Errorf("Error reading function arguments: %s", err.Error())
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

func processJSONSchemaParse(schemaSource string, targetSource string) (interface{}, error) {
	schemaSourceData, err := resolveSchemaOrTargetSource(schemaSource, "schema source")
	if err != nil {
		return nil, err
	}

	schemaParsed, err := parseStructuredDocument(schemaSourceData, "schema source")
	if err != nil {
		return nil, err
	}

	schemaObject, ok := schemaParsed.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema source must resolve to an object")
	}

	compiledSchema, err := compileJSONSchemaDocument(schemaObject)
	if err != nil {
		return nil, err
	}

	targetSourceData, err := resolveSchemaOrTargetSource(targetSource, "target source")
	if err != nil {
		return nil, err
	}

	targetParsed, err := parseStructuredDocument(targetSourceData, "target source")
	if err != nil {
		return nil, err
	}

	defaultedTarget := applyDefaultsFromSchema(schemaObject, targetParsed)

	validationResult := compiledSchema.Validate(defaultedTarget)
	if !validationResult.IsValid() {
		return nil, &jsonSchemaValidationError{details: validationResult.Error()}
	}

	return defaultedTarget, nil
}

func processJSONSchemaValidate(schemaSource string, targetSource string) (bool, error) {
	_, err := processJSONSchemaParse(schemaSource, targetSource)
	if err == nil {
		return true, nil
	}

	var validationErr *jsonSchemaValidationError
	if errors.As(err, &validationErr) {
		return false, nil
	}

	return false, err
}

func resolveSchemaOrTargetSource(source string, sourceLabel string) ([]byte, error) {
	trimmedSource := strings.TrimSpace(source)
	if trimmedSource == "" {
		return nil, fmt.Errorf("%s cannot be empty", sourceLabel)
	}

	if isRemoteURL(trimmedSource) {
		return readURLSource(trimmedSource, sourceLabel)
	}

	fileContent, err := readFileSource(trimmedSource)
	if err == nil {
		return fileContent, nil
	}

	if isInlineDocument(trimmedSource) {
		return []byte(trimmedSource), nil
	}

	if looksLikeFilePath(trimmedSource) {
		return nil, fmt.Errorf("error reading %s '%s': %w", sourceLabel, trimmedSource, err)
	}

	return []byte(trimmedSource), nil
}

func parseStructuredDocument(data []byte, sourceLabel string) (interface{}, error) {
	var parsed interface{}

	jsonErr := json.Unmarshal(data, &parsed)
	if jsonErr == nil {
		return normalizeGenericData(parsed), nil
	}

	yamlErr := yaml.Unmarshal(data, &parsed)
	if yamlErr == nil {
		return normalizeGenericData(parsed), nil
	}

	return nil, fmt.Errorf("%s is not valid JSON or YAML (json: %v, yaml: %v)", sourceLabel, jsonErr, yamlErr)
}

func compileJSONSchemaDocument(schemaObject map[string]interface{}) (*jsonschema.Schema, error) {
	schemaJSON, err := json.Marshal(schemaObject)
	if err != nil {
		return nil, fmt.Errorf("error marshaling schema document: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	compiledSchema, err := compiler.Compile(schemaJSON)
	if err != nil {
		return nil, fmt.Errorf("error compiling schema: %w", err)
	}

	return compiledSchema, nil
}

func isRemoteURL(value string) bool {
	parsedURL, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}

	return parsedURL.Scheme == "http" || parsedURL.Scheme == "https"
}

func readURLSource(sourceURL string, sourceLabel string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing %s URL request '%s': %w", sourceLabel, sourceURL, err)
	}

	response, err := (&http.Client{Timeout: 10 * time.Second}).Do(request)
	if err != nil {
		return nil, fmt.Errorf("error requesting %s URL '%s': %w", sourceLabel, sourceURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("error requesting %s URL '%s': unexpected status code %d", sourceLabel, sourceURL, response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading %s URL response '%s': %w", sourceLabel, sourceURL, err)
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

	candidateRoots := []string{os.Getenv("PWD"), os.Getenv("TF_WORKING_DIR"), os.Getenv("INIT_CWD")}
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

func isInlineDocument(value string) bool {
	if strings.Contains(value, "\n") || strings.Contains(value, "\r") {
		return true
	}

	if strings.HasPrefix(value, "{") || strings.HasPrefix(value, "[") || strings.HasPrefix(value, "-") {
		return true
	}

	return strings.Contains(value, ":")
}

func looksLikeFilePath(value string) bool {
	if filepath.IsAbs(value) || strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") {
		return true
	}

	if strings.Contains(value, "/") || strings.Contains(value, `\\`) {
		return true
	}

	fileExtension := strings.ToLower(filepath.Ext(value))
	return fileExtension == ".json" || fileExtension == ".yaml" || fileExtension == ".yml"
}

func applyDefaultsFromSchema(schema interface{}, value interface{}) interface{} {
	return applyDefaultsRecursively(schema, value)
}

func applyDefaultsRecursively(schema interface{}, value interface{}) interface{} {
	schemaObject, ok := schema.(map[string]interface{})
	if !ok {
		return value
	}

	if value == nil {
		if schemaDefault, hasDefault := schemaObject["default"]; hasDefault {
			value = deepCopyValue(schemaDefault)
		}
	}

	if isObjectSchema(schemaObject) {
		objectValue := map[string]interface{}{}
		switch current := value.(type) {
		case map[string]interface{}:
			for key, nestedValue := range current {
				objectValue[key] = nestedValue
			}
		case nil:
			// keep an empty object to allow nested defaults to materialize
		default:
			return value
		}

		properties, _ := schemaObject["properties"].(map[string]interface{})
		for propertyName, propertySchema := range properties {
			currentValue, exists := objectValue[propertyName]
			if !exists || currentValue == nil {
				if defaultValue, shouldSet := defaultValueForMissingProperty(propertySchema); shouldSet {
					objectValue[propertyName] = defaultValue
				}
				continue
			}

			objectValue[propertyName] = applyDefaultsRecursively(propertySchema, currentValue)
		}

		if additionalPropertiesSchema, hasAdditionalSchema := schemaObject["additionalProperties"].(map[string]interface{}); hasAdditionalSchema {
			for key, nestedValue := range objectValue {
				if _, declaredProperty := properties[key]; declaredProperty {
					continue
				}
				objectValue[key] = applyDefaultsRecursively(additionalPropertiesSchema, nestedValue)
			}
		}

		return objectValue
	}

	if itemSchema, hasItems := schemaObject["items"]; hasItems {
		arrayValue, isArray := value.([]interface{})
		if !isArray {
			return value
		}

		defaultedArray := make([]interface{}, len(arrayValue))
		for index, item := range arrayValue {
			defaultedArray[index] = applyDefaultsRecursively(itemSchema, item)
		}

		return defaultedArray
	}

	if value == nil {
		if schemaDefault, hasDefault := schemaObject["default"]; hasDefault {
			return deepCopyValue(schemaDefault)
		}
	}

	return value
}

func defaultValueForMissingProperty(propertySchema interface{}) (interface{}, bool) {
	propertySchemaObject, ok := propertySchema.(map[string]interface{})
	if !ok {
		return nil, false
	}

	if propertyDefault, hasDefault := propertySchemaObject["default"]; hasDefault {
		defaultedValue := applyDefaultsRecursively(propertySchemaObject, deepCopyValue(propertyDefault))
		return defaultedValue, true
	}

	if isObjectSchema(propertySchemaObject) {
		materializedObject := applyDefaultsRecursively(propertySchemaObject, map[string]interface{}{})
		materializedObjectMap, isMap := materializedObject.(map[string]interface{})
		if isMap && len(materializedObjectMap) > 0 {
			return materializedObjectMap, true
		}
	}

	return nil, false
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

func normalizeGenericData(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		normalized := make(map[string]interface{}, len(typedValue))
		for key, nestedValue := range typedValue {
			normalized[key] = normalizeGenericData(nestedValue)
		}
		return normalized
	case map[interface{}]interface{}:
		normalized := make(map[string]interface{}, len(typedValue))
		for key, nestedValue := range typedValue {
			normalized[fmt.Sprintf("%v", key)] = normalizeGenericData(nestedValue)
		}
		return normalized
	case []interface{}:
		normalized := make([]interface{}, len(typedValue))
		for index, item := range typedValue {
			normalized[index] = normalizeGenericData(item)
		}
		return normalized
	default:
		return typedValue
	}
}

func convertToTerraformDynamicValue(ctx context.Context, data interface{}) (basetypes.DynamicValue, error) {
	terraformValue, err := convertInterfaceToTerraformValue(ctx, normalizeGenericData(data))
	if err != nil {
		return basetypes.DynamicValue{}, fmt.Errorf("failed to convert to Terraform value: %w", err)
	}

	return basetypes.NewDynamicValue(terraformValue), nil
}

func convertInterfaceToTerraformValue(ctx context.Context, data interface{}) (attr.Value, error) {
	if data == nil {
		return types.DynamicNull(), nil
	}

	switch typedValue := data.(type) {
	case bool:
		return types.BoolValue(typedValue), nil
	case int:
		return types.Int64Value(int64(typedValue)), nil
	case int8:
		return types.Int64Value(int64(typedValue)), nil
	case int16:
		return types.Int64Value(int64(typedValue)), nil
	case int32:
		return types.Int64Value(int64(typedValue)), nil
	case int64:
		return types.Int64Value(typedValue), nil
	case uint:
		return types.Int64Value(int64(typedValue)), nil
	case uint8:
		return types.Int64Value(int64(typedValue)), nil
	case uint16:
		return types.Int64Value(int64(typedValue)), nil
	case uint32:
		return types.Int64Value(int64(typedValue)), nil
	case uint64:
		if typedValue > math.MaxInt64 {
			return types.DynamicNull(), fmt.Errorf("unsigned integer value %d overflows int64", typedValue)
		}
		return types.Int64Value(int64(typedValue)), nil
	case float32:
		return convertFloatToTerraformNumber(float64(typedValue)), nil
	case float64:
		return convertFloatToTerraformNumber(typedValue), nil
	case string:
		return types.StringValue(typedValue), nil
	case map[string]interface{}:
		attributeTypes := make(map[string]attr.Type, len(typedValue))
		attributeValues := make(map[string]attr.Value, len(typedValue))

		for key, nestedValue := range typedValue {
			convertedValue, err := convertInterfaceToTerraformValue(ctx, nestedValue)
			if err != nil {
				return types.DynamicNull(), fmt.Errorf("failed to convert map value for key '%s': %w", key, err)
			}
			attributeTypes[key] = convertedValue.Type(ctx)
			attributeValues[key] = convertedValue
		}

		objectValue, diags := types.ObjectValue(attributeTypes, attributeValues)
		if diags.HasError() {
			return types.DynamicNull(), fmt.Errorf("failed to create object value: %s", diags.Errors())
		}

		return objectValue, nil
	case []interface{}:
		if len(typedValue) == 0 {
			return types.ListValueMust(types.DynamicType, []attr.Value{}), nil
		}

		elements := make([]attr.Value, len(typedValue))
		for index, item := range typedValue {
			convertedValue, err := convertInterfaceToTerraformValue(ctx, item)
			if err != nil {
				return types.DynamicNull(), fmt.Errorf("failed to convert array element at index %d: %w", index, err)
			}
			elements[index] = basetypes.NewDynamicValue(convertedValue)
		}

		listValue, diags := types.ListValue(types.DynamicType, elements)
		if diags.HasError() {
			return types.DynamicNull(), fmt.Errorf("failed to create list value: %s", diags.Errors())
		}

		return listValue, nil
	default:
		return types.DynamicNull(), fmt.Errorf("unsupported data type: %T", data)
	}
}

func convertFloatToTerraformNumber(value float64) attr.Value {
	if value >= math.MinInt64 && value <= math.MaxInt64 && math.Trunc(value) == value {
		return types.Int64Value(int64(value))
	}

	return types.Float64Value(value)
}

package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveSchemaOrTargetSource(t *testing.T) {
	t.Run("empty input returns role-aware error", func(t *testing.T) {
		_, err := resolveSchemaOrTargetSource(context.Background(), "   ", jsonSchemaSourceRoleSchema)
		if err == nil {
			t.Fatal("expected error for empty schema source")
		}

		if !strings.Contains(err.Error(), "schema source cannot be empty") {
			t.Fatalf("expected empty source error, got: %v", err)
		}
	})

	t.Run("inline JSON and YAML are detected", func(t *testing.T) {
		testCases := []struct {
			name   string
			source string
		}{
			{name: "inline JSON", source: `{"type":"object"}`},
			{name: "inline JSON array", source: `[1, 2, 3]`},
			{name: "inline YAML list", source: "- item"},
			{name: "inline YAML", source: "kind: config"},
			{name: "inline multiline", source: "line1\nline2"},
			{name: "inline fallback", source: "plain-text-value"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				resolvedSource, err := resolveSchemaOrTargetSource(context.Background(), tc.source, jsonSchemaSourceRoleTarget)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if resolvedSource.kind != jsonSchemaSourceKindInline {
					t.Fatalf("expected inline source kind, got: %q", resolvedSource.kind)
				}

				if string(resolvedSource.data) != tc.source {
					t.Fatalf("expected inline data %q, got %q", tc.source, string(resolvedSource.data))
				}
			})
		}
	})

	t.Run("non-http URL-like value is not treated as URL", func(t *testing.T) {
		source := "ftp://example.com/schema.json"

		resolvedSource, err := resolveSchemaOrTargetSource(context.Background(), source, jsonSchemaSourceRoleSchema)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resolvedSource.kind != jsonSchemaSourceKindInline {
			t.Fatalf("expected inline source kind, got: %q", resolvedSource.kind)
		}

		if string(resolvedSource.data) != source {
			t.Fatalf("expected inline data %q, got %q", source, string(resolvedSource.data))
		}
	})

	t.Run("relative file path is resolved from PWD", func(t *testing.T) {
		tempDirectory := t.TempDir()
		t.Setenv("PWD", tempDirectory)

		fileName := "schema.yaml"
		fileContent := "type: object\n"
		filePath := filepath.Join(tempDirectory, fileName)

		err := os.WriteFile(filePath, []byte(fileContent), 0o644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		resolvedSource, resolveErr := resolveSchemaOrTargetSource(context.Background(), fileName, jsonSchemaSourceRoleSchema)
		if resolveErr != nil {
			t.Fatalf("unexpected error: %v", resolveErr)
		}

		if resolvedSource.kind != jsonSchemaSourceKindFile {
			t.Fatalf("expected file source kind, got: %q", resolvedSource.kind)
		}

		if string(resolvedSource.data) != fileContent {
			t.Fatalf("expected file content %q, got %q", fileContent, string(resolvedSource.data))
		}
	})

	t.Run("URL source is resolved", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			_, _ = responseWriter.Write([]byte(`{"type":"object"}`))
		}))
		defer testServer.Close()

		resolvedSource, err := resolveSchemaOrTargetSource(context.Background(), testServer.URL, jsonSchemaSourceRoleSchema)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resolvedSource.kind != jsonSchemaSourceKindURL {
			t.Fatalf("expected URL source kind, got: %q", resolvedSource.kind)
		}

		if string(resolvedSource.data) != `{"type":"object"}` {
			t.Fatalf("expected URL payload to be returned, got %q", string(resolvedSource.data))
		}
	})

	t.Run("missing file-like path returns read error shape", func(t *testing.T) {
		missingPath := filepath.Join("missing", "schema.yaml")

		resolvedSource, err := resolveSchemaOrTargetSource(context.Background(), missingPath, jsonSchemaSourceRoleSchema)
		if err == nil {
			t.Fatal("expected error for missing file path")
		}

		if resolvedSource.kind != jsonSchemaSourceKindFile {
			t.Fatalf("expected file source kind on read failure, got: %q", resolvedSource.kind)
		}

		expectedPrefix := "error reading schema source '" + missingPath + "'"
		if !strings.Contains(err.Error(), expectedPrefix) {
			t.Fatalf("expected error to contain %q, got: %v", expectedPrefix, err)
		}
	})

	t.Run("missing extension-only file-like path returns read error shape", func(t *testing.T) {
		missingPath := "schema.json"

		resolvedSource, err := resolveSchemaOrTargetSource(context.Background(), missingPath, jsonSchemaSourceRoleTarget)
		if err == nil {
			t.Fatal("expected error for missing extension-only file path")
		}

		if resolvedSource.kind != jsonSchemaSourceKindFile {
			t.Fatalf("expected file source kind on read failure, got: %q", resolvedSource.kind)
		}

		expectedPrefix := "error reading target source '" + missingPath + "'"
		if !strings.Contains(err.Error(), expectedPrefix) {
			t.Fatalf("expected error to contain %q, got: %v", expectedPrefix, err)
		}
	})

	t.Run("URL returning non-2xx status returns error", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			http.Error(responseWriter, "not found", http.StatusNotFound)
		}))
		defer testServer.Close()

		_, err := resolveSchemaOrTargetSource(context.Background(), testServer.URL, jsonSchemaSourceRoleSchema)
		if err == nil {
			t.Fatal("expected error for non-2xx HTTP response")
		}

		if !strings.Contains(err.Error(), "unexpected status code") {
			t.Fatalf("expected status code error, got: %v", err)
		}
	})
}

func TestParseStructuredDocument(t *testing.T) {
	t.Parallel()

	t.Run("parses JSON document", func(t *testing.T) {
		parsedDocument, err := parseStructuredDocument([]byte(`{"name":"api","enabled":true}`), jsonSchemaSourceRoleSchema)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parsedDocument.parseFormat != jsonSchemaParseFormatJSON {
			t.Fatalf("expected JSON parse format, got: %q", parsedDocument.parseFormat)
		}

		parsedObject, ok := parsedDocument.value.(map[string]interface{})
		if !ok {
			t.Fatalf("expected parsed JSON object, got: %T", parsedDocument.value)
		}

		if parsedObject["name"] != "api" || parsedObject["enabled"] != true {
			t.Fatalf("unexpected parsed JSON object: %#v", parsedObject)
		}
	})

	t.Run("parses YAML document", func(t *testing.T) {
		parsedDocument, err := parseStructuredDocument([]byte("name: api\nenabled: true\n"), jsonSchemaSourceRoleTarget)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parsedDocument.parseFormat != jsonSchemaParseFormatYAML {
			t.Fatalf("expected YAML parse format, got: %q", parsedDocument.parseFormat)
		}

		parsedObject, ok := parsedDocument.value.(map[string]interface{})
		if !ok {
			t.Fatalf("expected parsed YAML object, got: %T", parsedDocument.value)
		}

		if parsedObject["name"] != "api" || parsedObject["enabled"] != true {
			t.Fatalf("unexpected parsed YAML object: %#v", parsedObject)
		}
	})

	t.Run("invalid document returns role-aware error", func(t *testing.T) {
		_, err := parseStructuredDocument([]byte("{"), jsonSchemaSourceRoleTarget)
		if err == nil {
			t.Fatal("expected parsing error")
		}

		if !strings.Contains(err.Error(), "target source is not valid JSON or YAML") {
			t.Fatalf("expected invalid document message, got: %v", err)
		}
	})
}

func TestApplyDefaults(t *testing.T) {
	t.Parallel()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"app": map[string]interface{}{
				"type":    "object",
				"default": map[string]interface{}{},
				"properties": map[string]interface{}{
					"name":    map[string]interface{}{"type": "string", "default": "default-app"},
					"enabled": map[string]interface{}{"type": "boolean", "default": true},
				},
			},
		},
		"additionalProperties": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"enabled": map[string]interface{}{"type": "boolean", "default": true},
			},
		},
	}

	target := map[string]interface{}{
		"app":    map[string]interface{}{"enabled": false},
		"custom": map[string]interface{}{},
	}

	defaultedValue := applyDefaults(schema, target)
	defaultedObject, ok := defaultedValue.(map[string]interface{})
	if !ok {
		t.Fatalf("expected defaulted object map, got: %T", defaultedValue)
	}

	app, ok := defaultedObject["app"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected app object map, got: %T", defaultedObject["app"])
	}

	if app["name"] != "default-app" {
		t.Fatalf("expected app.name default, got: %#v", app["name"])
	}

	if app["enabled"] != false {
		t.Fatalf("expected app.enabled to preserve existing value, got: %#v", app["enabled"])
	}

	custom, ok := defaultedObject["custom"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected custom object map, got: %T", defaultedObject["custom"])
	}

	if custom["enabled"] != true {
		t.Fatalf("expected additionalProperties defaults to apply, got: %#v", custom)
	}
}

func TestApplyDefaultsArrayItems(t *testing.T) {
	t.Parallel()

	schema := map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"enabled": map[string]interface{}{"type": "boolean", "default": true},
			},
		},
	}

	target := []interface{}{
		map[string]interface{}{},
		map[string]interface{}{"enabled": false},
	}

	result := applyDefaults(schema, target)
	items, ok := result.([]interface{})
	if !ok {
		t.Fatalf("expected array result, got: %T", result)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got: %d", len(items))
	}

	first, firstOK := items[0].(map[string]interface{})
	if !firstOK {
		t.Fatalf("expected first item to be a map, got: %T", items[0])
	}

	if first["enabled"] != true {
		t.Fatalf("expected enabled default applied to first item, got: %#v", first["enabled"])
	}

	second, secondOK := items[1].(map[string]interface{})
	if !secondOK {
		t.Fatalf("expected second item to be a map, got: %T", items[1])
	}

	if second["enabled"] != false {
		t.Fatalf("expected enabled to remain false in second item, got: %#v", second["enabled"])
	}
}

func TestProcessJSONSchemaParseNonObjectSchemaError(t *testing.T) {
	t.Parallel()

	_, err := processJSONSchemaParse(context.Background(), `[1, 2, 3]`, `{"key": "value"}`)
	if err == nil {
		t.Fatal("expected error when schema resolves to a non-object")
	}

	if !strings.Contains(err.Error(), "schema source must resolve to an object") {
		t.Fatalf("expected non-object schema error, got: %v", err)
	}
}

func TestMaterializedDefaultForMissingProperty(t *testing.T) {
	t.Parallel()

	t.Run("non-schema input returns false", func(t *testing.T) {
		value, ok := materializedDefaultForMissingProperty("not-a-schema")
		if ok {
			t.Fatalf("expected no default value, got: %#v", value)
		}
	})

	t.Run("default value is recursively defaulted", func(t *testing.T) {
		propertySchema := map[string]interface{}{
			"type":    "object",
			"default": map[string]interface{}{"child": map[string]interface{}{}},
			"properties": map[string]interface{}{
				"child": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"enabled": map[string]interface{}{"type": "boolean", "default": true},
					},
				},
			},
		}

		value, ok := materializedDefaultForMissingProperty(propertySchema)
		if !ok {
			t.Fatal("expected default value")
		}

		asMap, mapOK := value.(map[string]interface{})
		if !mapOK {
			t.Fatalf("expected defaulted map value, got: %T", value)
		}

		child, childOK := asMap["child"].(map[string]interface{})
		if !childOK {
			t.Fatalf("expected nested child map, got: %T", asMap["child"])
		}

		if child["enabled"] != true {
			t.Fatalf("expected nested default to be materialized, got: %#v", child)
		}
	})

	t.Run("object schema without explicit default materializes nested defaults", func(t *testing.T) {
		propertySchema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"port": map[string]interface{}{"type": "integer", "default": 8080},
			},
		}

		value, ok := materializedDefaultForMissingProperty(propertySchema)
		if !ok {
			t.Fatal("expected materialized default object")
		}

		asMap := value.(map[string]interface{})
		if asMap["port"] != 8080 {
			t.Fatalf("expected port default, got: %#v", asMap)
		}
	})

	t.Run("object schema without any defaults returns false", func(t *testing.T) {
		propertySchema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{"type": "string"},
			},
		}

		value, ok := materializedDefaultForMissingProperty(propertySchema)
		if ok {
			t.Fatalf("expected no default materialization, got: %#v", value)
		}
	})
}

func TestIsObjectSchema(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		schema   map[string]interface{}
		expected bool
	}{
		{name: "explicit object type", schema: map[string]interface{}{"type": "object"}, expected: true},
		{name: "union includes object", schema: map[string]interface{}{"type": []interface{}{"null", "object"}}, expected: true},
		{name: "properties implies object", schema: map[string]interface{}{"properties": map[string]interface{}{}}, expected: true},
		{name: "non-object type", schema: map[string]interface{}{"type": "string"}, expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isObjectSchema(tc.schema); got != tc.expected {
				t.Fatalf("isObjectSchema(%#v) = %t, expected %t", tc.schema, got, tc.expected)
			}
		})
	}
}

func TestProcessJSONSchemaValidateOperationalError(t *testing.T) {
	t.Parallel()

	// An empty schema source is an operational error (not a jsonSchemaValidationError),
	// so processJSONSchemaValidate must return false AND a non-nil error.
	ok, err := processJSONSchemaValidate(context.Background(), "", `{"key": "value"}`)
	if err == nil {
		t.Fatal("expected error for empty schema source")
	}

	if ok {
		t.Fatal("expected false when an operational error occurs")
	}

	if !strings.Contains(err.Error(), "schema source cannot be empty") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestProcessJSONSchemaValidateValidationFailureReturnsFalseNoError(t *testing.T) {
	t.Parallel()

	schema := `{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`
	// target is missing the required "name" property — validation failure, not an operational error
	target := `{"age": 30}`

	ok, err := processJSONSchemaValidate(context.Background(), schema, target)
	if err != nil {
		t.Fatalf("expected no error for validation failure, got: %v", err)
	}

	if ok {
		t.Fatal("expected false for schema validation failure")
	}
}

func TestCompileJSONSchemaDocumentMarshalError(t *testing.T) {
	t.Parallel()

	// A channel value cannot be marshaled to JSON, which exercises the marshal error branch.
	schemaObject := map[string]interface{}{
		"type":    "object",
		"channel": make(chan int),
	}

	_, err := compileJSONSchemaDocument(context.Background(), schemaObject)
	if err == nil {
		t.Fatal("expected error for unmarshalable schema object")
	}

	if !strings.Contains(err.Error(), "error marshaling schema document") {
		t.Fatalf("expected marshal error, got: %v", err)
	}
}

func TestDeepCopyValue(t *testing.T) {
	t.Parallel()

	original := map[string]interface{}{
		"nested": map[string]interface{}{"name": "api"},
		"items":  []interface{}{map[string]interface{}{"enabled": true}},
	}

	copyValue := deepCopyValue(original)
	copyMap, ok := copyValue.(map[string]interface{})
	if !ok {
		t.Fatalf("expected copied map value, got: %T", copyValue)
	}

	copyMap["nested"].(map[string]interface{})["name"] = "changed"
	copyMap["items"].([]interface{})[0].(map[string]interface{})["enabled"] = false

	originalNested := original["nested"].(map[string]interface{})
	if originalNested["name"] != "api" {
		t.Fatalf("expected original nested object to remain unchanged, got: %#v", originalNested)
	}

	originalItem := original["items"].([]interface{})[0].(map[string]interface{})
	if originalItem["enabled"] != true {
		t.Fatalf("expected original list item to remain unchanged, got: %#v", originalItem)
	}
}

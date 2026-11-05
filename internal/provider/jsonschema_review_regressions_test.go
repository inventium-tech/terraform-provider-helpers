package provider

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
)

func TestJSONSchemaReviewRegressions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		schema   string
		target   string
		expected string
	}{
		{
			name: "sibling schemas apply to referenced defaults",
			schema: `{
				"type": "object",
				"$defs": {"settings": {"type": "object", "default": {}}},
				"properties": {
					"settings": {
						"$ref": "#/$defs/settings",
						"properties": {"enabled": {"type": "boolean", "default": true}},
						"allOf": [{"properties": {"mode": {"default": "safe"}}}]
					}
				}
			}`,
			target:   `{}`,
			expected: `{"settings":{"enabled":true,"mode":"safe"}}`,
		},
		{
			name: "nested resource uses its own definitions",
			schema: `{
				"type": "object",
				"$defs": {"setting": {"default": "outer"}},
				"properties": {
					"settings": {
						"$id": "https://example.com/settings",
						"type": "object", "default": {},
						"$defs": {"setting": {"default": "inner"}},
						"properties": {"value": {"$ref": "#/$defs/setting"}}
					}
				}
			}`,
			target:   `{}`,
			expected: `{"settings":{"value":"inner"}}`,
		},
		{
			name: "nested resource scope survives a pointer into it",
			schema: `{
				"type": "object",
				"$defs": {
					"setting": {"default": "outer"},
					"resource": {
						"$id": "https://example.com/settings",
						"$defs": {"setting": {"default": "inner"}},
						"properties": {
							"nested": {
								"type": "object", "default": {},
								"properties": {"value": {"$ref": "#/$defs/setting"}}
							}
						}
					}
				},
				"properties": {"settings": {"$ref": "#/$defs/resource/properties/nested"}}
			}`,
			target:   `{}`,
			expected: `{"settings":{"value":"inner"}}`,
		},
		{
			name: "encoded slash between pointer tokens",
			schema: `{
				"type": "object",
				"$defs": {"settings": {"default": "resolved"}},
				"properties": {"value": {"$ref": "#/$defs%2Fsettings"}}
			}`,
			target:   `{}`,
			expected: `{"value":"resolved"}`,
		},
		{
			name: "encoded leading slash in pointer",
			schema: `{
				"type": "object",
				"$defs": {"settings": {"default": "resolved"}},
				"properties": {"value": {"$ref": "#%2F$defs%2Fsettings"}}
			}`,
			target:   `{}`,
			expected: `{"value":"resolved"}`,
		},
		{
			name: "encoded pointer preserves escaped key characters",
			schema: `{
				"type": "object",
				"$defs": {"a/b~c%d": {"default": "resolved"}},
				"properties": {"value": {"$ref": "#%2F$defs%2Fa~1b~0c%25d"}}
			}`,
			target:   `{}`,
			expected: `{"value":"resolved"}`,
		},
		{
			name:     "multiline YAML literal containing a path",
			schema:   `{"type":"string"}`,
			target:   "|\n  /var/lib/data",
			expected: `"/var/lib/data"`,
		},
		{
			name:     "multiline YAML folded scalar containing a path",
			schema:   `{"type":"string"}`,
			target:   ">\r\n  /var/lib/data",
			expected: `"/var/lib/data"`,
		},
		{
			name: "recursive allOf defaults terminate",
			schema: `{
				"$defs": {"node": {
					"type": "object",
					"properties": {
						"enabled": {"default": true},
						"child": {"default": {}, "allOf": [{"$ref": "#/$defs/node"}]}
					}
				}},
				"$ref": "#/$defs/node"
			}`,
			target:   `{}`,
			expected: `{"enabled":true,"child":{"enabled":true}}`,
		},
		{
			name: "recursive allOf defaults preserve finite authored descendants",
			schema: `{
				"$defs": {"node": {
					"type": "object",
					"properties": {
						"enabled": {"default": true},
						"child": {"default": {}, "allOf": [{"$ref": "#/$defs/node"}]}
					}
				}},
				"$ref": "#/$defs/node"
			}`,
			target:   `{"child":{"child":{"label":"authored"}}}`,
			expected: `{"enabled":true,"child":{"enabled":true,"child":{"label":"authored","enabled":true,"child":{"enabled":true}}}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var expected interface{}
			if err := json.Unmarshal([]byte(tc.expected), &expected); err != nil {
				t.Fatal(err)
			}

			actual, err := processJSONSchemaParse(t.Context(), tc.schema, tc.target)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("got %#v, want %#v", actual, expected)
			}

			valid, err := processJSONSchemaValidate(t.Context(), tc.schema, tc.target)
			if err != nil || !valid {
				t.Fatalf("validate = %t, %v; want true, nil", valid, err)
			}
		})
	}
}

func TestReadURLSourceDoesNotLogCredentials(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "32")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	sourceURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	sourceURL.User = url.UserPassword("review-user", "review-password")
	sourceURL.RawQuery = "token=review-secret"

	for _, role := range []jsonSchemaSourceRole{jsonSchemaSourceRoleSchema, jsonSchemaSourceRoleTarget} {
		t.Run(string(role), func(t *testing.T) {
			var logs bytes.Buffer
			ctx := tflogtest.RootLogger(t.Context(), &logs)
			if _, err := readURLSource(ctx, sourceURL.String(), role); err == nil {
				t.Fatal("expected a truncated response error")
			}

			output := logs.String()
			if !strings.Contains(output, "response_read_failed") || !strings.Contains(output, sourceURL.Host) {
				t.Fatalf("expected host and failure metadata in logs, got %s", output)
			}
			for _, secret := range []string{"review-user", "review-password", "review-secret"} {
				if strings.Contains(output, secret) {
					t.Errorf("URL credentials leaked into logs: %s", output)
				}
			}
		})
	}
}

package provider

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestJsonschemaYamlParseFunction(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a sample JSON schema
	schemaContent := `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "default": "defaultName"
    },
    "version": {
      "type": "string"
    },
    "enabled": {
      "type": "boolean",
      "default": true
    },
    "config": {
      "type": "object",
      "properties": {
        "timeout": {
          "type": "integer",
          "default": 30
        },
        "retries": {
          "type": "integer"
        }
      },
      "default": {}
    }
  },
  "required": ["version"]
}`

	// Create a valid YAML file
	yamlContent := `version: "1.0.0"
enabled: false
config:
  retries: 3`

	// Create an invalid YAML file (missing required field)
	invalidYamlContent := `
		name: "test"
		enabled: true`

	// Create test files
	schemaFile := filepath.Join(tmpDir, "schema.json")
	validYamlFile := filepath.Join(tmpDir, "valid.yaml")
	invalidYamlFile := filepath.Join(tmpDir, "invalid.yaml")

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to create schema file: %v", err)
	}
	if err := os.WriteFile(validYamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create valid YAML file: %v", err)
	}
	if err := os.WriteFile(invalidYamlFile, []byte(invalidYamlContent), 0644); err != nil {
		t.Fatalf("Failed to create invalid YAML file: %v", err)
	}

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test valid YAML file with schema validation and defaults
				Config: `
					locals {
						schema_file = "` + schemaFile + `"
						yaml_file   = "` + validYamlFile + `"
					}
					
					output "parsed_yaml" {
						value = provider::helpers::jsonschema_yaml_parse(local.schema_file, local.yaml_file)
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("parsed_yaml", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"name":    knownvalue.StringExact("defaultName"), // Should apply default
						"version": knownvalue.StringExact("1.0.0"),
						"enabled": knownvalue.Bool(false),
						"config": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"timeout": knownvalue.Int64Exact(30), // Should apply default
							"retries": knownvalue.Int64Exact(3),
						}),
					})),
				},
			},
			{
				// Test error case - invalid YAML (missing required field)
				Config: `
					locals {
						schema_file = "` + schemaFile + `"
						yaml_file   = "` + invalidYamlFile + `"
					}
					
					output "parsed_yaml" {
						value = provider::helpers::jsonschema_yaml_parse(local.schema_file, local.yaml_file)
					}
				`,
				ExpectError: regexp.MustCompile(".*evaluation failed.*"),
			},
			{
				// Test error case - non-existent schema file
				Config: `
					output "parsed_yaml" {
						value = provider::helpers::jsonschema_yaml_parse("/non/existent/schema.json", "` + validYamlFile + `")
					}
				`,
				ExpectError: regexp.MustCompile(".*no such file or directory.*"),
			},
			{
				// Test error case - non-existent YAML file
				Config: `
					output "parsed_yaml" {
						value = provider::helpers::jsonschema_yaml_parse("` + schemaFile + `", "/non/existent/file.yaml")
					}
				`,
				ExpectError: regexp.MustCompile(".*no such file or directory.*"),
			},
		},
	})
}

func TestJsonschemaYamlParseFunctionWithEmptyYaml(t *testing.T) {
	t.Parallel()

	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Create a schema with defaults
	schemaContent := `{
	  "$schema": "https://json-schema.org/draft/2020-12/schema",
	  "type": "object",
	  "properties": {
		"name": {
		  "type": "string",
		  "default": "defaultName"
		},
		"enabled": {
		  "type": "boolean",
		  "default": true
		}
	  },
	  "default": {
		"name": "rootDefault",
		"enabled": false
	  }
	}`

	// Create an empty YAML file
	emptyYamlContent := `{}`

	// Create test files
	schemaFile := filepath.Join(tmpDir, "schema.json")
	emptyYamlFile := filepath.Join(tmpDir, "empty.yaml")

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to create schema file: %v", err)
	}
	if err := os.WriteFile(emptyYamlFile, []byte(emptyYamlContent), 0644); err != nil {
		t.Fatalf("Failed to create empty YAML file: %v", err)
	}

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test with empty YAML file - should apply defaults
				Config: `
					locals {
						schema_file = "` + schemaFile + `"
						yaml_file   = "` + emptyYamlFile + `"
					}
					
					output "parsed_yaml" {
						value = provider::helpers::jsonschema_yaml_parse(local.schema_file, local.yaml_file)
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("parsed_yaml", knownvalue.ObjectExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestJsonschemaYamlParseFunctionWithComplexTypes(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a schema with array and nested objects
	schemaContent := `{
	  "$schema": "https://json-schema.org/draft/2020-12/schema",
	  "type": "object",
	  "properties": {
		"servers": {
		  "type": "array",
		  "items": {
			"type": "object",
			"properties": {
			  "name": {"type": "string"},
			  "port": {"type": "integer"}
			}
		  }
		},
		"metadata": {
		  "type": "object",
		  "properties": {
			"created": {"type": "string"},
			"tags": {
			  "type": "array",
			  "items": {"type": "string"}
			}
		  }
		}
	  }
	}`

	// Create a YAML file with complex types
	yamlContent := `servers:
  - name: "web-server"
    port: 8080
  - name: "api-server"
    port: 3000
metadata:
  created: "2024-01-01"
  tags:
    - "production"
    - "web"
    - "api"`

	// Create test files
	schemaFile := filepath.Join(tmpDir, "schema.json")
	yamlFile := filepath.Join(tmpDir, "complex.yaml")

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to create schema file: %v", err)
	}
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test complex types (arrays and nested objects)
				Config: `
					locals {
						schema_file = "` + schemaFile + `"
						yaml_file   = "` + yamlFile + `"
					}
					
					output "parsed_yaml" {
						value = provider::helpers::jsonschema_yaml_parse(local.schema_file, local.yaml_file)
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("parsed_yaml", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"servers": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("web-server"),
								"port": knownvalue.Int64Exact(8080),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"name": knownvalue.StringExact("api-server"),
								"port": knownvalue.Int64Exact(3000),
							}),
						}),
						"metadata": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"created": knownvalue.StringExact("2024-01-01"),
							"tags": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.StringExact("production"),
								knownvalue.StringExact("web"),
								knownvalue.StringExact("api"),
							}),
						}),
					})),
				},
			},
		},
	})
}

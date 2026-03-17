package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestJsonschemaParseFunctionInlineSchemaInlineTarget(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
locals {
  schema = <<-SCHEMA
type: object
properties:
  app:
    type: object
    default: {}
    properties:
      name:
        type: string
        default: default-app
      enabled:
        type: boolean
        default: true
      timeout:
        type: integer
        default: 30
SCHEMA

  target = <<-TARGET
app:
  enabled: false
TARGET
}

output "parsed" {
  value = provider::helpers::jsonschema_parse(local.schema, local.target)
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("parsed", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"app": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name":    knownvalue.StringExact("default-app"),
							"enabled": knownvalue.Bool(false),
							"timeout": knownvalue.Int64Exact(30),
						}),
					})),
				},
			},
		},
	})
}

func TestJsonschemaParseFunctionFilePathSchemaAndTarget(t *testing.T) {
	testDirectory := t.TempDir()
	t.Setenv("PWD", testDirectory)

	writeTestFile(t, filepath.Join(testDirectory, "schema.yaml"), `
type: object
properties:
  service:
    type: object
    default: {}
    properties:
      name:
        type: string
      port:
        type: integer
        default: 8080
required:
  - service
`)

	writeTestFile(t, filepath.Join(testDirectory, "target.yaml"), `
service:
  name: api
`)

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
output "parsed" {
  value = provider::helpers::jsonschema_parse("schema.yaml", "target.yaml")
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("parsed", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"service": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"name": knownvalue.StringExact("api"),
							"port": knownvalue.Int64Exact(8080),
						}),
					})),
				},
			},
		},
	})
}

func TestJsonschemaParseFunctionURLSchemaAndTarget(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/schema":
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write([]byte(`{
  "type": "object",
  "properties": {
    "metadata": {
      "type": "object",
      "default": {},
      "properties": {
        "env": {"type": "string", "default": "dev"},
        "owner": {"type": "string", "default": "platform"}
      }
    }
  }
}`))
		case "/target":
			responseWriter.Header().Set("Content-Type", "application/yaml")
			_, _ = responseWriter.Write([]byte(`metadata:
  env: prod
`))
		default:
			http.NotFound(responseWriter, request)
		}
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
output "parsed" {
  value = provider::helpers::jsonschema_parse("%s/schema", "%s/target")
}
`, server.URL, server.URL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("parsed", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"metadata": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"env":   knownvalue.StringExact("prod"),
							"owner": knownvalue.StringExact("platform"),
						}),
					})),
				},
			},
		},
	})
}

func TestJsonschemaParseFunctionDefaultsNestedFromEmptyTarget(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
locals {
  schema = <<-SCHEMA
type: object
properties:
  app:
    type: object
    default: {}
    properties:
      logging:
        type: object
        default: {}
        properties:
          level:
            type: string
            default: info
          json:
            type: boolean
            default: true
SCHEMA

  target = "{}"
}

output "parsed" {
  value = provider::helpers::jsonschema_parse(local.schema, local.target)
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("parsed", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"app": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"logging": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"level": knownvalue.StringExact("info"),
								"json":  knownvalue.Bool(true),
							}),
						}),
					})),
				},
			},
		},
	})
}

func TestJsonschemaParseFunctionValidationFailure(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
locals {
  schema = jsonencode({
    type = "object"
    properties = {
      version = { type = "string" }
    }
    required = ["version"]
  })

  target = jsonencode({
    name = "missing-version"
  })
}

output "parsed" {
  value = provider::helpers::jsonschema_parse(local.schema, local.target)
}
`,
				ExpectError: regexp.MustCompile(`schema\s+validation failed`),
			},
		},
	})
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()

	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

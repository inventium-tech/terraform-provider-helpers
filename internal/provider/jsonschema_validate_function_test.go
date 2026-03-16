package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestJsonschemaValidateFunctionInlineSchemaInlineTargetValid(t *testing.T) {
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
      name = { type = "string" }
    }
    required = ["name"]
  })

  target = jsonencode({
    name = "example"
  })
}

output "is_valid" {
  value = provider::helpers::jsonschema_validate(local.schema, local.target)
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("is_valid", knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestJsonschemaValidateFunctionValidationFailureReturnsFalse(t *testing.T) {
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

output "is_valid" {
  value = provider::helpers::jsonschema_validate(local.schema, local.target)
}
`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("is_valid", knownvalue.Bool(false)),
				},
			},
		},
	})
}

func TestJsonschemaValidateFunctionURLSchemaAndTarget(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/schema":
			responseWriter.Header().Set("Content-Type", "application/json")
			_, _ = responseWriter.Write([]byte(`{
  "type": "object",
  "properties": {
    "enabled": {"type": "boolean"}
  },
  "required": ["enabled"]
}`))
		case "/target":
			responseWriter.Header().Set("Content-Type", "application/yaml")
			_, _ = responseWriter.Write([]byte(`enabled: true
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
output "is_valid" {
  value = provider::helpers::jsonschema_validate("%s/schema", "%s/target")
}
`, server.URL, server.URL),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("is_valid", knownvalue.Bool(true)),
				},
			},
		},
	})
}

func TestJsonschemaValidateFunctionOperationalFailureReturnsError(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
output "is_valid" {
  value = provider::helpers::jsonschema_validate("./this/path/does/not/exist/schema.yaml", "{}")
}
`,
				ExpectError: regexp.MustCompile(`error\s+reading schema source`),
			},
		},
	})
}

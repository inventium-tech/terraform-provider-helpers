package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"testing"
)

func TestObjectFilterKeysFunction(t *testing.T) {
	t.Parallel()

	mockLocalsObjects := `locals {
	  # Sample object with mixed data types
	  sample_object = {
		name        = "example"
		age         = 25
		active      = true
		tags        = ["tag1", "tag2"]
		metadata    = { created = "2024-01-01", version = "1.0" }
		description = null
		count       = 0
	  }
	
	  # Sample map
	  sample_map = {
		"env"     = "production"
		"region"  = "eu-central-1"
		"project" = "my-project"
		"owner"   = "team-alpha"
	  }
	}`

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test filtering object with matching keys
				Config: mockLocalsObjects + `
					locals {
						filter_keys = toset(["name", "age", "active", "metadata"])
					}
					
					output "filtered_object" { 
						value = provider::helpers::object_filter_keys(local.sample_object, local.filter_keys) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("filtered_object", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"name":   knownvalue.StringExact("example"),
						"age":    knownvalue.Int32Exact(25),
						"active": knownvalue.Bool(true),
						"metadata": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"created": knownvalue.StringExact("2024-01-01"),
							"version": knownvalue.StringExact("1.0"),
						}),
					})),
				},
			},
			{
				// Test filtering map with matching keys
				Config: mockLocalsObjects + `
					locals {
						filter_keys = toset(["env", "project"])
					}
					
					output "filtered_object" { 
						value = provider::helpers::object_filter_keys(local.sample_map, local.filter_keys) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("filtered_object", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"env":     knownvalue.StringExact("production"),
						"project": knownvalue.StringExact("my-project"),
					})),
				},
			},
			{
				// Test with an empty filter set
				Config: mockLocalsObjects + `
					locals {
						filter_keys = toset([])
					}
					
					output "filtered_object" { 
						value = provider::helpers::object_filter_keys(local.sample_object, local.filter_keys) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("filtered_object", knownvalue.ObjectExact(map[string]knownvalue.Check{})),
				},
			},
			{
				// Test with non-matching keys
				Config: mockLocalsObjects + `
					locals {
						filter_keys = toset(["key3", "key4"])
					}
					
					output "filtered_object" { 
						value = provider::helpers::object_filter_keys(local.sample_object, local.filter_keys) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("filtered_object", knownvalue.ObjectExact(map[string]knownvalue.Check{})),
				},
			},
			{
				// Test with all keys matching
				Config: mockLocalsObjects + `
					locals {
						filter_keys = toset(["env", "region", "project", "owner"])
					}
					
					output "filtered_object" { 
						value = provider::helpers::object_filter_keys(local.sample_map, local.filter_keys) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("filtered_object", knownvalue.ObjectExact(map[string]knownvalue.Check{
						"env":     knownvalue.StringExact("production"),
						"region":  knownvalue.StringExact("eu-central-1"),
						"project": knownvalue.StringExact("my-project"),
						"owner":   knownvalue.StringExact("team-alpha"),
					})),
				},
			},
		},
	})
}

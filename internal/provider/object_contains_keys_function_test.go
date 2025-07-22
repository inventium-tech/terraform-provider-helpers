package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"testing"
)

func TestObjectContainsKeysFunction(t *testing.T) {
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

	  # Empty object
	  empty_object = {}
	}`

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test strict=true (default) with all keys present in the object
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset(["name", "age", "active"])
					}
					
					output "contains_keys_strict_true" { 
						value = provider::helpers::object_contains_keys(local.sample_object, local.check_keys) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_strict_true", knownvalue.Bool(true)),
				},
			},
			{
				// Test strict=true explicitly with all keys present in the object
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset(["name", "age", "active"])
					}
					
					output "contains_keys_strict_explicit" { 
						value = provider::helpers::object_contains_keys(local.sample_object, local.check_keys, true) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_strict_explicit", knownvalue.Bool(true)),
				},
			},
			{
				// Test strict=true with missing keys in the object
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset(["name", "age", "missing_key"])
					}
					
					output "contains_keys_strict_missing" { 
						value = provider::helpers::object_contains_keys(local.sample_object, local.check_keys, true) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_strict_missing", knownvalue.Bool(false)),
				},
			},
			{
				// Test strict=false with some keys present in the object
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset(["name", "missing_key1", "missing_key2"])
					}
					
					output "contains_keys_non_strict_partial" { 
						value = provider::helpers::object_contains_keys(local.sample_object, local.check_keys, false) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_non_strict_partial", knownvalue.Bool(true)),
				},
			},
			{
				// Test strict=false with no keys present in the object
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset(["missing_key1", "missing_key2", "missing_key3"])
					}
					
					output "contains_keys_non_strict_none" { 
						value = provider::helpers::object_contains_keys(local.sample_object, local.check_keys, false) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_non_strict_none", knownvalue.Bool(false)),
				},
			},
			{
				// Test strict=true with all keys present the map
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset(["env", "region"])
					}
					
					output "contains_keys_map_strict" { 
						value = provider::helpers::object_contains_keys(local.sample_map, local.check_keys, true) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_map_strict", knownvalue.Bool(true)),
				},
			},
			{
				// Test strict=false with some keys present the map
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset(["env", "missing_key"])
					}
					
					output "contains_keys_map_non_strict" { 
						value = provider::helpers::object_contains_keys(local.sample_map, local.check_keys, false) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_map_non_strict", knownvalue.Bool(true)),
				},
			},
			{
				// Test strict=true (default) with an empty set of keys
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset([])
					}
					
					output "contains_keys_empty_set" { 
						value = provider::helpers::object_contains_keys(local.sample_object, local.check_keys) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_empty_set", knownvalue.Bool(false)),
				},
			},
			{
				// Test strict=true with an empty object
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset(["any_key"])
					}
					
					output "contains_keys_empty_object_strict" { 
						value = provider::helpers::object_contains_keys(local.empty_object, local.check_keys, true) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_empty_object_strict", knownvalue.Bool(false)),
				},
			},
			{
				// Test strict=false with an empty object
				Config: mockLocalsObjects + `
					locals {
						check_keys = toset(["any_key"])
					}
					
					output "contains_keys_empty_object_non_strict" { 
						value = provider::helpers::object_contains_keys(local.empty_object, local.check_keys, false) 
					}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("contains_keys_empty_object_non_strict", knownvalue.Bool(false)),
				},
			},
		},
	})
}

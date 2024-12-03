package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"testing"
)

func TestCollectionFilterFunction(t *testing.T) {
	t.Parallel()

	mockLocals := `locals {
	  test_object_collection = [
        { key1 = "value1", key2 = true, key3 = 3, key4 = null },
        { key1 = "value2", key2 = false, key3 = 0, key4 = {} },
        { key1 = "value3", key2 = true, key3 = 5, key4 = null },
	    { key1 = "value4", key2 = false, key3 = 1, key4 = { key5 = "value5", key6 = true } },
      ]
	
	  test_string_array = ["value1", "value2", "value3", "value1"]
	
	  test_number_array = [5, 8, 3, 5]
	
	  test_bool_array = [true, false, true]
	}`

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// test filter array of objects
			{
				Config: mockLocals + `
			
				output "test_match_object_value_string" {
				  value = provider::helpers::collection_filter(local.test_object_collection, "key1", "value1")
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_object_value_string", knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownOutputValue("test_match_object_value_string", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key1": knownvalue.StringExact("value1"),
							"key2": knownvalue.Bool(true),
							"key3": knownvalue.Int64Exact(3),
							"key4": knownvalue.Null(),
						}),
					})),
				},
			},
			{
				Config: mockLocals + `
			
				output "test_match_object_bool_value" {
				  value = provider::helpers::collection_filter(local.test_object_collection, "key2", true)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_object_bool_value", knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownOutputValue("test_match_object_bool_value", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key1": knownvalue.StringExact("value1"),
							"key2": knownvalue.Bool(true),
							"key3": knownvalue.Int64Exact(3),
							"key4": knownvalue.Null(),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key1": knownvalue.StringExact("value3"),
							"key2": knownvalue.Bool(true),
							"key3": knownvalue.Int64Exact(5),
							"key4": knownvalue.Null(),
						}),
					})),
				},
			},
			{
				Config: mockLocals + `
			
				output "test_match_object_number_value" {
				  value = provider::helpers::collection_filter(local.test_object_collection, "key3", 5)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_object_number_value", knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownOutputValue("test_match_object_number_value", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key1": knownvalue.StringExact("value3"),
							"key2": knownvalue.Bool(true),
							"key3": knownvalue.Int64Exact(5),
							"key4": knownvalue.Null(),
						}),
					})),
				},
			},
			{
				Config: mockLocals + `
			
				output "test_match_object_null_value" {
				  value = provider::helpers::collection_filter(local.test_object_collection, "key4", null)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_null", knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownOutputValue("test_match_null", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key1": knownvalue.StringExact("value1"),
							"key2": knownvalue.Bool(true),
							"key3": knownvalue.Int64Exact(3),
							"key4": knownvalue.Null(),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key1": knownvalue.StringExact("value3"),
							"key2": knownvalue.Bool(true),
							"key3": knownvalue.Int64Exact(5),
							"key4": knownvalue.Null(),
						}),
					})),
				},
			},
			{
				Config: mockLocals + `
			
				output "test_match_object_string_nested_value" {
				  value = provider::helpers::collection_filter(local.test_object_collection, "key4.key5", "value5")
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_object_string_nested_value", knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownOutputValue("test_match_object_string_nested_value", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key1": knownvalue.StringExact("value4"),
							"key2": knownvalue.Bool(false),
							"key3": knownvalue.Int64Exact(1),
							"key4": knownvalue.ObjectExact(map[string]knownvalue.Check{
								"key5": knownvalue.StringExact("value5"),
								"key6": knownvalue.Bool(true),
							}),
						}),
					})),
				},
			},
			{
				Config: mockLocals + `
			
				output "test_no_match_object_string_value" {
				  value = provider::helpers::collection_filter(local.test_object_collection, "key1", "new_value")
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_no_match_object_string_value", knownvalue.ListSizeExact(0)),
				},
			},

			// test filter array of strings
			{
				Config: mockLocals + `
			
				output "test_match_string_array" {
				  value = provider::helpers::collection_filter(local.test_string_array, "", "value2")
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_string_array", knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownOutputValue("test_match_string_array", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("value2"),
					})),
				},
			},
			{
				Config: mockLocals + `
			
				output "test_match_string_array" {
				  value = provider::helpers::collection_filter(local.test_string_array, "", "value1")
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_string_array", knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownOutputValue("test_match_string_array", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("value1"),
						knownvalue.StringExact("value1"),
					})),
				},
			},

			// test filter array of numbers
			{
				Config: mockLocals + `
			
				output "test_match_number_array" {
				  value = provider::helpers::collection_filter(local.test_number_array, "", 3)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_number_array", knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownOutputValue("test_match_number_array", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.Int64Exact(3),
					})),
				},
			},
			{
				Config: mockLocals + `
			
				output "test_match_number_array" {
				  value = provider::helpers::collection_filter(local.test_number_array, "", 5)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_number_array", knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownOutputValue("test_match_number_array", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.Int64Exact(5),
						knownvalue.Int64Exact(5),
					})),
				},
			},

			// test filter bool array
			{
				Config: mockLocals + `
			
				output "test_match_bool_array" {
				  value = provider::helpers::collection_filter(local.test_bool_array, "", false)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_bool_array", knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownOutputValue("test_match_bool_array", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.Bool(false),
					})),
				},
			},
			{
				Config: mockLocals + `
			
				output "test_match_bool_array" {
				  value = provider::helpers::collection_filter(local.test_bool_array, "", true)
				}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownOutputValue("test_match_bool_array", knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownOutputValue("test_match_bool_array", knownvalue.ListExact([]knownvalue.Check{
						knownvalue.Bool(true),
						knownvalue.Bool(true),
					})),
				},
			},
		},
	})
}

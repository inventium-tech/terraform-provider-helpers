package provider

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"testing"
)

type mockTFObjectPartial map[string]knownvalue.Check

func (m mockTFObjectPartial) modify(t *testing.T, key string, value knownvalue.Check) mockTFObjectPartial {
	t.Helper()
	result := make(mockTFObjectPartial, len(m))
	for k, v := range m {
		result[k] = v
	}
	result[key] = value
	return result
}

func TestObjectSetValueFunction(t *testing.T) {
	t.Parallel()

	mockObject := mockTFObjectPartial{
		"key1": knownvalue.StringExact("value1"),
		"key2": knownvalue.Bool(true),
		"key3": knownvalue.Int32Exact(3),
		"key4": knownvalue.StringExact(""),
		"key5": knownvalue.Null(),
	}
	mockTerraformLocalsTestObject := `locals {
		test_object = { key1 = "value1", key2 = true, key3 = 3, key4 = "", key5 = null }
	`

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// test "write_all" operation mode
				Config: `` +
					mockTerraformLocalsTestObject + `
			
					output "expect_value_change" { value = provider::helpers::object_set_value(local.test_object, "key1", "new_value", "write_all") }
			
					output "expect_new_key" { value = provider::helpers::object_set_value(local.test_object, "new_key", "new_value", "write_all") }
					`,
				ConfigStateChecks: []statecheck.StateCheck{
					// case: existing key => update value
					statecheck.ExpectKnownOutputValue("expect_value_change", knownvalue.ObjectExact(mockObject.modify(t, "key1", knownvalue.StringExact("new_value")))),
					// case: missing key => add key & value
					statecheck.ExpectKnownOutputValue("expect_new_key", knownvalue.ObjectExact(mockObject.modify(t, "new_key", knownvalue.StringExact("new_value")))),
				},
			},
			{
				// test "write_value" operation mode
				Config: `` +
					mockTerraformLocalsTestObject + `
			
					output "expect_value_change" { value = provider::helpers::object_set_value(local.test_object, "key1", "new_value", "write_value") }
			
					output "expect_no_changes" { value = provider::helpers::object_set_value(local.test_object, "new_key", "new_value", "write_value") }
					`,
				ConfigStateChecks: []statecheck.StateCheck{
					// case: existing key => update value
					statecheck.ExpectKnownOutputValue("expect_value_change", knownvalue.ObjectExact(mockObject.modify(t, "key1", knownvalue.StringExact("new_value")))),
					// case: missing key => no changes
					statecheck.ExpectKnownOutputValue("expect_no_changes", knownvalue.ObjectExact(mockObject)),
				},
			},
			{
				// test "write_safe" operation mode
				Config: `` +
					mockTerraformLocalsTestObject + `

					output "expect_value_change_1" { value = provider::helpers::object_set_value(local.test_object, "key4", "new_value", "write_safe") }

					output "expect_value_change_2" { value = provider::helpers::object_set_value(local.test_object, "key5", "new_value", "write_safe") }
					
					output "expect_no_changes_1" { value = provider::helpers::object_set_value(local.test_object, "key1", "new_value", "write_safe") }

					output "expect_no_changes_2" { value = provider::helpers::object_set_value(local.test_object, "new_key", "new_value", "write_safe") }
					`,
				ConfigStateChecks: []statecheck.StateCheck{
					// case: existing key, empty string value => update value
					statecheck.ExpectKnownOutputValue("expect_value_change_1", knownvalue.ObjectExact(mockObject.modify(t, "key4", knownvalue.StringExact("new_value")))),
					// case: existing key, null value => update value
					statecheck.ExpectKnownOutputValue("expect_value_change_2", knownvalue.ObjectExact(mockObject.modify(t, "key5", knownvalue.StringExact("new_value")))),
					// case: existing key, non-empty value => no changes
					statecheck.ExpectKnownOutputValue("expect_no_changes_1", knownvalue.ObjectExact(mockObject)),
					// case: missing key => no changes
					statecheck.ExpectKnownOutputValue("expect_no_changes_2", knownvalue.ObjectExact(mockObject)),
				},
			},
		},
	})
}

locals {
  test_object = {
    key1 = "value1"
    key2 = true
    key3 = 3
    key4 = ""
    key5 = null
  }
}

## Expected output
# write_all_operation = {
#   test_value_change_on_existing_key = { key1 = "new_value", key2 = true, key3 = 3, key4 = "", key5 = null },
#   test_add_new_key_value            = { key1 = "value1", key2 = true, key3 = 3, key4 = "", key5 = null, new_key = "new_value" }
# }
output "write_all_operation" {
  value = {
    test_value_change_on_existing_key = provider::helpers::object_set_value(local.test_object, "key1", "new_value", "write_all")
    test_add_new_key_value            = provider::helpers::object_set_value(local.test_object, "new_key", "new_value", "write_all")
  }
}

## Expected output
# write_value_operation = {
#   test_value_change_on_existing_key = { key1 = "new_value", key2 = true, key3 = 3, key4 = "", key5 = null },
#   test_no_changes_on_missing_key    = { key1 = "value1", key2 = true, key3 = 3, key4 = "", key5 = null }
# }
output "write_value_operation" {
  value = {
    test_value_change_on_existing_key = provider::helpers::object_set_value(local.test_object, "key1", "new_value", "write_value")
    test_no_changes_on_missing_key    = provider::helpers::object_set_value(local.test_object, "new_key", "new_value", "write_value")
  }
}

## Expected output
# write_safe_operation = {
#   test_value_change_on_empty_string     = { key1 = "value1", key2 = true, key3 = 3, key4 = "new_value", key5 = null },
#   test_value_change_on_null_value       = { key1 = "value1", key2 = true, key3 = 3, key4 = "", key5 = "new_value" },
#   test_no_changes_on_existing_key_value = { key1 = "value1", key2 = true, key3 = 3, key4 = "", key5 = null },
#   test_no_changes_on_missing_key        = { key1 = "value1", key2 = true, key3 = 3, key4 = "", key5 = null }
# }
output "write_safe_operation" {
  value = {
    test_value_change_on_empty_string     = provider::helpers::object_set_value(local.test_object, "key4", "new_value", "write_safe")
    test_value_change_on_null_value       = provider::helpers::object_set_value(local.test_object, "key5", "new_value", "write_safe")
    test_no_changes_on_existing_key_value = provider::helpers::object_set_value(local.test_object, "key1", "new_value", "write_safe")
    test_no_changes_on_missing_key        = provider::helpers::object_set_value(local.test_object, "new_key", "new_value", "write_safe")
  }
}

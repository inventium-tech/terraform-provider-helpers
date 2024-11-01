---
page_title: "object_set_value function - helpers"
subcategory: "Object Functions"
description: |-
    Sets a value in an Object or creates a new key with the value
---

# Function: object_set_value

Sets a value in an Object or creates a new key with the value

The function `object_set_value` have different modes of operation, depending on the value of the `operation` argument. 
Check the [Operation Modes](#operation-modes) section for more information.

## Example Usage

```terraform
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
```

## Signature

<!-- signature generated by tfplugindocs -->
```text
object_set_value(object dynamic, key string, value dynamic, operation string) dynamic
```

## Arguments

<!-- arguments generated by tfplugindocs -->
1. `object` (Dynamic) The Object to set the value in
1. `key` (String) The key to set the value in
1. `value` (Dynamic, Nullable) The value to set in the key
1. `operation` (String) The operation mode to use when setting the value


### Operation Mode

The `operation` argument can have the following values:

- `write_all`: This mode will write the value to the key, if the key does not exist it will be created.
- `write_value`: This mode will write the value to the key ONLY if the key exists, otherwise expect no changes.
- `write_safe`: This mode will write the value to the key ONLY if the key exists and the value is `null` or empty 
  string.

## Return Type

The return type of `object_set_value` is an Object from the input argument `object` with the desired changes.
---
page_title: "collection_filter function - helpers"
subcategory: "Collection Functions"
description: |-
    Filter collection of objects.
---

# Function: collection_filter

Filter collection of objects.

The function `collection_filter` can be used, as indicated by the name, to quickly filter a collection of values.
Terraform does not offer out-of-the-box such functionality in a direct way, with the only reasonable solution of loop
through the collection and perform the filtering.

In the current version the function is able to filter collection of primitives (number, bool, string) and objects, with
the last one able to also filter by a nested attribute. The filter right now is only using an "equal" check operation,
however, we might put some effort in the future to support different operators.


## Example Usage

```terraform
locals {
  test_object_collection = [
    { key1 = "value1", key2 = true, key3 = 3, key4 = null },
    { key1 = "value2", key2 = false, key3 = 0, key4 = {} },
    { key1 = "value3", key2 = true, key3 = 5, key4 = null },
    { key1 = "value4", key2 = false, key3 = 1, key4 = { key5 = "value5", key6 = true } },
  ]

  test_string_array = ["value1", "value2", "value3", "value1"]

  test_number_array = [5, 8, 3, 5]

  test_bool_array = [true, false, true]
}

# Expected return:
# [
#   { key1 = "value1", key2 = true, key3 = 3, key4 = null },
# ]
output "test_match_object_string_value" {
  value = provider::helpers::collection_filter(local.test_object_collection, "key1", "value1")
}

# Expected return:
# [
#   { key1 = "value1", key2 = true, key3 = 3, key4 = null },
#   { key1 = "value3", key2 = true, key3 = 5, key4 = null },
# ]
output "test_match_object_bool_value" {
  value = provider::helpers::collection_filter(local.test_object_collection, "key2", true)
}

# Expected return:
# [
#   { key1 = "value3", key2 = true, key3 = 5, key4 = null },
# ]
output "test_match_object_number_value" {
  value = provider::helpers::collection_filter(local.test_object_collection, "key3", 5)
}

# Expected return:
# [
#   { key1 = "value1", key2 = true, key3 = 3, key4 = null },
#   { key1 = "value3", key2 = true, key3 = 5, key4 = null },
# ]
output "test_match_object_null_value" {
  value = provider::helpers::collection_filter(local.test_object_collection, "key4", null)
}

# Expected return:
# [
#   { key1 = "value4", key2 = false, key3 = 1, key4 = { key5 = "value5", key6 = true } },
# ]
output "test_match_object_string_nested_value" {
  value = provider::helpers::collection_filter(local.test_object_collection, "key4.key5", "value5")
}

# Expected return:
# [ ]
output "test_no_match_object_string_value" {
  value = provider::helpers::collection_filter(local.test_object_collection, "key1", "new_value")
}

# Expected return:
# ["value1", "value1"]
output "test_match_string_array" {
  value = provider::helpers::collection_filter(local.test_string_array, "", "value1")
}

# Expected return:
# [5, 5]
output "test_match_number_array" {
  value = provider::helpers::collection_filter(local.test_number_array, "", 5)
}

# Expected return:
# [false]
output "test_match_bool_array" {
  value = provider::helpers::collection_filter(local.test_bool_array, "", false)
}
```

## Signature

<!-- signature generated by tfplugindocs -->
```text
collection_filter(collection dynamic, key string, value dynamic) dynamic
```

## Arguments

<!-- arguments generated by tfplugindocs -->
1. `collection` (Dynamic) The collection of objects to filter
1. `key` (String) The key from the object to filter by
1. `value` (Dynamic, Nullable) The value used to compare against


## Return Type

The signature shows a dynamic type of return because in order to support multiple types of collections the return must
be specified in such way. You can always expect a collection of the same type as used in the input.

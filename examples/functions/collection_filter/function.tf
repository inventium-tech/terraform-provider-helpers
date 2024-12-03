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

locals {
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
    "region"  = "us-west-2"
    "project" = "my-project"
    "owner"   = "team-alpha"
  }

  # Different filter sets
  basic_keys       = toset(["name", "age", "active"])
  metadata_keys    = toset(["name", "metadata", "description"])
  map_keys         = toset(["env", "project"])
  empty_keys       = toset([])
  nonexistent_keys = toset(["foo", "bar", "baz"])
}

## Expected output
# basic_filtering = {
#   name   = "example"
#   age    = 25
#   active = true
# }
output "basic_filtering" {
  description = "Filter object to keep only basic information"
  value       = provider::helpers::object_filter_keys(local.sample_object, local.basic_keys)
}

## Expected output
# metadata_filtering = {
#   name        = "example"
#   metadata    = { created = "2024-01-01", version = "1.0" }
#   description = null
# }
output "metadata_filtering" {
  description = "Filter object to keep metadata and null values"
  value       = provider::helpers::object_filter_keys(local.sample_object, local.metadata_keys)
}

## Expected output
# map_filtering = {
#   env     = "production"
#   project = "my-project"
# }
output "map_filtering" {
  description = "Filter map to keep only specific keys"
  value       = provider::helpers::object_filter_keys(local.sample_map, local.map_keys)
}

## Expected output
# empty_filtering = {}
output "empty_filtering" {
  description = "Filter with empty key set returns empty object"
  value       = provider::helpers::object_filter_keys(local.sample_object, local.empty_keys)
}

## Expected output
# nonexistent_filtering = {}
output "nonexistent_filtering" {
  description = "Filter with non-existent keys returns empty object"
  value       = provider::helpers::object_filter_keys(local.sample_object, local.nonexistent_keys)
}

## Expected output
# all_keys_filtering = {
#   name        = "example"
#   age         = 25
#   active      = true
#   tags        = ["tag1", "tag2"]
#   metadata    = { created = "2024-01-01", version = "1.0" }
#   description = null
#   count       = 0
# }
output "all_keys_filtering" {
  description = "Filter with all existing keys returns identical object"
  value = provider::helpers::object_filter_keys(
    local.sample_object,
    toset(keys(local.sample_object))
  )
}

## Expected output
# complex_filtering = {
#   tags  = ["tag1", "tag2"]
#   count = 0
# }
output "complex_filtering" {
  description = "Filter to keep complex data types and zero values"
  value = provider::helpers::object_filter_keys(
    local.sample_object,
    toset(["tags", "count", "missing_key"])
  )
}

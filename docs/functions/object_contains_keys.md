---
page_title: "object_contains_keys function - helpers"
subcategory: "Object Functions"
description: |-
    Check if an object contains specified keys.
---

# Function: object_contains_keys

Check if an object contains specified keys.

The function `object_contains_keys` checks whether an object or map contains specified keys. 
It supports two modes: strict mode (default) where all specified keys must be present, and non-strict mode where at least one key must be present.

## Example Usage

```terraform
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

  # Empty object for testing
  empty_object = {}

  # Different key sets for testing
  existing_keys   = toset(["name", "age", "active"])
  partial_keys    = toset(["name", "missing_key1", "missing_key2"])
  missing_keys    = toset(["foo", "bar", "baz"])
  single_existing = toset(["name"])
  single_missing  = toset(["missing_key"])
  empty_keys      = toset([])
  all_object_keys = toset(["name", "age", "active", "tags", "metadata", "description", "count"])
  map_keys        = toset(["env", "region"])
}

## STRICT MODE EXAMPLES (default behavior)

## Expected output: true
output "strict_all_present" {
  description = "Strict mode: all specified keys are present in object"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.existing_keys)
}

## Expected output: true
output "strict_explicit_true" {
  description = "Strict mode explicitly set to true: all keys present"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.existing_keys, true)
}

## Expected output: false
output "strict_some_missing" {
  description = "Strict mode: some keys are missing from object"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.partial_keys, true)
}

## Expected output: false
output "strict_all_missing" {
  description = "Strict mode: all specified keys are missing from object"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.missing_keys, true)
}

## Expected output: true
output "strict_single_present" {
  description = "Strict mode: single key is present"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.single_existing, true)
}

## Expected output: false
output "strict_single_missing" {
  description = "Strict mode: single key is missing"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.single_missing, true)
}

## NON-STRICT MODE EXAMPLES

## Expected output: true
output "non_strict_all_present" {
  description = "Non-strict mode: all specified keys are present"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.existing_keys, false)
}

## Expected output: true
output "non_strict_partial_present" {
  description = "Non-strict mode: at least one key is present (name exists)"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.partial_keys, false)
}

## Expected output: false
output "non_strict_none_present" {
  description = "Non-strict mode: none of the specified keys are present"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.missing_keys, false)
}

## Expected output: true
output "non_strict_single_present" {
  description = "Non-strict mode: single key is present"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.single_existing, false)
}

## Expected output: false
output "non_strict_single_missing" {
  description = "Non-strict mode: single key is missing"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.single_missing, false)
}

## MAP EXAMPLES

## Expected output: true
output "map_strict_present" {
  description = "Map with strict mode: all specified keys are present"
  value       = provider::helpers::object_contains_keys(local.sample_map, local.map_keys, true)
}

## Expected output: true
output "map_non_strict_partial" {
  description = "Map with non-strict mode: at least one key is present"
  value       = provider::helpers::object_contains_keys(local.sample_map, toset(["env", "missing_key"]), false)
}

## EDGE CASES

## Expected output: false
output "empty_key_set" {
  description = "Empty key set always returns false"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.empty_keys)
}

## Expected output: false
output "empty_object_strict" {
  description = "Empty object with strict mode returns false"
  value       = provider::helpers::object_contains_keys(local.empty_object, local.single_existing, true)
}

## Expected output: false
output "empty_object_non_strict" {
  description = "Empty object with non-strict mode returns false"
  value       = provider::helpers::object_contains_keys(local.empty_object, local.single_existing, false)
}

## Expected output: true
output "all_keys_present" {
  description = "All object keys are present (comprehensive check)"
  value       = provider::helpers::object_contains_keys(local.sample_object, local.all_object_keys, true)
}

## PRACTICAL USE CASES

## Expected output: true
output "required_fields_check" {
  description = "Check if object has all required fields for validation"
  value = provider::helpers::object_contains_keys(
    local.sample_object,
    toset(["name", "active"]), # Required fields
    true                       # All must be present
  )
}

## Expected output: true
output "optional_fields_check" {
  description = "Check if object has any of the optional identification fields"
  value = provider::helpers::object_contains_keys(
    local.sample_object,
    toset(["id", "name", "uuid"]), # Any identification field
    false                          # At least one must be present
  )
}

## Expected output: false
output "missing_critical_fields" {
  description = "Check for critical fields that are missing"
  value = provider::helpers::object_contains_keys(
    local.sample_object,
    toset(["password", "secret_key", "api_token"]), # Critical security fields
    false                                           # Any would be concerning
  )
}
```

## Signature

<!-- signature generated by tfplugindocs -->
```text
object_contains_keys(object dynamic, keys set of string, strict bool...) bool
```

## Arguments

<!-- arguments generated by tfplugindocs -->
1. `object` (Dynamic) The object or map to check for keys
1. `keys` (Set of String) Set of keys to check for in the object
<!-- variadic argument generated by tfplugindocs -->
1. `strict` (Variadic, Boolean) When true (default), all keys must be present. When false, at least one key must be present

## Return Type

The return type of `object_contains_keys` is a boolean:
- `true` if the key presence condition is met based on the `strict` parameter
- `false` if the key presence condition is not met

## Behavior

### Strict Mode (default: `strict = true`)
- **All** specified keys must be present in the object for the function to return `true`
- If any key is missing, the function returns `false`
- This is the default behavior when the `strict` parameter is omitted

### Non-Strict Mode (`strict = false`)
- **At least one** of the specified keys must be present in the object for the function to return `true`
- Only returns `false` if none of the specified keys are found

### General Behavior
- The function works with both Terraform objects and maps
- Key matching is case-sensitive and exact
- If the `keys` set is empty, the function returns `false`
- If the input object is empty, the function returns `false` (regardless of strict mode)
- The original object is not modified; this is a read-only operation

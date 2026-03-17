---
page_title: "jsonschema_parse function - helpers"
subcategory: "Configuration Functions"
description: |-
    Parse and validate data against JSON Schema with defaults.
---

# Function: jsonschema_parse

Parse and validate data against JSON Schema with defaults.

The function `jsonschema_parse` resolves both schema and target from **URL**, **file path** (including relative paths), or **inline JSON/YAML content**. It validates the target against the schema and returns a structured object with schema defaults applied recursively.

Key features:
- **Flexible Inputs**: Schema and target can each be URL, path, or inline content
- **Schema Validation**: Ensures input content conforms to the specified JSON Schema
- **Default Application**: Automatically applies default values defined in the schema
- **Type Safety**: Validates data types according to the schema definition
- **Error Handling**: Provides clear error messages for validation failures or file access issues

## Example Usage

```terraform
# Example usage of the jsonschema_parse function
# This function parses YAML files against JSON Schema definitions and applies defaults

# Create sample schema and YAML files for demonstration
resource "local_file" "sample_schema" {
  filename = "${path.module}/sample_schema.json"
  content = jsonencode({
    "$schema" = "https://json-schema.org/draft/2020-12/schema"
    type      = "object"
    properties = {
      name = {
        type    = "string"
        default = "default-app"
      }
      version = {
        type = "string"
      }
      enabled = {
        type    = "boolean"
        default = true
      }
      config = {
        type = "object"
        properties = {
          timeout = {
            type    = "integer"
            default = 30
          }
          retries = {
            type = "integer"
          }
          database = {
            type = "object"
            properties = {
              host = {
                type    = "string"
                default = "localhost"
              }
              port = {
                type    = "integer"
                default = 5432
              }
            }
          }
        }
        default = {}
      }
      tags = {
        type = "array"
        items = {
          type = "string"
        }
      }
    }
    required = ["version"]
  })
}

resource "local_file" "complete_config" {
  filename = "${path.module}/complete_config.yaml"
  content  = <<-EOT
    version: "1.2.0"
    enabled: false
    config:
      retries: 5
      database:
        host: "prod-db.example.com"
        port: 3306
    tags:
      - "production"
      - "web"
      - "critical"
  EOT
}

resource "local_file" "minimal_config" {
  filename = "${path.module}/minimal_config.yaml"
  content  = <<-EOT
    version: "0.1.0"
  EOT
}

resource "local_file" "partial_config" {
  filename = "${path.module}/partial_config.yaml"
  content  = <<-EOT
    version: "2.0.0"
    name: "my-custom-app"
    config:
      timeout: 60
      database: {}
    tags: []
  EOT
}

## BASIC USAGE EXAMPLES

# Parse a complete YAML configuration
output "complete_parsed_config" {
  description = "Parse YAML with all fields specified"
  value = provider::helpers::jsonschema_parse(
    local_file.sample_schema.filename,
    local_file.complete_config.filename
  )
}

# Parse a minimal YAML configuration (defaults will be applied)
output "minimal_parsed_config" {
  description = "Parse minimal YAML - defaults should be applied"
  value = provider::helpers::jsonschema_parse(
    local_file.sample_schema.filename,
    local_file.minimal_config.filename
  )
}

# Parse a partial YAML configuration
output "partial_parsed_config" {
  description = "Parse partial YAML configuration with mixed defaults"
  value = provider::helpers::jsonschema_parse(
    local_file.sample_schema.filename,
    local_file.partial_config.filename
  )
}

## PRACTICAL USE CASES

# Extract specific configuration values
locals {
  parsed_config = provider::helpers::jsonschema_parse(
    local_file.sample_schema.filename,
    local_file.complete_config.filename
  )
}

output "app_name" {
  description = "Application name from parsed config"
  value       = local.parsed_config.name
}

output "database_config" {
  description = "Database configuration with defaults applied"
  value       = local.parsed_config.config.database
}

output "is_enabled" {
  description = "Whether the application is enabled"
  value       = local.parsed_config.enabled
}

output "timeout_setting" {
  description = "Timeout configuration value"
  value       = local.parsed_config.config.timeout
}

# Use parsed configuration in resource creation
resource "local_file" "generated_config" {
  filename = "${path.module}/generated_app_config.json"
  content = jsonencode({
    app_name = local.parsed_config.name
    version  = local.parsed_config.version
    enabled  = local.parsed_config.enabled
    database = {
      connection_string = "postgresql://${local.parsed_config.config.database.host}:${local.parsed_config.config.database.port}/myapp"
      timeout           = local.parsed_config.config.timeout
      max_retries       = local.parsed_config.config.retries
    }
    deployment_tags = local.parsed_config.tags
  })
}

## ADVANCED SCHEMA EXAMPLES

# Create a more complex schema for microservice configuration
resource "local_file" "microservice_schema" {
  filename = "${path.module}/microservice_schema.json"
  content = jsonencode({
    "$schema" = "https://json-schema.org/draft/2020-12/schema"
    type      = "object"
    properties = {
      service = {
        type = "object"
        properties = {
          name = {
            type = "string"
          }
          port = {
            type    = "integer"
            default = 8080
          }
          health_check = {
            type = "object"
            properties = {
              endpoint = {
                type    = "string"
                default = "/health"
              }
              interval = {
                type    = "integer"
                default = 30
              }
            }
            default = {}
          }
        }
        required = ["name"]
      }
      logging = {
        type = "object"
        properties = {
          level = {
            type    = "string"
            default = "info"
          }
          format = {
            type    = "string"
            default = "json"
          }
        }
        default = {}
      }
      metrics = {
        type = "object"
        properties = {
          enabled = {
            type    = "boolean"
            default = true
          }
          port = {
            type    = "integer"
            default = 9090
          }
        }
        default = {}
      }
    }
    required = ["service"]
  })
}

resource "local_file" "microservice_config" {
  filename = "${path.module}/microservice_config.yaml"
  content  = <<-EOT
    service:
      name: "user-api"
      port: 3000
    logging:
      level: "debug"
  EOT
}

# Parse microservice configuration
output "microservice_parsed" {
  description = "Parsed microservice configuration with defaults"
  value = provider::helpers::jsonschema_parse(
    local_file.microservice_schema.filename,
    local_file.microservice_config.filename
  )
}

# Use in conditional logic
locals {
  microservice_config = provider::helpers::jsonschema_parse(
    local_file.microservice_schema.filename,
    local_file.microservice_config.filename
  )

  should_enable_metrics = local.microservice_config.metrics.enabled
  metrics_port          = local.microservice_config.metrics.port
}

output "conditional_metrics_config" {
  description = "Metrics configuration only if enabled"
  value = local.should_enable_metrics ? {
    enabled  = true
    port     = local.metrics_port
    endpoint = "http://localhost:${local.metrics_port}/metrics"
  } : null
}

## VALIDATION AND ERROR HANDLING

# The following would cause validation errors if uncommented:
# 
# resource "local_file" "invalid_config" {
#   filename = "${path.module}/invalid_config.yaml"
#   content = <<-EOT
#     # Missing required 'version' field
#     name: "invalid-app"
#     enabled: true
#   EOT
# }
# 
# output "invalid_config_parse" {
#   description = "This would fail validation due to missing required field"
#   value       = provider::helpers::jsonschema_parse(
#     local_file.sample_schema.filename,
#     local_file.invalid_config.filename
#   )
# }

## INTEGRATION WITH OTHER TERRAFORM FEATURES

# Use parsed configuration with for_each
locals {
  services_config = provider::helpers::jsonschema_parse(
    local_file.microservice_schema.filename,
    local_file.microservice_config.filename
  )
}

# Create multiple resources based on parsed configuration
resource "local_file" "service_port_config" {
  filename = "${path.module}/port_${local.services_config.service.port}.txt"
  content  = "Service ${local.services_config.service.name} running on port ${local.services_config.service.port}"
}

# Output summary of all configurations
output "configuration_summary" {
  description = "Summary of all parsed configurations"
  value = {
    complete_config = {
      name    = local.parsed_config.name
      version = local.parsed_config.version
      enabled = local.parsed_config.enabled
    }
    microservice_config = {
      name          = local.microservice_config.service.name
      port          = local.microservice_config.service.port
      logging_level = local.microservice_config.logging.level
    }
    default_values_applied = {
      database_host         = local.parsed_config.config.database.host
      health_check_endpoint = local.microservice_config.service.health_check.endpoint
      metrics_enabled       = local.microservice_config.metrics.enabled
    }
  }
}
```

## Signature

<!-- signature generated by tfplugindocs -->
```text
jsonschema_parse(schema_source string, target_source string) dynamic
```

## Arguments

<!-- arguments generated by tfplugindocs -->
1. `schema_source` (String) JSON Schema source: URL, file path, or inline JSON/YAML schema
1. `target_source` (String) Target source: URL, file path, or inline JSON/YAML value


## Return Type

The return type of `jsonschema_parse` is a dynamic object containing parsed and validated data with schema defaults applied. The structure of the returned object matches the schema definition.

## Behavior

### Schema Validation
- The schema source is resolved from URL/path/inline and compiled for validation
- The target source is resolved from URL/path/inline and parsed as JSON or YAML
- If validation fails, the function returns an error with details about what failed
- Supports JSON Schema Draft 2020-12 format

### Default Value Application
- Default values defined in the schema are automatically applied to missing properties
- Nested objects receive defaults recursively
- Defaults are materialized recursively for nested object schemas (including `default: {}` patterns)
- Empty objects (`{}`) can still materialize nested defaults where schema defaults define them

### Source Resolution and Format Detection
- Resolution order is URL (`http://` or `https://`), then file path lookup, then inline content
- JSON parsing is attempted first, then YAML parsing
- Relative file paths are resolved from Terraform execution context
- Clear errors are returned for unreachable URLs, file access failures, parse failures, or schema compilation errors

### Data Type Conversion
- YAML data is converted to appropriate Terraform types
- Numbers are converted to Int64 or Float64 as appropriate
- Booleans, strings, arrays, and objects are preserved with proper typing
- Complex nested structures are fully supported

## Common Use Cases

### Configuration Management
Use `jsonschema_parse` to load and validate application configuration files:

```hcl
locals {
  app_config = provider::helpers::jsonschema_parse(
    "${path.module}/schemas/app-config.schema.json",
    "${path.module}/configs/${var.environment}.yaml"
  )
}

resource "kubernetes_config_map" "app_config" {
  metadata {
    name = "app-config"
  }
  
  data = {
    "config.json" = jsonencode(local.app_config)
  }
}
```

### Microservice Configuration
Standardize microservice configurations across your infrastructure:

```hcl
locals {
  service_config = provider::helpers::jsonschema_parse(
    "${path.module}/schemas/service.schema.json",
    "${path.module}/services/${var.service_name}/config.yaml"
  )
}

resource "helm_release" "microservice" {
  name  = var.service_name
  chart = "./charts/microservice"
  
  values = [
    jsonencode({
      service = local.service_config.service
      logging = local.service_config.logging
      metrics = local.service_config.metrics
    })
  ]
}
```

### Infrastructure as Code Templates
Create reusable infrastructure templates with validated configurations:

```hcl
locals {
  infrastructure_config = provider::helpers::jsonschema_parse(
    "${path.module}/schemas/infrastructure.schema.json", 
    var.config_file
  )
}

module "vpc" {
  source = "./modules/vpc"
  
  cidr_block           = local.infrastructure_config.network.vpc_cidr
  availability_zones   = local.infrastructure_config.network.availability_zones
  enable_dns_hostnames = local.infrastructure_config.network.dns_hostnames
}
```

## Error Handling

The function will return errors in the following scenarios:

- **Source Access Errors**: When schema/target URL or file cannot be read
- **JSON Schema Compilation Errors**: When the schema source contains invalid schema definitions
- **Parse Errors**: When schema or target content is not valid JSON/YAML
- **Schema Validation Errors**: When target content doesn't conform to the JSON Schema requirements

Error messages are descriptive and include the specific reason for failure to help with debugging.

## Best Practices

1. **Schema Design**: Design your JSON schemas with sensible defaults to minimize configuration requirements
2. **Error Handling**: Use Terraform's error handling capabilities to provide fallback configurations
3. **File Organization**: Keep schema files in a dedicated directory and version them alongside your Terraform code
4. **Validation**: Use required fields in your schema to enforce critical configuration values
5. **Documentation**: Document your schemas and provide example YAML files for users

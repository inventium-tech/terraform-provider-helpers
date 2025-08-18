# Example usage of the jsonschema_yaml_parse function
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
  content = <<-EOT
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
  content = <<-EOT
    version: "0.1.0"
  EOT
}

resource "local_file" "partial_config" {
  filename = "${path.module}/partial_config.yaml"
  content = <<-EOT
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
  value       = provider::helpers::jsonschema_yaml_parse(
    local_file.sample_schema.filename,
    local_file.complete_config.filename
  )
}

# Parse a minimal YAML configuration (defaults will be applied)
output "minimal_parsed_config" {
  description = "Parse minimal YAML - defaults should be applied"
  value       = provider::helpers::jsonschema_yaml_parse(
    local_file.sample_schema.filename,
    local_file.minimal_config.filename
  )
}

# Parse a partial YAML configuration
output "partial_parsed_config" {
  description = "Parse partial YAML configuration with mixed defaults"
  value       = provider::helpers::jsonschema_yaml_parse(
    local_file.sample_schema.filename,
    local_file.partial_config.filename
  )
}

## PRACTICAL USE CASES

# Extract specific configuration values
locals {
  parsed_config = provider::helpers::jsonschema_yaml_parse(
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
      timeout          = local.parsed_config.config.timeout
      max_retries      = local.parsed_config.config.retries
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
  content = <<-EOT
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
  value       = provider::helpers::jsonschema_yaml_parse(
    local_file.microservice_schema.filename,
    local_file.microservice_config.filename
  )
}

# Use in conditional logic
locals {
  microservice_config = provider::helpers::jsonschema_yaml_parse(
    local_file.microservice_schema.filename,
    local_file.microservice_config.filename
  )
  
  should_enable_metrics = local.microservice_config.metrics.enabled
  metrics_port         = local.microservice_config.metrics.port
}

output "conditional_metrics_config" {
  description = "Metrics configuration only if enabled"
  value = local.should_enable_metrics ? {
    enabled = true
    port    = local.metrics_port
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
#   value       = provider::helpers::jsonschema_yaml_parse(
#     local_file.sample_schema.filename,
#     local_file.invalid_config.filename
#   )
# }

## INTEGRATION WITH OTHER TERRAFORM FEATURES

# Use parsed configuration with for_each
locals {
  services_config = provider::helpers::jsonschema_yaml_parse(
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
      name = local.microservice_config.service.name
      port = local.microservice_config.service.port
      logging_level = local.microservice_config.logging.level
    }
    default_values_applied = {
      database_host = local.parsed_config.config.database.host
      health_check_endpoint = local.microservice_config.service.health_check.endpoint
      metrics_enabled = local.microservice_config.metrics.enabled
    }
  }
}

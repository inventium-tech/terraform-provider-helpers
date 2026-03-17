locals {
  schema_inline = jsonencode({
    type = "object"
    properties = {
      name    = { type = "string" }
      enabled = { type = "boolean" }
    }
    required = ["name", "enabled"]
  })

  valid_target_inline = jsonencode({
    name    = "service-a"
    enabled = true
  })

  invalid_target_inline = jsonencode({
    name = "service-a"
  })
}

output "inline_valid" {
  value = provider::helpers::jsonschema_validate(local.schema_inline, local.valid_target_inline)
}

output "inline_invalid" {
  value = provider::helpers::jsonschema_validate(local.schema_inline, local.invalid_target_inline)
}

# Example file-path usage
resource "local_file" "schema_file" {
  filename = "${path.module}/schema.yaml"
  content  = <<-YAML
    type: object
    properties:
      replicas:
        type: integer
    required:
      - replicas
  YAML
}

resource "local_file" "target_file" {
  filename = "${path.module}/target.yaml"
  content  = <<-YAML
    replicas: 3
  YAML
}

output "file_valid" {
  value = provider::helpers::jsonschema_validate(local_file.schema_file.filename, local_file.target_file.filename)
}

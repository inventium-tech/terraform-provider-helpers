package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/kaptinlin/jsonschema"
	"gopkg.in/yaml.v3"
)

var _ function.Function = &JsonschemaYamlParseFunction{}

type JsonschemaYamlParseFunction struct{}

func NewJsonschemaYamlParseFunction() function.Function {
	return &JsonschemaYamlParseFunction{}
}

func (j JsonschemaYamlParseFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "jsonschema_yaml_parse"
}

func (j JsonschemaYamlParseFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Parse and validate YAML file against JSON Schema with defaults.",
		Description: "Reads a YAML file, validates it against a JSON Schema, and returns a dynamic object with schema defaults applied. The function provides schema validation and applies default values where specified.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:               "schema_file",
				Description:        "Path to the JSON Schema file for validation and default values",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
			function.StringParameter{
				Name:               "target_file",
				Description:        "Path to the YAML file to parse and validate",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
		},

		Return: function.DynamicReturn{},
	}
}

func (j JsonschemaYamlParseFunction) Run(ctx context.Context, request function.RunRequest, resp *function.RunResponse) {
	var schemaFile types.String
	var targetFile types.String

	// Get the parameters
	err := request.Arguments.Get(ctx, &schemaFile, &targetFile)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Error reading function arguments: %s", err.Error()))
		return
	}

	// Read and parse JSON Schema file
	schemaData, fileErr := readFile(schemaFile.ValueString())
	if fileErr != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Error reading schema file '%s': %s", schemaFile.ValueString(), fileErr.Error()))
		return
	}

	// Create JSON Schema compiler and compile schema
	compiler := jsonschema.NewCompiler()
	schema, compileErr := compiler.Compile(schemaData)
	if compileErr != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Error compiling schema: %s", compileErr.Error()))
		return
	}

	// Read and parse YAML file
	yamlData, yamlFileErr := readFile(targetFile.ValueString())
	if yamlFileErr != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Error reading target file '%s': %s", targetFile.ValueString(), yamlFileErr.Error()))
		return
	}

	// Parse YAML to interface{}
	var yamlContent interface{}
	if err := yaml.Unmarshal(yamlData, &yamlContent); err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Error parsing YAML file '%s': %s", targetFile.ValueString(), err.Error()))
		return
	}

	// Validate against schema
	result := schema.Validate(yamlContent)
	if !result.IsValid() {
		resp.Error = function.NewFuncError(fmt.Sprintf("Schema validation failed: %s", result.Error()))
		return
	}

	// Apply defaults from schema
	defaultedContent := applySchemaDefaults(schema, yamlContent)

	// Convert the result to Terraform dynamic value
	terraformValue, convertErr := convertToTerraformValue(ctx, defaultedContent)
	if convertErr != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Error converting to Terraform value: %s", convertErr.Error()))
		return
	}

	// Set the result in the response
	setErr := resp.Result.Set(ctx, terraformValue)
	if setErr != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Error setting result: %s", setErr.Error()))
		return
	}
}

// readFile reads a file and returns its contents as bytes
func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

// applySchemaDefaults applies default values from the schema to the content
func applySchemaDefaults(schema *jsonschema.Schema, content interface{}) interface{} {
	return applyDefaultsRecursively(schema, content)
}

// applyDefaultsRecursively recursively applies defaults from schema to content
func applyDefaultsRecursively(schema *jsonschema.Schema, content interface{}) interface{} {
	if schema == nil {
		return content
	}

	// If content is nil and schema has a default, use the default
	if content == nil && schema.Default != nil {
		return schema.Default
	}

	// Handle object types
	if contentMap, ok := content.(map[string]interface{}); ok {
		result := make(map[string]interface{})

		// Copy existing values
		for k, v := range contentMap {
			result[k] = v
		}

		// Only apply defaults for missing properties if the object has some content
		// or if this is not the root level (to avoid applying defaults to explicitly empty objects)
		if len(result) > 0 && schema.Properties != nil {
			for propName, propSchema := range *schema.Properties {
				if _, exists := result[propName]; !exists && propSchema.Default != nil {
					// Property missing, apply default
					result[propName] = propSchema.Default
				} else if exists && result[propName] != nil {
					// Property exists, recursively apply defaults to nested content
					result[propName] = applyDefaultsRecursively(propSchema, result[propName])
				}
			}
		} else if len(result) > 0 && schema.Properties != nil {
			// For non-empty objects, recursively apply defaults to existing properties
			for propName, propSchema := range *schema.Properties {
				if val, exists := result[propName]; exists && val != nil {
					result[propName] = applyDefaultsRecursively(propSchema, result[propName])
				}
			}
		}

		return result
	}

	// Handle array types
	if contentArray, ok := content.([]interface{}); ok {
		result := make([]interface{}, len(contentArray))
		for i, item := range contentArray {
			if schema.Items != nil {
				result[i] = applyDefaultsRecursively(schema.Items, item)
			} else {
				result[i] = item
			}
		}
		return result
	}

	// For primitive types, return as is
	return content
}

// convertToTerraformValue converts a Go interface{} to a Terraform dynamic value
func convertToTerraformValue(ctx context.Context, data interface{}) (basetypes.DynamicValue, error) {
	// Convert the interface{} to JSON then back to ensure proper type handling
	jsonData, err := json.Marshal(data)
	if err != nil {
		return basetypes.DynamicValue{}, fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	var normalizedData interface{}
	if err := json.Unmarshal(jsonData, &normalizedData); err != nil {
		return basetypes.DynamicValue{}, fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	// Convert to Terraform value
	terraformValue, err := convertInterfaceToTerraformValue(ctx, normalizedData)
	if err != nil {
		return basetypes.DynamicValue{}, fmt.Errorf("failed to convert to Terraform value: %w", err)
	}

	return basetypes.NewDynamicValue(terraformValue), nil
}

// convertInterfaceToTerraformValue recursively converts interface{} to terraform attr.Value
func convertInterfaceToTerraformValue(ctx context.Context, data interface{}) (attr.Value, error) {
	if data == nil {
		return types.DynamicNull(), nil
	}

	switch v := data.(type) {
	case bool:
		return types.BoolValue(v), nil
	case int:
		return types.Int64Value(int64(v)), nil
	case int32:
		return types.Int64Value(int64(v)), nil
	case int64:
		return types.Int64Value(v), nil
	case float32:
		return types.Float64Value(float64(v)), nil
	case float64:
		return types.Float64Value(v), nil
	case string:
		return types.StringValue(v), nil
	case map[string]interface{}:
		attrTypes := make(map[string]attr.Type)
		attrValues := make(map[string]attr.Value)

		for key, value := range v {
			terraformValue, err := convertInterfaceToTerraformValue(ctx, value)
			if err != nil {
				return types.DynamicNull(), fmt.Errorf("failed to convert map value for key '%s': %w", key, err)
			}
			attrTypes[key] = terraformValue.Type(ctx)
			attrValues[key] = terraformValue
		}

		objectValue, diags := types.ObjectValue(attrTypes, attrValues)
		if diags.HasError() {
			return types.DynamicNull(), fmt.Errorf("failed to create object value: %s", diags.Errors())
		}
		return objectValue, nil
	case []interface{}:
		if len(v) == 0 {
			return types.ListValueMust(types.DynamicType, []attr.Value{}), nil
		}

		// Convert all elements to dynamic values
		elements := make([]attr.Value, len(v))
		for i, item := range v {
			terraformValue, err := convertInterfaceToTerraformValue(ctx, item)
			if err != nil {
				return types.DynamicNull(), fmt.Errorf("failed to convert array element at index %d: %w", i, err)
			}
			// Wrap each element in a dynamic value to ensure consistent type
			elements[i] = basetypes.NewDynamicValue(terraformValue)
		}

		// Use dynamic type for all arrays to handle mixed types
		listValue, diags := types.ListValue(types.DynamicType, elements)
		if diags.HasError() {
			return types.DynamicNull(), fmt.Errorf("failed to create list value: %s", diags.Errors())
		}
		return listValue, nil
	default:
		return types.DynamicNull(), fmt.Errorf("unsupported data type: %T", data)
	}
}

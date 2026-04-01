package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ function.Function = &JsonschemaParseFunction{}

type JsonschemaParseFunction struct{}

func NewJsonschemaParseFunction() function.Function {
	return &JsonschemaParseFunction{}
}

func (j JsonschemaParseFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "jsonschema_parse"
}

func (j JsonschemaParseFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Parse and validate data against JSON Schema with defaults.",
		Description: "Resolves schema and target from URL, file path, or inline content; validates target against schema; then returns parsed data with schema defaults applied recursively.",

		Parameters: jsonSchemaSourceParameters(),

		Return: function.DynamicReturn{},
	}
}

func (j JsonschemaParseFunction) Run(ctx context.Context, request function.RunRequest, resp *function.RunResponse) {
	ctx = tflog.SetField(ctx, "function_name", jsonschemaParseFunctionName)

	schemaSource, targetSource, err := readJSONSchemaSources(ctx, request)
	if err != nil {
		tflog.Error(ctx, "Failed to read function arguments", map[string]interface{}{
			"stage":      "read_arguments",
			"error_kind": "argument_decode_failed",
			"error":      err.Error(),
		})
		resp.Error = function.NewFuncError(err.Error())
		return
	}

	defaultedContent, processErr := processJSONSchemaParse(ctx, schemaSource, targetSource)
	if processErr != nil {
		resp.Error = function.NewFuncError(processErr.Error())
		return
	}

	terraformValue, convertErr := convertToTerraformDynamicValue(ctx, defaultedContent)
	if convertErr != nil {
		tflog.Error(ctx, "Failed to convert value to Terraform dynamic value", map[string]interface{}{
			"stage":      "convert_result",
			"error_kind": "terraform_conversion_failed",
			"error":      convertErr.Error(),
		})
		resp.Error = function.NewFuncError(fmt.Sprintf("Error converting to Terraform value: %s", convertErr.Error()))
		return
	}

	// Set the result in the response
	setErr := resp.Result.Set(ctx, terraformValue)
	if setErr != nil {
		tflog.Error(ctx, "Failed to set function result", map[string]interface{}{
			"stage":      "set_result",
			"error_kind": "result_set_failed",
			"error":      setErr.Error(),
		})
		resp.Error = function.NewFuncError(fmt.Sprintf("Error setting result: %s", setErr.Error()))
		return
	}
}

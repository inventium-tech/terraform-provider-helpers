package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ function.Function = &JsonschemaValidateFunction{}

type JsonschemaValidateFunction struct{}

func NewJsonschemaValidateFunction() function.Function {
	return &JsonschemaValidateFunction{}
}

func (j JsonschemaValidateFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "jsonschema_validate"
}

func (j JsonschemaValidateFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Validate data against JSON Schema.",
		Description: "Resolves schema and target from URL, file path, or inline JSON/YAML content; validates target against schema; returns true when valid and false when schema validation fails.",

		Parameters: jsonSchemaSourceParameters(),

		Return: function.BoolReturn{},
	}
}

func (j JsonschemaValidateFunction) Run(ctx context.Context, request function.RunRequest, resp *function.RunResponse) {
	ctx = tflog.SetField(ctx, "function_name", jsonschemaValidateFunctionName)

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

	isValid, processErr := processJSONSchemaValidate(ctx, schemaSource, targetSource)
	if processErr != nil {
		resp.Error = function.NewFuncError(processErr.Error())
		return
	}

	setErr := resp.Result.Set(ctx, isValid)
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

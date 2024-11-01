package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ function.Function = &ObjectSetValueFunction{}

type ObjectSetValueFunction struct{}

func NewObjectSetValueFunction() function.Function {
	return &ObjectSetValueFunction{}
}

func (c ObjectSetValueFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "object_set_value"
}

func (c ObjectSetValueFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Sets a value in an Object or creates a new key with the value",
		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:               "object",
				Description:        "The Object to set the value in",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
			function.StringParameter{
				Name:               "key",
				Description:        "The key to set the value in",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
			function.DynamicParameter{
				Name:               "value",
				Description:        "The value to set in the key",
				AllowNullValue:     true,
				AllowUnknownValues: true,
			},
			function.StringParameter{
				Name:               "operation",
				Description:        "The operation mode to use when setting the value",
				AllowNullValue:     false,
				AllowUnknownValues: false,
				Validators: []function.StringParameterValidator{
					stringvalidator.OneOf("write_all", "write_value", "write_safe"),
				},
			},
		},
		Return: function.DynamicReturn{},
	}
}

func (c ObjectSetValueFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var object types.Dynamic
	var key string
	var value types.Dynamic
	var operation string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &object, &key, &value, &operation))
	if resp.Error != nil {
		return
	}

	objectParsed := object.UnderlyingValue().(types.Object)
	attrTypes := objectParsed.AttributeTypes(ctx)
	attrValues := objectParsed.Attributes()

	var isCurrentValueNull, keyExists = true, false

	if _, ok := attrValues[key]; ok {
		isCurrentValueNull = attrValues[key].IsNull() || attrValues[key].Equal(basetypes.NewStringValue(""))
		//asNull := attrValues[key].IsNull()
		//asString := attrValues[key].Equal(basetypes.NewStringValue(""))
		//isCurrentValueNull = asNull || asString == ""
		//isCurrentValueNull = attrValues[key].IsNull() || attrValues[key].String() == ""
		keyExists = true
	}

	attrTypes[key] = value.UnderlyingValue().Type(ctx)
	attrValues[key] = value.UnderlyingValue()

	var result basetypes.ObjectValue
	var diags diag.Diagnostics

	isWriteSafeOp := operation == "write_safe" && keyExists && isCurrentValueNull
	isWriteValueOp := operation == "write_value" && keyExists

	if !isWriteSafeOp && !isWriteValueOp && operation != "write_all" {
		result, diags = objectParsed.ToObjectValue(ctx)
	} else {
		result, diags = basetypes.NewObjectValue(attrTypes, attrValues)
	}

	resp.Error = function.ConcatFuncErrors(function.FuncErrorFromDiags(ctx, diags))
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, basetypes.NewDynamicValue(result)))
}

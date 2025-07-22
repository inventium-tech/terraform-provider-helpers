package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ function.Function = &ObjectFilterKeysFunction{}

type ObjectFilterKeysFunction struct{}

func NewObjectFilterKeysFunction() function.Function {
	return &ObjectFilterKeysFunction{}
}

func (o ObjectFilterKeysFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "object_filter_keys"
}

func (o ObjectFilterKeysFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Filter object keys based on a set of target keys.",
		Description: "Returns a new object containing only the keys that match the provided set of keys. Works with both objects and maps.",

		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:               "object",
				Description:        "The object or map to filter keys from",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
			function.SetParameter{
				ElementType:        types.StringType,
				Name:               "keys",
				Description:        "Set of keys to keep in the filtered object",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
		},

		Return: function.DynamicReturn{},
	}
}

func (o ObjectFilterKeysFunction) Run(ctx context.Context, request function.RunRequest, resp *function.RunResponse) {
	var object types.Dynamic
	var keys types.Set

	// Get the parameters
	err := request.Arguments.Get(ctx, &object, &keys)
	if err != nil {
		resp.Error = err
		return
	}

	// Convert the set of keys to a map for faster lookup
	keyElements := keys.Elements()
	allowedKeys := make(map[string]bool, len(keyElements))
	for _, keyElement := range keyElements {
		if keyStr, ok := keyElement.(types.String); ok {
			allowedKeys[keyStr.ValueString()] = true
		}
	}

	// Handle both Object and Map types
	underlyingValue := object.UnderlyingValue()

	filteredAttrTypes := make(map[string]attr.Type)
	filteredAttrValues := make(map[string]attr.Value)

	switch v := underlyingValue.(type) {
	case types.Object:
		// Handle values from the type: Object
		attrTypes := v.AttributeTypes(ctx)
		attrValues := v.Attributes()

		for key, value := range attrValues {
			if allowedKeys[key] {
				filteredAttrTypes[key] = attrTypes[key]
				filteredAttrValues[key] = value
			}
		}
	case types.Map:
		// Handle values from the type: Map
		elementType := v.ElementType(ctx)
		mapElements := v.Elements()

		for key, value := range mapElements {
			if allowedKeys[key] {
				filteredAttrTypes[key] = elementType
				filteredAttrValues[key] = value
			}
		}
	default:
		resp.Error = function.NewFuncError("First parameter must be an object or map")
		return
	}

	// Create the filtered object
	result, diags := basetypes.NewObjectValue(filteredAttrTypes, filteredAttrValues)
	if diags.HasError() {
		resp.Error = function.FuncErrorFromDiags(ctx, diags)
		return
	}

	// Set the result in the response
	err = resp.Result.Set(ctx, basetypes.NewDynamicValue(result))
	if err != nil {
		resp.Error = err
		return
	}
}

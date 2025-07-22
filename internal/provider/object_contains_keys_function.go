package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-helpers/internal/utils/xslices"
)

var _ function.Function = &ObjectContainsKeysFunction{}

type ObjectContainsKeysFunction struct{}

func NewObjectContainsKeysFunction() function.Function {
	return &ObjectContainsKeysFunction{}
}

func (o ObjectContainsKeysFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "object_contains_keys"
}

func (o ObjectContainsKeysFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Check if an object contains a target set of keys.",
		Description: `Returns true if the object contains the specified keys. When strict=true (default), all keys must 
		be present. When strict=false, at least one key must be present. Works with both objects and maps.`,

		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:               "object",
				Description:        "The object or map to check for keys",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
			function.SetParameter{
				ElementType:        types.StringType,
				Name:               "keys",
				Description:        "Set of keys to check for in the object",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
		},
		VariadicParameter: function.BoolParameter{
			Name:               "strict",
			Description:        "When true (default), all keys must be present. When false, at least one key must be present",
			AllowNullValue:     false,
			AllowUnknownValues: false,
		},

		Return: function.BoolReturn{},
	}
}

func (o ObjectContainsKeysFunction) Run(ctx context.Context, request function.RunRequest, resp *function.RunResponse) {
	var object types.Dynamic
	var keys types.Set
	var strictTuple types.Tuple

	// Get the parameters
	err := request.Arguments.Get(ctx, &object, &keys, &strictTuple)
	if err != nil {
		resp.Error = err
		return
	}

	// Default strict to true if not provided
	strict := true
	if len(strictTuple.Elements()) > 0 {
		strict = strictTuple.Elements()[0].(types.Bool).ValueBool()
	}

	// Check the length of keys to ensure it is not empty (fail early)
	targetKeysLength := len(keys.Elements())
	if targetKeysLength == 0 {
		resp.Error = resp.Result.Set(ctx, false)
		return
	}

	// Convert the set of keys to a slice for checking
	targetKeys := make([]string, 0, targetKeysLength)
	diags := keys.ElementsAs(ctx, &targetKeys, false)
	if diags.HasError() {
		resp.Error = function.FuncErrorFromDiags(ctx, diags)
		return
	}

	// Handle both Object and Map types
	underlyingValue := object.UnderlyingValue()
	var objectKeys []string

	switch v := underlyingValue.(type) {
	case types.Object:
		// Handle values from the type: Object
		attrValues := v.Attributes()
		objectKeys = make([]string, len(attrValues))
		for key := range attrValues {
			objectKeys = append(objectKeys, key)
		}
	case types.Map:
		// Handle values from the type: Map
		mapElements := v.Elements()
		objectKeys = make([]string, len(mapElements))
		for key := range mapElements {
			objectKeys = append(objectKeys, key)
		}
	default:
		resp.Error = function.NewFuncError("First parameter must be an object or map")
		return
	}

	intersectedKeys := xslices.Intersection(objectKeys, targetKeys)

	switch len(intersectedKeys) {
	case 0:
		// No keys found
		resp.Error = resp.Result.Set(ctx, false)
	case len(targetKeys):
		// All target keys found
		resp.Error = resp.Result.Set(ctx, true)
	default:
		// Some but not all the target keys found
		resp.Error = resp.Result.Set(ctx, !strict)
	}
}

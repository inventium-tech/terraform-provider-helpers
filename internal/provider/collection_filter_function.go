package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"terraform-provider-helpers/internal/validators/dynamicvalidator"
)

type CollectionFilterFunction struct{}

var _ function.Function = &CollectionFilterFunction{}

func NewCollectionFilterFunction() function.Function {
	return &CollectionFilterFunction{}
}

func (o CollectionFilterFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "collection_filter"
}

func (o CollectionFilterFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Filter collection of objects.",
		Description: "Filter a collection of objects using a simple comparison.",

		Parameters: []function.Parameter{
			function.DynamicParameter{
				Name:               "collection",
				Description:        "The collection of objects to filter",
				AllowNullValue:     false,
				AllowUnknownValues: false,
				Validators: []function.DynamicParameterValidator{
					dynamicvalidator.ElementsOfSameTypeValidator{},
				},
			},
			function.StringParameter{
				Name:               "key",
				Description:        "The key from the object to filter by",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
			function.DynamicParameter{
				Name:               "value",
				Description:        "The value used to compare against",
				AllowNullValue:     true,
				AllowUnknownValues: false,
			},
		},

		Return: function.DynamicReturn{},
	}
}

func (o CollectionFilterFunction) Run(ctx context.Context, request function.RunRequest, resp *function.RunResponse) {
	var collection types.Dynamic
	var value types.Dynamic
	var key string
	var filteredTypes []attr.Type
	var filteredValues []attr.Value

	if err := request.Arguments.Get(ctx, &collection, &key, &value); err != nil {
		resp.Error = err
		return
	}

	// cast validation of the collection parameter is done by the ElementsOfSameTypeValidator
	collectionParsed, _ := collection.UnderlyingValue().(types.Tuple)

	var valueParsed tftypes.Value
	var valueCastErr error
	if value.IsNull() {
		valueParsed = tftypes.NewValue(tftypes.DynamicPseudoType, nil)
	} else if valueParsed, valueCastErr = value.UnderlyingValue().ToTerraformValue(ctx); valueCastErr != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(valueCastErr.Error()))
		return
	}

	elements := collectionParsed.Elements()
	elementTypes := collectionParsed.ElementTypes(ctx)
	//baseType := elementTypes[0].TerraformType(ctx)

	for i, elem := range elements {
		found := false
		targetValue := elem

		if elemAsObject, isObject := elem.(types.Object); isObject {
			attrs := elemAsObject.Attributes()
			flattenObject := FlatObjectMap(ctx, attrs)
			if targetValue, found = flattenObject[key]; !found {
				continue
			}
		}

		targetTFValue, toTFErr := targetValue.ToTerraformValue(ctx)
		if toTFErr != nil {
			resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(toTFErr.Error()))
			return
		}

		if targetTFValue.Equal(valueParsed) {
			filteredTypes = append(filteredTypes, elementTypes[i])
			filteredValues = append(filteredValues, elem)
		}
	}

	resultList := basetypes.NewTupleValueMust(filteredTypes, filteredValues)
	if err := resp.Result.Set(ctx, basetypes.NewDynamicValue(resultList)); err != nil {
		resp.Error = err
		return
	}
}

// FlatObjectMap recursively flattens a map of attr.Value objects into a map of interface{} objects.
// It will only flatten values that are of type types.Object.
func FlatObjectMap(ctx context.Context, elements map[string]attr.Value) map[string]attr.Value {
	flatMap := make(map[string]attr.Value)
	for key, value := range elements {
		if obj, ok := value.(types.Object); ok {
			nestedMap := FlatObjectMap(ctx, obj.Attributes())
			for nestedKey, nestedValue := range nestedMap {
				flatMap[key+"."+nestedKey] = nestedValue
			}
		} else {
			flatMap[key] = value
		}
	}
	return flatMap
}

func MapLookup(ctx context.Context, m map[string]attr.Value, ks ...string) (r attr.Value, err error) {
	var ok bool
	var match interface{}
	var nestedMap map[string]attr.Value

	if len(ks) == 0 { // degenerate input
		return nil, fmt.Errorf("NestedMapLookup needs at least one key")
	}
	if r, ok = m[ks[0]]; !ok {
		return nil, fmt.Errorf("key not found; remaining keys: %v", ks)
	} else if len(ks) == 1 { // we've reached the final key
		return r, nil
	} else if r.Type(ctx).Equal(types.DynamicType) && !r.(types.Dynamic).UnderlyingValue().Type(ctx).Equal(types.TupleType{}) {
		nestedMap = r.(types.Dynamic).UnderlyingValue().(types.Object).Attributes()
		return r, nil
	} else if nestedMap, ok = match.(map[string]attr.Value); !ok {
		return nil, fmt.Errorf("malformed structure at %#v", match)
	} else { // 1+ more keys
		return MapLookup(ctx, nestedMap, ks[1:]...)
	}
}

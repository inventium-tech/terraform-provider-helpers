package provider

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Provider-shared normalization and Terraform value conversion helpers.
func normalizeGenericData(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case map[string]interface{}:
		normalized := make(map[string]interface{}, len(typedValue))
		for key, nestedValue := range typedValue {
			normalized[key] = normalizeGenericData(nestedValue)
		}
		return normalized
	case map[interface{}]interface{}:
		normalized := make(map[string]interface{}, len(typedValue))
		for key, nestedValue := range typedValue {
			normalized[fmt.Sprintf("%v", key)] = normalizeGenericData(nestedValue)
		}
		return normalized
	case []interface{}:
		normalized := make([]interface{}, len(typedValue))
		for index, item := range typedValue {
			normalized[index] = normalizeGenericData(item)
		}
		return normalized
	default:
		return typedValue
	}
}

func convertToTerraformDynamicValue(ctx context.Context, data interface{}) (basetypes.DynamicValue, error) {
	terraformValue, err := convertInterfaceToTerraformValue(ctx, normalizeGenericData(data))
	if err != nil {
		return basetypes.DynamicValue{}, fmt.Errorf("failed to convert to Terraform value: %w", err)
	}

	return basetypes.NewDynamicValue(terraformValue), nil
}

func convertInterfaceToTerraformValue(ctx context.Context, data interface{}) (attr.Value, error) {
	if data == nil {
		return types.DynamicNull(), nil
	}

	switch typedValue := data.(type) {
	case bool:
		return types.BoolValue(typedValue), nil
	case int:
		return types.Int64Value(int64(typedValue)), nil
	case int8:
		return types.Int64Value(int64(typedValue)), nil
	case int16:
		return types.Int64Value(int64(typedValue)), nil
	case int32:
		return types.Int64Value(int64(typedValue)), nil
	case int64:
		return types.Int64Value(typedValue), nil
	case uint:
		return types.Int64Value(int64(typedValue)), nil
	case uint8:
		return types.Int64Value(int64(typedValue)), nil
	case uint16:
		return types.Int64Value(int64(typedValue)), nil
	case uint32:
		return types.Int64Value(int64(typedValue)), nil
	case uint64:
		if typedValue > math.MaxInt64 {
			return types.DynamicNull(), fmt.Errorf("unsigned integer value %d overflows int64", typedValue)
		}
		return types.Int64Value(int64(typedValue)), nil
	case float32:
		return convertFloatToTerraformNumber(float64(typedValue)), nil
	case float64:
		return convertFloatToTerraformNumber(typedValue), nil
	case string:
		return types.StringValue(typedValue), nil
	case map[string]interface{}:
		attributeTypes := make(map[string]attr.Type, len(typedValue))
		attributeValues := make(map[string]attr.Value, len(typedValue))

		for key, nestedValue := range typedValue {
			convertedValue, err := convertInterfaceToTerraformValue(ctx, nestedValue)
			if err != nil {
				return types.DynamicNull(), fmt.Errorf("failed to convert map value for key '%s': %w", key, err)
			}
			attributeTypes[key] = convertedValue.Type(ctx)
			attributeValues[key] = convertedValue
		}

		objectValue, diags := types.ObjectValue(attributeTypes, attributeValues)
		if diags.HasError() {
			return types.DynamicNull(), fmt.Errorf("failed to create object value: %s", diags.Errors())
		}

		return objectValue, nil
	case []interface{}:
		if len(typedValue) == 0 {
			return types.ListValueMust(types.DynamicType, []attr.Value{}), nil
		}

		elements := make([]attr.Value, len(typedValue))
		for index, item := range typedValue {
			convertedValue, err := convertInterfaceToTerraformValue(ctx, item)
			if err != nil {
				return types.DynamicNull(), fmt.Errorf("failed to convert array element at index %d: %w", index, err)
			}
			elements[index] = basetypes.NewDynamicValue(convertedValue)
		}

		listValue, diags := types.ListValue(types.DynamicType, elements)
		if diags.HasError() {
			return types.DynamicNull(), fmt.Errorf("failed to create list value: %s", diags.Errors())
		}

		return listValue, nil
	default:
		return types.DynamicNull(), fmt.Errorf("unsupported data type: %T", data)
	}
}

func convertFloatToTerraformNumber(value float64) attr.Value {
	if value >= math.MinInt64 && value <= math.MaxInt64 && math.Trunc(value) == value {
		return types.Int64Value(int64(value))
	}

	return types.Float64Value(value)
}

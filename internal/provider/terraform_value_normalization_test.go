package provider

import (
	"context"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestNormalizeGenericData(t *testing.T) {
	t.Parallel()

	input := map[interface{}]interface{}{
		"name": "api",
		5:      []interface{}{map[interface{}]interface{}{"enabled": true}},
		"meta": map[interface{}]interface{}{"port": 8080},
	}

	normalized := normalizeGenericData(input)

	expected := map[string]interface{}{
		"name": "api",
		"5":    []interface{}{map[string]interface{}{"enabled": true}},
		"meta": map[string]interface{}{"port": 8080},
	}

	if !reflect.DeepEqual(normalized, expected) {
		t.Fatalf("unexpected normalized value\nexpected: %#v\nactual:   %#v", expected, normalized)
	}
}

func TestConvertFloatToTerraformNumber(t *testing.T) {
	t.Parallel()

	t.Run("integral float becomes int64", func(t *testing.T) {
		value := convertFloatToTerraformNumber(42.0)

		intValue, ok := value.(basetypes.Int64Value)
		if !ok {
			t.Fatalf("expected Int64Value, got: %T", value)
		}

		if intValue.ValueInt64() != 42 {
			t.Fatalf("expected 42, got %d", intValue.ValueInt64())
		}
	})

	t.Run("fractional float becomes float64", func(t *testing.T) {
		value := convertFloatToTerraformNumber(42.5)

		floatValue, ok := value.(basetypes.Float64Value)
		if !ok {
			t.Fatalf("expected Float64Value, got: %T", value)
		}

		if floatValue.ValueFloat64() != 42.5 {
			t.Fatalf("expected 42.5, got %v", floatValue.ValueFloat64())
		}
	})
}

func TestConvertInterfaceToTerraformValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("uint64 overflow returns deterministic error", func(t *testing.T) {
		_, err := convertInterfaceToTerraformValue(ctx, uint64(math.MaxInt64)+1)
		if err == nil {
			t.Fatal("expected overflow conversion error")
		}

		if !strings.Contains(err.Error(), "overflows int64") {
			t.Fatalf("expected overflow error, got: %v", err)
		}
	})

	t.Run("integral and fractional floats map to expected Terraform numbers", func(t *testing.T) {
		testCases := []struct {
			name          string
			input         float64
			expectInt64   bool
			expectedInt   int64
			expectedFloat float64
		}{
			{name: "integral float", input: 42.0, expectInt64: true, expectedInt: 42},
			{name: "fractional float", input: 42.5, expectInt64: false, expectedFloat: 42.5},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				value, err := convertInterfaceToTerraformValue(ctx, tc.input)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if tc.expectInt64 {
					intValue, ok := value.(basetypes.Int64Value)
					if !ok {
						t.Fatalf("expected Int64Value, got: %T", value)
					}

					if intValue.ValueInt64() != tc.expectedInt {
						t.Fatalf("expected int value %d, got %d", tc.expectedInt, intValue.ValueInt64())
					}
					return
				}

				floatValue, ok := value.(basetypes.Float64Value)
				if !ok {
					t.Fatalf("expected Float64Value, got: %T", value)
				}

				if floatValue.ValueFloat64() != tc.expectedFloat {
					t.Fatalf("expected float value %v, got %v", tc.expectedFloat, floatValue.ValueFloat64())
				}
			})
		}
	})

	t.Run("nested map/list values convert recursively", func(t *testing.T) {
		input := map[string]interface{}{
			"service": "api",
			"values":  []interface{}{float64(1), "two", map[string]interface{}{"enabled": true}},
		}

		value, err := convertInterfaceToTerraformValue(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		objectValue, ok := value.(basetypes.ObjectValue)
		if !ok {
			t.Fatalf("expected ObjectValue, got: %T", value)
		}

		attributes := objectValue.Attributes()

		serviceValue, ok := attributes["service"].(basetypes.StringValue)
		if !ok {
			t.Fatalf("expected service to be StringValue, got: %T", attributes["service"])
		}

		if serviceValue.ValueString() != "api" {
			t.Fatalf("unexpected service value: %q", serviceValue.ValueString())
		}

		valuesList, ok := attributes["values"].(basetypes.ListValue)
		if !ok {
			t.Fatalf("expected values to be ListValue, got: %T", attributes["values"])
		}

		elements := valuesList.Elements()
		if len(elements) != 3 {
			t.Fatalf("expected three list elements, got: %d", len(elements))
		}

		firstElement, ok := elements[0].(basetypes.DynamicValue)
		if !ok {
			t.Fatalf("expected first element to be DynamicValue, got: %T", elements[0])
		}

		firstUnderlying, ok := firstElement.UnderlyingValue().(basetypes.Int64Value)
		if !ok || firstUnderlying.ValueInt64() != 1 {
			t.Fatalf("expected first element underlying int64(1), got: %#v", firstElement.UnderlyingValue())
		}

		thirdElement, ok := elements[2].(basetypes.DynamicValue)
		if !ok {
			t.Fatalf("expected third element to be DynamicValue, got: %T", elements[2])
		}

		thirdUnderlyingObject, ok := thirdElement.UnderlyingValue().(basetypes.ObjectValue)
		if !ok {
			t.Fatalf("expected third element underlying object, got: %T", thirdElement.UnderlyingValue())
		}

		thirdAttributes := thirdUnderlyingObject.Attributes()
		enabledValue, ok := thirdAttributes["enabled"].(basetypes.BoolValue)
		if !ok {
			t.Fatalf("expected enabled to be BoolValue, got: %T", thirdAttributes["enabled"])
		}

		if !enabledValue.ValueBool() {
			t.Fatal("expected enabled=true in nested object")
		}
	})

	t.Run("empty list becomes dynamic list", func(t *testing.T) {
		value, err := convertInterfaceToTerraformValue(ctx, []interface{}{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		listValue, ok := value.(basetypes.ListValue)
		if !ok {
			t.Fatalf("expected ListValue, got: %T", value)
		}

		if len(listValue.Elements()) != 0 {
			t.Fatalf("expected empty list, got %d elements", len(listValue.Elements()))
		}

		if !reflect.DeepEqual(listValue.ElementType(ctx), types.DynamicType) {
			t.Fatalf("expected list element type DynamicType, got: %#v", listValue.ElementType(ctx))
		}
	})
}

func TestConvertInterfaceToTerraformValueScalarTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("nil input returns DynamicNull", func(t *testing.T) {
		value, err := convertInterfaceToTerraformValue(ctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		dynValue, ok := value.(basetypes.DynamicValue)
		if !ok || !dynValue.IsNull() {
			t.Fatalf("expected DynamicNull, got: %T / %v", value, value)
		}
	})

	t.Run("bool true", func(t *testing.T) {
		value, err := convertInterfaceToTerraformValue(ctx, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		boolVal, ok := value.(basetypes.BoolValue)
		if !ok || !boolVal.ValueBool() {
			t.Fatalf("expected BoolValue(true), got: %T / %v", value, value)
		}
	})

	t.Run("bool false", func(t *testing.T) {
		value, err := convertInterfaceToTerraformValue(ctx, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		boolVal, ok := value.(basetypes.BoolValue)
		if !ok || boolVal.ValueBool() {
			t.Fatalf("expected BoolValue(false), got: %T / %v", value, value)
		}
	})

	t.Run("string value", func(t *testing.T) {
		value, err := convertInterfaceToTerraformValue(ctx, "hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		strVal, ok := value.(basetypes.StringValue)
		if !ok || strVal.ValueString() != "hello" {
			t.Fatalf("expected StringValue(hello), got: %T / %v", value, value)
		}
	})

	intCases := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{"int", int(10), 10},
		{"int8", int8(10), 10},
		{"int16", int16(10), 10},
		{"int32", int32(10), 10},
		{"int64", int64(10), 10},
		{"uint", uint(10), 10},
		{"uint8", uint8(10), 10},
		{"uint16", uint16(10), 10},
		{"uint32", uint32(10), 10},
	}

	for _, tc := range intCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := convertInterfaceToTerraformValue(ctx, tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			intVal, ok := value.(basetypes.Int64Value)
			if !ok || intVal.ValueInt64() != tc.expected {
				t.Fatalf("expected Int64Value(%d), got: %T / %v", tc.expected, value, value)
			}
		})
	}

	t.Run("float32 integral becomes int64", func(t *testing.T) {
		value, err := convertInterfaceToTerraformValue(ctx, float32(7.0))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		intVal, ok := value.(basetypes.Int64Value)
		if !ok || intVal.ValueInt64() != 7 {
			t.Fatalf("expected Int64Value(7), got: %T / %v", value, value)
		}
	})

	t.Run("float32 fractional becomes float64", func(t *testing.T) {
		value, err := convertInterfaceToTerraformValue(ctx, float32(3.5))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		floatVal, ok := value.(basetypes.Float64Value)
		if !ok {
			t.Fatalf("expected Float64Value, got: %T", value)
		}

		// float32(3.5) converts exactly to float64(3.5)
		if floatVal.ValueFloat64() != float64(float32(3.5)) {
			t.Fatalf("unexpected float value: %v", floatVal.ValueFloat64())
		}
	})
}

func TestConvertInterfaceToTerraformValueUnsupportedType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("unsupported type returns error", func(t *testing.T) {
		_, err := convertInterfaceToTerraformValue(ctx, struct{ Name string }{"test"})
		if err == nil {
			t.Fatal("expected error for unsupported struct type")
		}

		if !strings.Contains(err.Error(), "unsupported data type") {
			t.Fatalf("expected 'unsupported data type' in error, got: %v", err)
		}
	})

	t.Run("nested map with unsupported value propagates error", func(t *testing.T) {
		input := map[string]interface{}{
			"valid_key":   "valid-value",
			"invalid_key": make(chan int),
		}

		_, err := convertInterfaceToTerraformValue(ctx, input)
		if err == nil {
			t.Fatal("expected error for nested unsupported type in map")
		}

		if !strings.Contains(err.Error(), "failed to convert map value for key") {
			t.Fatalf("expected map conversion error, got: %v", err)
		}
	})

	t.Run("nested list with unsupported element propagates error", func(t *testing.T) {
		input := []interface{}{
			"valid-string",
			make(chan int),
		}

		_, err := convertInterfaceToTerraformValue(ctx, input)
		if err == nil {
			t.Fatal("expected error for nested unsupported type in list")
		}

		if !strings.Contains(err.Error(), "failed to convert array element at index") {
			t.Fatalf("expected list conversion error, got: %v", err)
		}
	})
}

func TestConvertToTerraformDynamicValueError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	_, err := convertToTerraformDynamicValue(ctx, struct{ Ch chan int }{Ch: make(chan int)})
	if err == nil {
		t.Fatal("expected error for unsupported type in dynamic conversion")
	}

	if !strings.Contains(err.Error(), "failed to convert to Terraform value") {
		t.Fatalf("expected wrapped conversion error, got: %v", err)
	}
}

func TestConvertToTerraformDynamicValue(t *testing.T) {
	t.Parallel()

	input := map[interface{}]interface{}{
		"name":   "api",
		"nested": map[interface{}]interface{}{"port": 8080},
		"items":  []interface{}{map[interface{}]interface{}{"enabled": true}},
	}

	dynamicValue, err := convertToTerraformDynamicValue(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dynamicValue.IsNull() || dynamicValue.IsUnknown() {
		t.Fatalf("expected known dynamic value, got: %s", dynamicValue.String())
	}

	underlyingObject, ok := dynamicValue.UnderlyingValue().(basetypes.ObjectValue)
	if !ok {
		t.Fatalf("expected underlying ObjectValue, got: %T", dynamicValue.UnderlyingValue())
	}

	attributes := underlyingObject.Attributes()

	nameValue, ok := attributes["name"].(basetypes.StringValue)
	if !ok || nameValue.ValueString() != "api" {
		t.Fatalf("expected normalized name string, got: %#v", attributes["name"])
	}

	nestedObject, ok := attributes["nested"].(basetypes.ObjectValue)
	if !ok {
		t.Fatalf("expected nested object, got: %T", attributes["nested"])
	}

	nestedPortValue, ok := nestedObject.Attributes()["port"].(basetypes.Int64Value)
	if !ok || nestedPortValue.ValueInt64() != 8080 {
		t.Fatalf("expected nested port int64(8080), got: %#v", nestedObject.Attributes()["port"])
	}

	itemsValue, ok := attributes["items"].(basetypes.ListValue)
	if !ok {
		t.Fatalf("expected items list, got: %T", attributes["items"])
	}

	if len(itemsValue.Elements()) != 1 {
		t.Fatalf("expected one item element, got: %d", len(itemsValue.Elements()))
	}

	firstItemDynamic, ok := itemsValue.Elements()[0].(basetypes.DynamicValue)
	if !ok {
		t.Fatalf("expected first list element dynamic, got: %T", itemsValue.Elements()[0])
	}

	firstItemObject, ok := firstItemDynamic.UnderlyingValue().(basetypes.ObjectValue)
	if !ok {
		t.Fatalf("expected first dynamic underlying object, got: %T", firstItemDynamic.UnderlyingValue())
	}

	enabledValue, ok := firstItemObject.Attributes()["enabled"].(basetypes.BoolValue)
	if !ok || !enabledValue.ValueBool() {
		t.Fatalf("expected enabled bool true, got: %#v", firstItemObject.Attributes()["enabled"])
	}
}

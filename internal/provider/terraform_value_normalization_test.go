package provider

import (
	"context"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

		valuesTuple, ok := attributes["values"].(basetypes.TupleValue)
		if !ok {
			t.Fatalf("expected values to be TupleValue, got: %T", attributes["values"])
		}

		elements := valuesTuple.Elements()
		if len(elements) != 3 {
			t.Fatalf("expected three tuple elements, got: %d", len(elements))
		}

		firstElement, ok := elements[0].(basetypes.Int64Value)
		if !ok {
			t.Fatalf("expected first element to be Int64Value, got: %T", elements[0])
		}

		if firstElement.ValueInt64() != 1 {
			t.Fatalf("expected first element int64(1), got: %d", firstElement.ValueInt64())
		}

		thirdElement, ok := elements[2].(basetypes.ObjectValue)
		if !ok {
			t.Fatalf("expected third element to be ObjectValue, got: %T", elements[2])
		}

		thirdAttributes := thirdElement.Attributes()
		enabledValue, ok := thirdAttributes["enabled"].(basetypes.BoolValue)
		if !ok {
			t.Fatalf("expected enabled to be BoolValue, got: %T", thirdAttributes["enabled"])
		}

		if !enabledValue.ValueBool() {
			t.Fatal("expected enabled=true in nested object")
		}
	})

	t.Run("empty array becomes empty tuple", func(t *testing.T) {
		value, err := convertInterfaceToTerraformValue(ctx, []interface{}{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		tupleValue, ok := value.(basetypes.TupleValue)
		if !ok {
			t.Fatalf("expected TupleValue, got: %T", value)
		}

		if len(tupleValue.Elements()) != 0 {
			t.Fatalf("expected empty tuple, got %d elements", len(tupleValue.Elements()))
		}
	})
}

func TestConvertInterfaceToTerraformValueScalarTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testCases := []struct {
		name     string
		input    interface{}
		expected attr.Value
	}{
		{name: "nil input returns DynamicNull", input: nil, expected: basetypes.NewDynamicNull()},
		{name: "bool true", input: true, expected: basetypes.NewBoolValue(true)},
		{name: "bool false", input: false, expected: basetypes.NewBoolValue(false)},
		{name: "string value", input: "hello", expected: basetypes.NewStringValue("hello")},
		{name: "int", input: int(10), expected: basetypes.NewInt64Value(10)},
		{name: "int8", input: int8(10), expected: basetypes.NewInt64Value(10)},
		{name: "int16", input: int16(10), expected: basetypes.NewInt64Value(10)},
		{name: "int32", input: int32(10), expected: basetypes.NewInt64Value(10)},
		{name: "int64", input: int64(10), expected: basetypes.NewInt64Value(10)},
		{name: "uint", input: uint(10), expected: basetypes.NewInt64Value(10)},
		{name: "uint8", input: uint8(10), expected: basetypes.NewInt64Value(10)},
		{name: "uint16", input: uint16(10), expected: basetypes.NewInt64Value(10)},
		{name: "uint32", input: uint32(10), expected: basetypes.NewInt64Value(10)},
		{name: "float32 integral becomes int64", input: float32(7.0), expected: basetypes.NewInt64Value(7)},
		{name: "float32 fractional becomes float64", input: float32(3.5), expected: basetypes.NewFloat64Value(3.5)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, err := convertInterfaceToTerraformValue(ctx, tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !value.Equal(tc.expected) {
				t.Fatalf("expected %T / %v, got: %T / %v", tc.expected, tc.expected, value, value)
			}
		})
	}
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

	itemsValue, ok := attributes["items"].(basetypes.TupleValue)
	if !ok {
		t.Fatalf("expected items tuple, got: %T", attributes["items"])
	}

	if len(itemsValue.Elements()) != 1 {
		t.Fatalf("expected one item element, got: %d", len(itemsValue.Elements()))
	}

	firstItemObject, ok := itemsValue.Elements()[0].(basetypes.ObjectValue)
	if !ok {
		t.Fatalf("expected first tuple element object, got: %T", itemsValue.Elements()[0])
	}

	enabledValue, ok := firstItemObject.Attributes()["enabled"].(basetypes.BoolValue)
	if !ok || !enabledValue.ValueBool() {
		t.Fatalf("expected enabled bool true, got: %#v", firstItemObject.Attributes()["enabled"])
	}
}

func TestConvertToTerraformDynamicValueHeterogeneousArray(t *testing.T) {
	t.Parallel()

	input := []interface{}{
		map[string]interface{}{"oidc": "web"},
		map[string]interface{}{"saml": true},
	}

	dynamicValue, err := convertToTerraformDynamicValue(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error converting heterogeneous array: %v", err)
	}

	items, ok := dynamicValue.UnderlyingValue().(basetypes.TupleValue)
	if !ok {
		t.Fatalf("expected tuple value, got: %T", dynamicValue.UnderlyingValue())
	}
	if len(items.Elements()) != 2 {
		t.Fatalf("expected two tuple elements, got: %d", len(items.Elements()))
	}
	if _, ok := items.Elements()[0].(basetypes.ObjectValue); !ok {
		t.Fatalf("expected first element to retain its object type, got: %T", items.Elements()[0])
	}
	if _, ok := items.Elements()[1].(basetypes.ObjectValue); !ok {
		t.Fatalf("expected second element to retain its object type, got: %T", items.Elements()[1])
	}
}

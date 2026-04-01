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

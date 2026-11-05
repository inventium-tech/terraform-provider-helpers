package dynamicvalidator

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestElementsOfSameTypeValidatorNullPassthrough(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ElementsOfSameTypeValidator{}

	req := function.DynamicParameterValidatorRequest{
		ArgumentPosition: 0,
		Value:            types.DynamicNull(),
	}
	resp := &function.DynamicParameterValidatorResponse{}

	v.ValidateParameterDynamic(ctx, req, resp)

	if resp.Error != nil {
		t.Fatalf("expected no error for null value, got: %v", resp.Error)
	}
}

func TestElementsOfSameTypeValidatorUnknownPassthrough(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ElementsOfSameTypeValidator{}

	req := function.DynamicParameterValidatorRequest{
		ArgumentPosition: 0,
		Value:            types.DynamicUnknown(),
	}
	resp := &function.DynamicParameterValidatorResponse{}

	v.ValidateParameterDynamic(ctx, req, resp)

	if resp.Error != nil {
		t.Fatalf("expected no error for unknown value, got: %v", resp.Error)
	}
}

func TestElementsOfSameTypeValidatorHomogeneousTupleSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ElementsOfSameTypeValidator{}

	tupleValue := basetypes.NewTupleValueMust(
		[]attr.Type{types.StringType, types.StringType, types.StringType},
		[]attr.Value{types.StringValue("a"), types.StringValue("b"), types.StringValue("c")},
	)
	req := function.DynamicParameterValidatorRequest{
		ArgumentPosition: 0,
		Value:            basetypes.NewDynamicValue(tupleValue),
	}
	resp := &function.DynamicParameterValidatorResponse{}

	v.ValidateParameterDynamic(ctx, req, resp)

	if resp.Error != nil {
		t.Fatalf("expected no error for homogeneous string tuple, got: %v", resp.Error)
	}
}

func TestElementsOfSameTypeValidatorHomogeneousInt64TupleSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ElementsOfSameTypeValidator{}

	tupleValue := basetypes.NewTupleValueMust(
		[]attr.Type{types.Int64Type, types.Int64Type},
		[]attr.Value{types.Int64Value(1), types.Int64Value(2)},
	)
	req := function.DynamicParameterValidatorRequest{
		ArgumentPosition: 0,
		Value:            basetypes.NewDynamicValue(tupleValue),
	}
	resp := &function.DynamicParameterValidatorResponse{}

	v.ValidateParameterDynamic(ctx, req, resp)

	if resp.Error != nil {
		t.Fatalf("expected no error for homogeneous int64 tuple, got: %v", resp.Error)
	}
}

func TestElementsOfSameTypeValidatorMixedTypeTupleRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ElementsOfSameTypeValidator{}

	tupleValue := basetypes.NewTupleValueMust(
		[]attr.Type{types.StringType, types.Int64Type},
		[]attr.Value{types.StringValue("text"), types.Int64Value(42)},
	)
	req := function.DynamicParameterValidatorRequest{
		ArgumentPosition: 0,
		Value:            basetypes.NewDynamicValue(tupleValue),
	}
	resp := &function.DynamicParameterValidatorResponse{}

	v.ValidateParameterDynamic(ctx, req, resp)

	if resp.Error == nil {
		t.Fatal("expected error for mixed-type tuple (string + int64), got no error")
	}
}

func TestElementsOfSameTypeValidatorNonTupleRejected(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	v := ElementsOfSameTypeValidator{}

	// Passing a string as the underlying value — not a tuple
	req := function.DynamicParameterValidatorRequest{
		ArgumentPosition: 0,
		Value:            basetypes.NewDynamicValue(types.StringValue("not-a-collection")),
	}
	resp := &function.DynamicParameterValidatorResponse{}

	v.ValidateParameterDynamic(ctx, req, resp)

	if resp.Error == nil {
		t.Fatal("expected error for non-tuple input, got no error")
	}
}

package dynamicvalidator

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatorfuncerr"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-helpers/internal/utils/xslices"
)

var _ function.DynamicParameterValidator = ElementsOfSameTypeValidator{}

type ElementsOfSameTypeValidator struct{}

func (v ElementsOfSameTypeValidator) ValidateParameterDynamic(ctx context.Context, req function.DynamicParameterValidatorRequest, resp *function.DynamicParameterValidatorResponse) {
	if req.Value.IsNull() || req.Value.IsUnknown() {
		return
	}

	errorMsg := "value must be a list of elements of the same type"

	inputValue, ok := req.Value.UnderlyingValue().(types.Tuple)
	if !ok {
		resp.Error = validatorfuncerr.InvalidParameterValueMatchFuncError(
			req.ArgumentPosition,
			errorMsg,
			inputValue.String(),
		)
	}

	elementTypes := inputValue.ElementTypes(ctx)
	funcEvery := func(element attr.Type) bool {
		targetType := elementTypes[0].TerraformType(ctx)
		return element.TerraformType(ctx).Is(targetType)
	}

	if everyCheck := xslices.Every[attr.Type](elementTypes, funcEvery); !everyCheck {
		resp.Error = validatorfuncerr.InvalidParameterValueMatchFuncError(
			req.ArgumentPosition,
			errorMsg,
			inputValue.String(),
		)
	}

}

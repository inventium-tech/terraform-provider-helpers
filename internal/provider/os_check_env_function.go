package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
)

var _ function.Function = &OsCheckEnvFunction{}

type OsCheckEnvFunction struct{}

func NewOsCheckEnvFunction() function.Function {
	return &OsCheckEnvFunction{}
}

func (o OsCheckEnvFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "os_check_env"
}

func (o OsCheckEnvFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Check if an environment variable is set",
		Description: "Checks if an environment variable is set and optionally validates that it's not empty.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:               "name",
				Description:        "The name of the environment variable to check",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
		},
		VariadicParameter: function.BoolParameter{
			Name:               "strict",
			Description:        "When true (default), the variable cannot have an empty string value",
			AllowNullValue:     false,
			AllowUnknownValues: false,
		},

		Return: function.BoolReturn{},
	}
}

func (o OsCheckEnvFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var name string
	var strictTuple types.Tuple

	// Get the parameters
	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &name, &strictTuple))
	if resp.Error != nil {
		return
	}

	// Default strict to true if not provided
	strict := true
	if len(strictTuple.Elements()) > 0 {
		strict = strictTuple.Elements()[0].(types.Bool).ValueBool()
	}

	value, isPresent := os.LookupEnv(name)

	// return false if the variable is not present, OR
	// if strict is true and the value is an empty string
	if !isPresent || (strict && value == "") {
		resp.Error = resp.Result.Set(ctx, false)
		return
	}

	resp.Error = resp.Result.Set(ctx, true)
}

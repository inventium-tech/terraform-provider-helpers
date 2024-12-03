package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
)

var _ function.Function = &OsGetEnvFunction{}

type OsGetEnvFunction struct{}

func NewOsGetEnvFunction() function.Function {
	return &OsGetEnvFunction{}
}

func (o OsGetEnvFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "os_get_env"
}

func (o OsGetEnvFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Get an environment variable",
		Description: "Retrieve a single environment variable from the current process environment or use a fallback value if the environment variable is not set.",

		Parameters: []function.Parameter{
			function.StringParameter{
				Name:               "name",
				Description:        "The name of the environment variable to get",
				AllowNullValue:     false,
				AllowUnknownValues: false,
			},
		},
		VariadicParameter: function.StringParameter{
			Name:               "fallback",
			Description:        "The fallback value to use if the environment variable is not set",
			AllowNullValue:     false,
			AllowUnknownValues: false,
		},

		Return: function.StringReturn{},
	}
}

func (o OsGetEnvFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var name, result string
	var fallback types.Tuple

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &name, &fallback))
	if resp.Error != nil {
		return
	}

	if value, ok := os.LookupEnv(name); ok {
		result = value
	} else if len(fallback.Elements()) > 0 {
		// assign element 0 to result if fallback is set
		result = fallback.Elements()[0].(types.String).ValueString()
	}

	resp.Error = resp.Result.Set(ctx, result)
	return
}

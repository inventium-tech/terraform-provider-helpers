package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = &HelpersProvider{}
var _ provider.ProviderWithFunctions = &HelpersProvider{}

type HelpersProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (h *HelpersProvider) Metadata(ctx context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "helpers"
	resp.Version = h.version
}

func (h *HelpersProvider) Schema(ctx context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes:  map[string]schema.Attribute{},
		Description: "The Helpers Provides offers a set of functions to help with common tasks.",
	}
}

func (h *HelpersProvider) Configure(ctx context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
	tflog.Trace(ctx, "This provider does not require any configuration")
}

func (h *HelpersProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	tflog.Trace(ctx, "This provider does not have any data sources")
	return nil
}

func (h *HelpersProvider) Resources(ctx context.Context) []func() resource.Resource {
	tflog.Trace(ctx, "This provider does not have any resources")
	return nil
}

func (h *HelpersProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewObjectSetValueFunction,
	}
}

func NewProvider(version string) func() provider.Provider {
	return func() provider.Provider {
		return &HelpersProvider{
			version: version,
		}
	}
}

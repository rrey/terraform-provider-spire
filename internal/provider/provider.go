// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Ensure SpireProvider satisfies various provider interfaces.
var _ provider.Provider = &SpireProvider{}
var _ provider.ProviderWithFunctions = &SpireProvider{}

// SpireProvider defines the provider implementation.
type SpireProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// SpireProviderModel describes the provider data model.
type SpireProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *SpireProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "spire"
	resp.Version = p.version
}

func (p *SpireProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
		},
	}
}

func (p *SpireProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SpireProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := grpc.Dial("unix:/tmp/spire-server/private/api.sock", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to setup grpc connection",
			err.Error(),
		)
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *SpireProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSpireEntryResource,
	}
}

func (p *SpireProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSpireEntryDataSource,
	}
}

func (p *SpireProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SpireProvider{
			version: version,
		}
	}
}

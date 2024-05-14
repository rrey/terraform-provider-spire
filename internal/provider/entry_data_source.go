// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	entryv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	spireTypes "github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"google.golang.org/grpc"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SpireEntryDataSource{}

func NewSpireEntryDataSource() datasource.DataSource {
	return &SpireEntryDataSource{}
}

// SpireEntryDataSource defines the data source implementation.
type SpireEntryDataSource struct {
	client entryv1.EntryClient
}

// SpireEntryDataSourceModel describes the data source data model.
type SpireEntryDataSourceModel struct {
	Id        types.String           `tfsdk:"id"`
	SpiffeId  *SpiffeIdModel         `tfsdk:"spiffe_id"`
	ParentId  *SpiffeIdModel         `tfsdk:"parent_id"`
	Selectors *[]SpiffeSelectorModel `tfsdk:"selectors"`
}

func (d *SpireEntryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entry"
}

func (d *SpireEntryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Register workloads with Spire Server Entry",

		Attributes: map[string]schema.Attribute{
			"parent_id": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"trust_domain": types.StringType,
					"path":         types.StringType,
				},
				MarkdownDescription: "The SPIFFE ID of this record's parent",
				Computed:            true,
			},
			"spiffe_id": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"trust_domain": types.StringType,
					"path":         types.StringType,
				},
				MarkdownDescription: "The SPIFFE ID of this record's parent",
				Required:            true,
			},
			"selectors": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
						},
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
				MarkdownDescription: "A type/value selector. Can be used more than once",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Registration Entry ID of the record",
			},
		},
	}
}

func (d *SpireEntryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*grpc.ClientConn)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *grpc.ClientConn, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = entryv1.NewEntryClient(client)
}

func (d *SpireEntryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SpireEntryDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	entries, err := d.client.ListEntries(ctx, &entryv1.ListEntriesRequest{Filter: &entryv1.ListEntriesRequest_Filter{
		BySpiffeId: &spireTypes.SPIFFEID{
			TrustDomain: data.SpiffeId.TrustDomain.ValueString(),
			Path:        data.SpiffeId.Path.ValueString(),
		},
	}})
	if err != nil {
		resp.Diagnostics.AddError("Failed to list entries", err.Error())
		return
	}

	if len(entries.Entries) != 1 {
		resp.Diagnostics.AddError("Failed to find resource matching datasource filter", fmt.Sprintf("Got '%d' results", len(entries.Entries)))
		return
	}
	entry := entries.Entries[len(entries.Entries)-1]
	data.Id = types.StringValue(entry.Id)
	data.SpiffeId = &SpiffeIdModel{
		TrustDomain: types.StringValue(entry.SpiffeId.TrustDomain),
		Path:        types.StringValue(entry.SpiffeId.Path),
	}
	data.ParentId = &SpiffeIdModel{
		TrustDomain: types.StringValue(entry.ParentId.TrustDomain),
		Path:        types.StringValue(entry.ParentId.Path),
	}
	var selectors []SpiffeSelectorModel
	for _, sel := range entry.Selectors {
		selectors = append(selectors, SpiffeSelectorModel{
			Type:  types.StringValue(sel.Type),
			Value: types.StringValue(sel.Value),
		})
	}
	data.Selectors = &selectors

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	entryv1 "github.com/spiffe/spire-api-sdk/proto/spire/api/server/entry/v1"
	spireTypes "github.com/spiffe/spire-api-sdk/proto/spire/api/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SpireEntryResource{}
var _ resource.ResourceWithImportState = &SpireEntryResource{}

func NewSpireEntryResource() resource.Resource {
	return &SpireEntryResource{}
}

// SpireEntryResource defines the resource implementation.
type SpireEntryResource struct {
	client entryv1.EntryClient
}

type SpiffeIdModel struct {
	TrustDomain types.String `tfsdk:"trust_domain"`
	Path        types.String `tfsdk:"path"`
}

type SpiffeSelectorModel struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type SpireEntryResourceModel struct {
	Id        types.String           `tfsdk:"id"`
	SpiffeId  *SpiffeIdModel         `tfsdk:"spiffe_id"`
	ParentId  *SpiffeIdModel         `tfsdk:"parent_id"`
	Selectors *[]SpiffeSelectorModel `tfsdk:"selectors"`
}

func (r *SpireEntryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_entry"
}

func (r *SpireEntryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Register workloads with Spire Server Entry",

		Attributes: map[string]schema.Attribute{
			"parent_id": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"trust_domain": types.StringType,
					"path":         types.StringType,
				},
				MarkdownDescription: "The SPIFFE ID of this record's parent",
				Required:            true,
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
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Registration Entry ID of the record",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SpireEntryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	conn, err := grpc.Dial("unix:/tmp/spire-server/private/api.sock", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Grpc",
			err.Error(),
		)
	}
	//defer conn.Close()
	r.client = entryv1.NewEntryClient(conn)
}

func (r *SpireEntryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SpireEntryResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var selectors []*spireTypes.Selector
	for _, sel := range *data.Selectors {
		selectors = append(selectors, &spireTypes.Selector{
			Type:  sel.Type.ValueString(),
			Value: sel.Value.ValueString(),
		})
	}
	response, err := r.client.BatchCreateEntry(ctx, &entryv1.BatchCreateEntryRequest{Entries: []*spireTypes.Entry{
		{
			Id: data.Id.ValueString(),
			SpiffeId: &spireTypes.SPIFFEID{
				TrustDomain: data.SpiffeId.TrustDomain.ValueString(),
				Path:        data.SpiffeId.Path.ValueString(),
			},
			ParentId: &spireTypes.SPIFFEID{
				TrustDomain: data.ParentId.TrustDomain.ValueString(),
				Path:        data.ParentId.Path.ValueString(),
			},
			Selectors: selectors,
		},
	}})

	if err != nil {
		resp.Diagnostics.AddError("Failed to create entries", err.Error())
		return
	}

	for _, r := range response.Results {
		data.Id = types.StringValue(r.Entry.Id)
		data.SpiffeId = &SpiffeIdModel{
			TrustDomain: types.StringValue(r.Entry.SpiffeId.TrustDomain),
			Path:        types.StringValue(r.Entry.SpiffeId.Path),
		}
		data.ParentId = &SpiffeIdModel{
			TrustDomain: types.StringValue(r.Entry.ParentId.TrustDomain),
			Path:        types.StringValue(r.Entry.ParentId.Path),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created an entry resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SpireEntryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SpireEntryResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	entry, err := r.client.GetEntry(ctx, &entryv1.GetEntryRequest{Id: data.Id.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Failed to get entry", err.Error())
		return
	}

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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SpireEntryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SpireEntryResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.BatchUpdateEntry(ctx, &entryv1.BatchUpdateEntryRequest{
		Entries: []*spireTypes.Entry{
			{
				Id: data.Id.ValueString(),
				SpiffeId: &spireTypes.SPIFFEID{
					TrustDomain: data.SpiffeId.TrustDomain.ValueString(),
					Path:        data.SpiffeId.Path.ValueString(),
				},
				ParentId: &spireTypes.SPIFFEID{
					TrustDomain: data.ParentId.TrustDomain.ValueString(),
					Path:        data.ParentId.Path.ValueString(),
				},
				Selectors: []*spireTypes.Selector{
					{
						Type:  "unix",
						Value: "uid:501",
					},
				},
			},
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Failed to update entry", err.Error())
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SpireEntryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SpireEntryResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.BatchDeleteEntry(ctx, &entryv1.BatchDeleteEntryRequest{
		Ids: []string{
			data.Id.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to get entry", err.Error())
		return
	}
}

func (r *SpireEntryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

package provider

import (
	"context"

	permify_payload "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
	permify_grpc "github.com/Permify/permify-go/grpc"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &schemaResource{}
var _ resource.ResourceWithConfigure = &schemaResource{}
var _ resource.ResourceWithImportState = &schemaResource{}

type schemaResource struct {
	client *permify_grpc.Client
}

func NewSchemaResource() resource.Resource {
	return &schemaResource{}
}

func (r *schemaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*permify_grpc.Client)
	if !ok {
		tflog.Error(ctx, "Unable to prepare client")
		return
	}
	r.client = client
}

func (r *schemaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schema"
}

func (r *schemaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Schema resource",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the tenant the schema belongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "The complete schema for the tenant",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema_version": schema.StringAttribute{
				MarkdownDescription: "The version of the schema",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier",
				Computed:            true,
			},
		},
	}
}

func (r *schemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create schema resource")
	var data SchemaModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Schema.Write(ctx, &permify_payload.SchemaWriteRequest{
		TenantId: data.TenantID.ValueString(),
		Schema:   data.Schema.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Permify Schema", err.Error())
		return
	}

	data.ID = data.TenantID
	data.SchemaVersion = types.StringValue(result.SchemaVersion)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "Created Schema resource", map[string]any{"success": true})
}

func (r *schemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read item resource")
	// Get current state
	var state SchemaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schema, err := r.client.Schema.Read(ctx, &permify_payload.SchemaReadRequest{
		TenantId: state.TenantID.ValueString(),
		Metadata: &permify_payload.SchemaReadRequestMetadata{SchemaVersion: state.SchemaVersion.ValueString()},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading Permify Schema", err.Error())
	}

	// TODO: Let's make sure we only do this if we're sure the schema is not found, not on intermittent errors
	if schema == nil {
		resp.State.RemoveResource(ctx)
	}

	state = SchemaModel{
		ID:            types.StringValue(state.TenantID.ValueString()),
		TenantID:      types.StringValue(state.TenantID.ValueString()),
		Schema:        types.StringValue(state.Schema.ValueString()),
		SchemaVersion: types.StringValue(state.SchemaVersion.ValueString()),
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Finished reading Permify Schema resource", map[string]any{"success": true})
}

func (r *schemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Preparing to update schema resource")
	var data SchemaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Schema.Write(ctx, &permify_payload.SchemaWriteRequest{
		TenantId: data.TenantID.ValueString(),
		Schema:   data.Schema.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Permify Schema", err.Error())
		return
	}
	data.SchemaVersion = types.StringValue(result.SchemaVersion)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "Updated Permify Schema resource", map[string]any{"success": true})
}

func (r *schemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *schemaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}

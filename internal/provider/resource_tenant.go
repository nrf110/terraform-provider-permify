package provider

import (
	"context"
	"time"

	permify_payload "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
	permify_grpc "github.com/Permify/permify-go/grpc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &tenantResource{}
var _ resource.ResourceWithConfigure = &tenantResource{}
var _ resource.ResourceWithImportState = &tenantResource{}

type tenantResource struct {
	client *permify_grpc.Client
}

func NewTenantResource() resource.Resource {
	return &tenantResource{}
}

func (r *tenantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tenantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *tenantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Tenant resource",
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Created timestamp",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Friendly name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Unique identifier",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *tenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create tenant resource")
	var data TenantModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Tenancy.Create(ctx, &permify_payload.TenantCreateRequest{
		Id:   data.ID.ValueString(),
		Name: data.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Permify Tenant", err.Error())
		return
	}

	data.CreatedAt = types.StringValue(result.Tenant.CreatedAt.AsTime().Format(time.RFC3339))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "Created Tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read item resource")
	// Get current state
	var state TenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenant, err := findTenant(ctx, r.client, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading Permify Tenant", err.Error())
	}

	if tenant == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Map resp body to model
	state = TenantModel{
		ID:        types.StringValue(tenant.Id),
		Name:      types.StringValue(tenant.Name),
		CreatedAt: types.StringValue(tenant.CreatedAt.AsTime().Format(time.RFC3339)),
	}

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Finished reading Permify Tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Updating tenants is unsupported", "Cannot update a tenant, delete and recreate is required")
}

func (r *tenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state TenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Preparing to delete Permify Tenant resource", map[string]any{"id": state.ID.ValueString()})

	// delete item
	_, deleteErr := r.client.Tenancy.Delete(ctx, &permify_payload.TenantDeleteRequest{
		Id: state.ID.ValueString(),
	})
	// TODO: Handle error.  Currently there is a bug in Permify causing an unmarshalling error, even if the delete was successful.
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Delete Permify Tenant",
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// TODO: Either remove this once the bug in Permify is fixed, or add paging logic in case of more than 100 tenants.
	listResult, err := r.client.Tenancy.List(ctx, &permify_payload.TenantListRequest{
		PageSize: 100,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error listing Permify Tenants", err.Error())
		return
	}
	found := false
	for _, tenant := range listResult.Tenants {
		if tenant.Id == state.ID.ValueString() {
			found = true
			break
		}
	}
	if found && deleteErr != nil {
		resp.Diagnostics.AddError(
			"Error deleting Permify Tenant",
			deleteErr.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Permify Tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

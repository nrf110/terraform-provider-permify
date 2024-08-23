package provider

import (
	permify_payload "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
	"context"
	permify_grpc "github.com/Permify/permify-go/grpc"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"time"
)

var _ resource.Resource = &tenantResource{}
var _ resource.ResourceWithConfigure = &tenantResource{}
var _ resource.ResourceWithImportState = &tenantResource{}

type tenantModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
}

type tenantResource struct {
	client *permify_grpc.Client
}

func NewTenantResource() resource.Resource {
	return &tenantResource{}
}

func (r *tenantResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*permify_grpc.Client)
	if !ok {
		tflog.Error(ctx, "Unable to prepare client")
		return
	}
	r.client = client
}

func (r *tenantResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_tenant"
}

func (r *tenantResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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

func (r *tenantResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create tenant resource")
	var data tenantModel
	// Read Terraform plan data into the model
	diags := request.Plan.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Tenancy.Create(ctx, &permify_payload.TenantCreateRequest{
		Id:   data.ID.ValueString(),
		Name: data.Name.ValueString(),
	})
	if err != nil {
		response.Diagnostics.AddError("Failed to create Permify Tenant", err.Error())
		return
	}

	data.CreatedAt = types.StringValue(result.Tenant.CreatedAt.AsTime().Format(time.RFC3339))

	// Save data into Terraform state
	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)

	tflog.Debug(ctx, "Created Tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read item resource")
	// Get current state
	var state tenantModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	tenant, err := r.findTenant(ctx, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error reading Permify Tenant", err.Error())
	}

	if tenant == nil {
		response.State.RemoveResource(ctx)
	}

	// Map response body to model
	state = tenantModel{
		ID:        types.StringValue(tenant.Id),
		Name:      types.StringValue(tenant.Name),
		CreatedAt: types.StringValue(tenant.CreatedAt.AsTime().Format(time.RFC3339)),
	}

	// Set refreshed state
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Finished reading Permify Tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Updating tenants is unsupported", "Cannot update a tenant, delete and recreate is required")
}

func (r *tenantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	tflog.Debug(ctx, "Preparing to delete Permify Tenant resource")
	// Retrieve values from state
	var state tenantModel
	diags := request.State.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	// delete item
	_, err := r.client.Tenancy.Delete(ctx, &permify_payload.TenantDeleteRequest{
		Id: state.ID.ValueString(),
	})
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to Delete Permify Tenant",
			err.Error(),
		)
		return
	}
	tflog.Debug(ctx, "Deleted Permify Tenant resource", map[string]any{"success": true})
}

func (r *tenantResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *tenantResource) findTenant(ctx context.Context, id string) (*permify_payload.Tenant, error) {
	token := ""
	firstRun := true

	for token != "" || firstRun {
		result, err := r.client.Tenancy.List(ctx, &permify_payload.TenantListRequest{
			PageSize:        100,
			ContinuousToken: token,
		})
		if err != nil {
			return nil, err
		}
		for _, tenant := range result.Tenants {
			if tenant.Id == id {
				return tenant, nil
			}
		}
		firstRun = false
		token = result.ContinuousToken
	}
	return nil, nil
}

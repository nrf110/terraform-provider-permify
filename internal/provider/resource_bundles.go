package provider

import (
	"context"
	"sync"

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

var _ resource.Resource = &bundlesResource{}
var _ resource.ResourceWithConfigure = &bundlesResource{}
var _ resource.ResourceWithImportState = &bundlesResource{}

type bundlesResource struct {
	client *permify_grpc.Client
}

func NewBundlesResource() resource.Resource {
	return &bundlesResource{}
}

func (r *bundlesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bundlesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bundles"
}

func (r *bundlesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Bundles resource",
		Attributes: map[string]schema.Attribute{
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the tenant the bundles belong to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bundles": schema.ListNestedAttribute{
				MarkdownDescription: "The bundles for the tenant",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the bundle",
							Required:            true,
						},
						"arguments": schema.ListAttribute{
							MarkdownDescription: "The arguments of the bundle",
							Required:            true,
							ElementType:         types.StringType,
						},
						"operations": schema.ListNestedAttribute{
							MarkdownDescription: "The operations of the bundle",
							Required:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"relationships_write": schema.ListAttribute{
										MarkdownDescription: "Relationships that should be written by the bundle",
										Required:            true,
										ElementType:         types.StringType,
									},
									"relationships_delete": schema.ListAttribute{
										MarkdownDescription: "Relationships that should be deleted by the bundle",
										Required:            true,
										ElementType:         types.StringType,
									},
									"attributes_write": schema.ListAttribute{
										MarkdownDescription: "Attributes that should be written by the bundle",
										Required:            true,
										ElementType:         types.StringType,
									},
									"attributes_delete": schema.ListAttribute{
										MarkdownDescription: "Attributes that should be deleted by the bundle",
										Required:            true,
										ElementType:         types.StringType,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *bundlesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Debug(ctx, "Preparing to create bundles resource")
	var data BundlesModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Bundle.Write(ctx, data.ToWriteRequest())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Permify Bundles", err.Error())
		return
	}

	data.ID = data.TenantID

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "Created Bundles resource", map[string]any{"success": true})
}

func (r *bundlesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Debug(ctx, "Preparing to read bundles resource")
	var data BundlesModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		bundles []BundleModel
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	for _, bundle := range data.Bundles {
		wg.Go(func() {
			result, err := r.client.Bundle.Read(ctx, &permify_payload.BundleReadRequest{
				TenantId: data.TenantID.ValueString(),
				Name:     bundle.Name.ValueString(),
			})

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				resp.Diagnostics.AddError("Failed to read Permify Bundle", err.Error())
			} else {
				bundles = append(bundles, FromBundleReadResponse(result))
			}
		})
	}

	wg.Wait()

	data.Bundles = bundles
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "Read Bundles resource", map[string]any{"success": true})
}

func (r *bundlesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Debug(ctx, "Preparing to update bundles resource")
	var data BundlesModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Bundle.Write(ctx, data.ToWriteRequest())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Permify Bundles", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Debug(ctx, "Updated Bundles resource", map[string]any{"success": true})
}

func (r *bundlesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Debug(ctx, "Preparing to delete bundles resource")
	var data BundlesModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		mu sync.Mutex
		wg sync.WaitGroup
		// We need to keep the bundles that failed to delete so we can retry them
		bundles []BundleModel
	)
	for _, bundle := range data.Bundles {
		wg.Go(func() {
			_, err := r.client.Bundle.Delete(ctx, &permify_payload.BundleDeleteRequest{
				TenantId: data.TenantID.ValueString(),
				Name:     bundle.Name.ValueString(),
			})

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				resp.Diagnostics.AddError("Failed to delete Permify Bundle", err.Error())
				bundles = append(bundles, bundle)
			}
		})
	}
	wg.Wait()

	data.Bundles = bundles
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	tflog.Debug(ctx, "Deleted Bundles resource", map[string]any{"success": true})
}

func (r *bundlesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

package provider

import (
	permify_payload "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
	"context"
	permify_grpc "github.com/Permify/permify-go/grpc"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TenantModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func findTenant(ctx context.Context, client *permify_grpc.Client, id string) (*permify_payload.Tenant, error) {
	token := ""
	firstRun := true

	for token != "" || firstRun {
		result, err := client.Tenancy.List(ctx, &permify_payload.TenantListRequest{
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

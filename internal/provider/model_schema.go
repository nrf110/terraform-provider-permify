package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type SchemaModel struct {
	ID            types.String `tfsdk:"id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	Schema        types.String `tfsdk:"schema"`
	SchemaVersion types.String `tfsdk:"schema_version"`
}

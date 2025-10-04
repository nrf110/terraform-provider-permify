package provider

import (
	permify_payload "buf.build/gen/go/permifyco/permify/protocolbuffers/go/base/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type OperationModel struct {
	RelationshipsWrite  []types.String `tfsdk:"relationships_write"`
	RelationshipsDelete []types.String `tfsdk:"relationships_delete"`
	AttributesWrite     []types.String `tfsdk:"attributes_write"`
	AttributesDelete    []types.String `tfsdk:"attributes_delete"`
}

type BundleModel struct {
	Name       types.String     `tfsdk:"name"`
	Arguments  []types.String   `tfsdk:"arguments"`
	Operations []OperationModel `tfsdk:"operations"`
}

type BundlesModel struct {
	ID       types.String  `tfsdk:"id"`
	TenantID types.String  `tfsdk:"tenant_id"`
	Bundles  []BundleModel `tfsdk:"bundles"`
}

func (o OperationModel) toOperation() *permify_payload.Operation {
	relationshipsWrite := make([]string, len(o.RelationshipsWrite))
	for i, relationshipWrite := range o.RelationshipsWrite {
		relationshipsWrite[i] = relationshipWrite.ValueString()
	}
	relationshipsDelete := make([]string, len(o.RelationshipsDelete))
	for i, relationshipDelete := range o.RelationshipsDelete {
		relationshipsDelete[i] = relationshipDelete.ValueString()
	}
	attributesWrite := make([]string, len(o.AttributesWrite))
	for i, attributeWrite := range o.AttributesWrite {
		attributesWrite[i] = attributeWrite.ValueString()
	}
	attributesDelete := make([]string, len(o.AttributesDelete))
	for i, attributeDelete := range o.AttributesDelete {
		attributesDelete[i] = attributeDelete.ValueString()
	}
	return &permify_payload.Operation{
		RelationshipsWrite:  relationshipsWrite,
		RelationshipsDelete: relationshipsDelete,
		AttributesWrite:     attributesWrite,
		AttributesDelete:    attributesDelete,
	}
}

func (b BundleModel) toDataBundle() *permify_payload.DataBundle {
	arguments := make([]string, len(b.Arguments))
	for i, argument := range b.Arguments {
		arguments[i] = argument.ValueString()
	}
	operations := make([]*permify_payload.Operation, len(b.Operations))
	for i, operation := range b.Operations {
		operations[i] = operation.toOperation()
	}
	return &permify_payload.DataBundle{
		Name:       b.Name.ValueString(),
		Arguments:  arguments,
		Operations: operations,
	}
}

func fromOperation(response *permify_payload.Operation) OperationModel {
	relationshipsWrite := make([]types.String, len(response.RelationshipsWrite))
	for i, relationshipWrite := range response.RelationshipsWrite {
		relationshipsWrite[i] = types.StringValue(relationshipWrite)
	}
	relationshipsDelete := make([]types.String, len(response.RelationshipsDelete))
	for i, relationshipDelete := range response.RelationshipsDelete {
		relationshipsDelete[i] = types.StringValue(relationshipDelete)
	}
	attributesWrite := make([]types.String, len(response.AttributesWrite))
	for i, attributeWrite := range response.AttributesWrite {
		attributesWrite[i] = types.StringValue(attributeWrite)
	}
	attributesDelete := make([]types.String, len(response.AttributesDelete))
	for i, attributeDelete := range response.AttributesDelete {
		attributesDelete[i] = types.StringValue(attributeDelete)
	}
	return OperationModel{
		RelationshipsWrite:  relationshipsWrite,
		RelationshipsDelete: relationshipsDelete,
		AttributesWrite:     attributesWrite,
		AttributesDelete:    attributesDelete,
	}
}

func FromBundleReadResponse(response *permify_payload.BundleReadResponse) BundleModel {
	arguments := make([]types.String, len(response.Bundle.Arguments))
	for i, argument := range response.Bundle.Arguments {
		arguments[i] = types.StringValue(argument)
	}
	operations := make([]OperationModel, len(response.Bundle.Operations))
	for i, operation := range response.Bundle.Operations {
		operations[i] = fromOperation(operation)
	}
	return BundleModel{
		Name:       types.StringValue(response.Bundle.Name),
		Arguments:  arguments,
		Operations: operations,
	}
}

func (b BundlesModel) ToWriteRequest() *permify_payload.BundleWriteRequest {
	bundles := make([]*permify_payload.DataBundle, len(b.Bundles))
	for i, bundle := range b.Bundles {
		bundles[i] = bundle.toDataBundle()
	}
	return &permify_payload.BundleWriteRequest{
		TenantId: b.TenantID.ValueString(),
		Bundles:  bundles,
	}
}

func (b BundlesModel) Removed(other BundlesModel) []BundleModel {
	inOther := make(map[string]bool)
	for _, bundle := range other.Bundles {
		inOther[bundle.Name.ValueString()] = true
	}
	removed := make([]BundleModel, 0)

	for _, bundle := range b.Bundles {
		_, found := inOther[bundle.Name.ValueString()]
		if !found {
			removed = append(removed, bundle)
		}
	}
	return removed
}

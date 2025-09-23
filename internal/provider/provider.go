package provider

import (
	"context"

	permify_grpc "github.com/Permify/permify-go/grpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var _ provider.Provider = &permifyProvider{}
var _ provider.ProviderWithFunctions = &permifyProvider{}

type permifyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type PermifyProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	Token    types.String `tfsdk:"token"`
}

func (p *permifyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "permify"
	resp.Version = p.version
}

func (p *permifyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "gRPC endpoint for the Permify API.  Defaults to `localhost:3478`.",
				Required:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Bearer Token to authenticated to the Permify API.  Can be an OAuth2 token a Pre-Shared Key.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *permifyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PermifyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.IsNull() || data.Endpoint.ValueString() == "" {
		data.Endpoint = types.StringValue("localhost:3478")
		return
	}

	client, err := permify_grpc.NewClient(
		permify_grpc.Config{
			Endpoint: data.Endpoint.ValueString(),
		},
		getOptions(data)...,
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to initialize Permify client", err.Error())
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *permifyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSchemaResource,
		NewTenantResource,
	}
}

func (p *permifyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTenantDataSource,
	}
}

func (p *permifyProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &permifyProvider{
			version: version,
		}
	}
}

func getOptions(data PermifyProviderModel) []grpc.DialOption {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if !data.Token.IsNull() && data.Token.ValueString() != "" {
		options = append(options, grpc.WithUnaryInterceptor(authInterceptor(data.Token.ValueString())))
	}

	return options
}

func authInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(metadata.AppendToOutgoingContext(ctx, "authorization", token), method, req, reply, cc, opts...)
	}
}

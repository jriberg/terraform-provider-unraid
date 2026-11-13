// Package provider implements the Unraid Terraform/OpenTofu provider.
package provider

import (
	"context"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/jriberg/terraform-provider-unraid/internal/client"
)

var _ provider.Provider = &unraidProvider{}

type unraidProvider struct {
	version string
}

type providerModel struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	APIKey             types.String `tfsdk:"api_key"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

// New returns a provider factory.
func New(version string) func() provider.Provider {
	return func() provider.Provider { return &unraidProvider{version: version} }
}

func (p *unraidProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unraid"
	resp.Version = p.version
}

func (p *unraidProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage an Unraid server through the Unraid GraphQL API.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Full URL of the GraphQL endpoint, e.g. `https://tower.local/graphql`. Can also be set with `UNRAID_ENDPOINT`.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Unraid API key. Can also be set with `UNRAID_API_KEY`.",
				Optional:            true,
				Sensitive:           true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification (for self-signed certificates). Defaults to `false`.",
				Optional:            true,
			},
		},
	}
}

// resolveConfig applies env fallbacks and validates; it fails closed on missing or invalid values.
func resolveConfig(m providerModel) (endpoint, apiKey string, diags diag.Diagnostics) {
	endpoint, apiKey = m.Endpoint.ValueString(), m.APIKey.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("UNRAID_ENDPOINT")
	}
	if apiKey == "" {
		apiKey = os.Getenv("UNRAID_API_KEY")
	}
	if endpoint == "" {
		diags.AddAttributeError(path.Root("endpoint"), "Missing endpoint", "Set `endpoint` or UNRAID_ENDPOINT.")
	} else if u, err := url.Parse(endpoint); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		diags.AddAttributeError(path.Root("endpoint"), "Invalid endpoint", "Must be an absolute http(s) URL, e.g. https://tower.local/graphql.")
	} else if u.Scheme == "http" {
		diags.AddAttributeWarning(path.Root("endpoint"), "Unencrypted endpoint", "The API key is sent in clear text over http. Use https if possible.")
	}
	if apiKey == "" {
		diags.AddAttributeError(path.Root("api_key"), "Missing API key", "Set `api_key` or UNRAID_API_KEY.")
	}
	return endpoint, apiKey, diags
}

func (p *unraidProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var m providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if m.Endpoint.IsUnknown() || m.APIKey.IsUnknown() {
		resp.Diagnostics.AddError("Unknown provider configuration", "endpoint and api_key must be known at plan time.")
		return
	}
	endpoint, apiKey, diags := resolveConfig(m)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	c, err := client.New(endpoint, apiKey, m.InsecureSkipVerify.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Unraid client", err.Error())
		return
	}
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *unraidProvider) Resources(context.Context) []func() resource.Resource {
	return nil
}

func (p *unraidProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newVarsDataSource,
		newServerDataSource,
		newSystemTimeDataSource,
		newUpsConfigurationDataSource,
		newSettingsDataSource,
		newSharesDataSource,
		newArrayDataSource,
		newDisksDataSource,
		newDockerContainersDataSource,
		newVmsDataSource,
		newAPIKeysDataSource,
		newRCloneRemotesDataSource,
	}
}

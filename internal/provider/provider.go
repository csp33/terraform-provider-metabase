// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure MetabaseProvider satisfies various provider interfaces.
var _ provider.Provider = &MetabaseProvider{}
var _ provider.ProviderWithFunctions = &MetabaseProvider{}
var _ provider.ProviderWithEphemeralResources = &MetabaseProvider{}

// MetabaseProvider defines the provider implementation.
type MetabaseProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type MetabaseClient struct {
	Host   string
	ApiKey string
	Client *http.Client
}

// MetabaseProviderModel describes the provider data model.
type MetabaseProviderModel struct {
	Host types.String `tfsdk:"host"`
	User types.String `tfsdk:"api_key"`
}

func (p *MetabaseProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "metabase"
	resp.Version = p.version
}

func (p *MetabaseProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Metabase API host URL",
				Required:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Metabase API Key",
				Required:            true,
			},
		},
	}
}

func (p *MetabaseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data MetabaseProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	metabaseClient := &MetabaseClient{
		Host:   data.Host.ValueString(),
		ApiKey: data.User.ValueString(),
		Client: http.DefaultClient,
	}

	resp.DataSourceData = metabaseClient
	resp.ResourceData = metabaseClient
}

func (p *MetabaseProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPermissionGroup,
	}
}

func (p *MetabaseProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *MetabaseProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *MetabaseProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MetabaseProvider{
			version: version,
		}
	}
}

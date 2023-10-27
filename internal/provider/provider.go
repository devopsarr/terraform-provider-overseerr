package provider

import (
	"context"
	"os"

	"github.com/devopsarr/overseerr-go/overseerr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// needed for tf debug mode
// var stderr = os.Stderr

// Ensure provider defined types fully satisfy framework interfaces.
var _ provider.Provider = &OverseerrProvider{}

// ScaffoldingProvider defines the provider implementation.
type OverseerrProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Overseerr describes the provider data model.
type Overseerr struct {
	APIKey types.String `tfsdk:"api_key"`
	URL    types.String `tfsdk:"url"`
}

func (p *OverseerrProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "overseerr"
	resp.Version = p.version
}

func (p *OverseerrProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Overseerr provider is used to interact with any [Overseerr](https://overseerr.dev/) installation.\nYou must configure the provider with the proper [credentials](#api_key) before you can use it.\nUse the left navigation to read about the available resources.\n\nFor more information about Overseerr and its resources, as well as configuration guides and hints, visit the [Servarr wiki](https://wiki.servarr.com/en/overseerr).",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key for Overseerr authentication. Can be specified via the `OVERSEERR_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Full Overseerr URL with protocol and port (e.g. `https://test.overseerr.tv:5055`). You should **NOT** supply any path (`/api`), the SDK will use the appropriate paths. Can be specified via the `OVERSEERR_URL` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *OverseerrProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data Overseerr

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// User must provide URL to the provider
	if data.URL.IsUnknown() {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as url",
		)

		return
	}

	var url string
	if data.URL.IsNull() {
		url = os.Getenv("OVERSEERR_URL")
	} else {
		url = data.URL.ValueString()
	}

	if url == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find URL",
			"URL cannot be an empty string",
		)

		return
	}

	// User must provide API key to the provider
	if data.APIKey.IsUnknown() {
		// Cannot connect to client with an unknown value
		resp.Diagnostics.AddWarning(
			"Unable to create client",
			"Cannot use unknown value as api_key",
		)

		return
	}

	var key string
	if data.APIKey.IsNull() {
		key = os.Getenv("OVERSEERR_API_KEY")
	} else {
		key = data.APIKey.ValueString()
	}

	if key == "" {
		// Error vs warning - empty value must stop execution
		resp.Diagnostics.AddError(
			"Unable to find API key",
			"API key cannot be an empty string",
		)

		return
	}

	// Configuring client. API Key management could be changed once new options avail in sdk.
	config := overseerr.NewConfiguration()
	config.AddDefaultHeader("X-Api-Key", key)
	config.Servers[0].URL = url
	client := overseerr.NewAPIClient(config)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OverseerrProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSettingsResource,
	}
}

func (p *OverseerrProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

// New returns the provider with a specific version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OverseerrProvider{
			version: version,
		}
	}
}

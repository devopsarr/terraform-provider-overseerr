package provider

import (
	"context"
	"fmt"

	"github.com/devopsarr/overseerr-go/overseerr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const settingsResourceName = "settings"

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &SettingsResource{}
	_ resource.ResourceWithImportState = &SettingsResource{}
)

func NewSettingsResource() resource.Resource {
	return &SettingsResource{}
}

// SettingsResource defines the settings implementation.
type SettingsResource struct {
	client *overseerr.APIClient
}

// Settings describes the settings data model.
type Settings struct {
	DefaultPermissions     types.Float64 `tfsdk:"default_permissions"`
	AppLanguage            types.String  `tfsdk:"app_language"`
	ApplicationTitle       types.String  `tfsdk:"application_title"`
	ApplicationURL         types.String  `tfsdk:"application_url"`
	TrustProxy             types.Bool    `tfsdk:"trust_proxy"`
	CsrfProtection         types.Bool    `tfsdk:"csrf_protection"`
	HideAvailable          types.Bool    `tfsdk:"hide_available"`
	PartialRequestsEnabled types.Bool    `tfsdk:"partial_requests_enabled"`
	LocalLogin             types.Bool    `tfsdk:"local_login"`
	NewPlexLogin           types.Bool    `tfsdk:"new_plex_login"`
}

func (r *SettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + settingsResourceName
}

func (r *SettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "<!-- subcategory:Settings -->Settings resource.\nFor more information refer to [Settings](https://docs.overseerr.dev/using-overseerr/settings#general) documentation.",
		Attributes: map[string]schema.Attribute{
			"new_plex_login": schema.BoolAttribute{
				MarkdownDescription: "New Plex login.",
				Computed:            true,
			},
			"local_login": schema.BoolAttribute{
				MarkdownDescription: "Local login.",
				Computed:            true,
			},
			"partial_requests_enabled": schema.BoolAttribute{
				MarkdownDescription: "Partial requests enabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"hide_available": schema.BoolAttribute{
				MarkdownDescription: "Hide available.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"csrf_protection": schema.BoolAttribute{
				MarkdownDescription: "CSRF protection.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"trust_proxy": schema.BoolAttribute{
				MarkdownDescription: "Trust policy.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"default_permissions": schema.Float64Attribute{
				MarkdownDescription: "Default permissions.",
				Computed:            true,
			},
			"application_url": schema.StringAttribute{
				MarkdownDescription: "Application URL.",
				Optional:            true,
				Computed:            true,
			},
			"application_title": schema.StringAttribute{
				MarkdownDescription: "Instance name.",
				Required:            true,
			},
			"app_language": schema.StringAttribute{
				MarkdownDescription: "Instance name.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("en"),
			},
		},
	}
}

func (r *SettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*overseerr.APIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sonarr.APIClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *SettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var settings *Settings

	resp.Diagnostics.Append(req.Plan.Get(ctx, &settings)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build Create resource
	request := settings.read()

	// Create new Settings
	response, _, err := r.client.SettingsAPI.CreateMain(ctx).MainSettings(*request).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create config, got error: %s", err))

		return
	}

	tflog.Trace(ctx, "created "+settingsResourceName)
	// Generate resource state struct
	settings.write(ctx, response)
	resp.Diagnostics.Append(resp.State.Set(ctx, &settings)...)
}

func (r *SettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var settings *Settings

	resp.Diagnostics.Append(req.State.Get(ctx, &settings)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get settings current value
	response, _, err := r.client.SettingsAPI.GetMain(ctx).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read config, got error: %s", err))

		return
	}

	tflog.Trace(ctx, "read "+settingsResourceName)
	// Map response body to resource schema attribute
	settings.write(ctx, response)
	resp.Diagnostics.Append(resp.State.Set(ctx, &settings)...)
}

func (r *SettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var settings *Settings

	resp.Diagnostics.Append(req.Plan.Get(ctx, &settings)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build Update resource
	request := settings.read()

	// Update Settings
	response, _, err := r.client.SettingsAPI.CreateMain(ctx).MainSettings(*request).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update config, got error: %s", err))

		return
	}

	tflog.Trace(ctx, "updated "+settingsResourceName)
	// Generate resource state struct
	settings.write(ctx, response)
	resp.Diagnostics.Append(resp.State.Set(ctx, &settings)...)
}

func (r *SettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Settings cannot be really deleted just removing configuration
	tflog.Trace(ctx, "decoupled "+settingsResourceName)
	resp.State.RemoveResource(ctx)
}

func (r *SettingsResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, _ *resource.ImportStateResponse) {
	tflog.Trace(ctx, "imported "+settingsResourceName)
}

func (s *Settings) write(_ context.Context, settings *overseerr.MainSettings) {
	s.AppLanguage = types.StringValue(settings.GetAppLanguage())
	s.ApplicationURL = types.StringValue(settings.GetApplicationUrl())
	s.ApplicationTitle = types.StringValue(settings.GetApplicationTitle())
	s.DefaultPermissions = types.Float64Value(float64(settings.GetDefaultPermissions()))
	s.TrustProxy = types.BoolValue(settings.GetTrustProxy())
	s.CsrfProtection = types.BoolValue(settings.GetCsrfProtection())
	s.HideAvailable = types.BoolValue(settings.GetHideAvailable())
	s.PartialRequestsEnabled = types.BoolValue(settings.GetPartialRequestsEnabled())
	s.LocalLogin = types.BoolValue(settings.GetLocalLogin())
	s.NewPlexLogin = types.BoolValue(settings.GetNewPlexLogin())
}

func (s *Settings) read() *overseerr.MainSettings {
	settings := overseerr.NewMainSettings()
	settings.SetAppLanguage(s.AppLanguage.ValueString())
	settings.SetApplicationUrl(s.ApplicationURL.ValueString())
	settings.SetApplicationTitle(s.ApplicationTitle.ValueString())
	settings.SetDefaultPermissions(float32(s.DefaultPermissions.ValueFloat64()))
	settings.SetTrustProxy(s.TrustProxy.ValueBool())
	settings.SetCsrfProtection(s.CsrfProtection.ValueBool())
	settings.SetHideAvailable(s.HideAvailable.ValueBool())
	settings.SetPartialRequestsEnabled(s.PartialRequestsEnabled.ValueBool())
	settings.SetLocalLogin(s.LocalLogin.ValueBool())
	settings.SetNewPlexLogin(s.NewPlexLogin.ValueBool())

	return settings
}

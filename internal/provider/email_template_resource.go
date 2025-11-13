// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/framebassman/infobip-api-go-client/v3/pkg/infobip"
	"github.com/framebassman/infobip-api-go-client/v3/pkg/infobip/api"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &EmailTemplateResource{}
var _ resource.ResourceWithImportState = &EmailTemplateResource{}

func NewEmailTemplateResource() resource.Resource {
	return &EmailTemplateResource{}
}

// EmailTemplateResource defines the resource implementation.
type EmailTemplateResource struct {
	infobipClient *api.APIClient
	apiKey        string
}

// EmailTemplateResourceModel describes the resource data model.
type EmailTemplateResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	From            types.String `tfsdk:"from"`
	ReplyTo         types.String `tfsdk:"reply_to"`
	Subject         types.String `tfsdk:"subject"`
	Preheader       types.String `tfsdk:"preheader"`
	Html            types.String `tfsdk:"html"`
	IsHtmlEditable  types.Bool   `tfsdk:"is_html_editable"`
	LandingPage     types.String `tfsdk:"landing_page"`
	ImagePreviewUrl types.String `tfsdk:"image_preview_url"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *EmailTemplateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_template"
}

func (r *EmailTemplateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Infobip Email Template resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the email template.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the email template.",
				Required:    true,
			},
			"from": schema.StringAttribute{
				Description: "Sender email address used in the template.",
				Required:    true,
			},
			"reply_to": schema.StringAttribute{
				Description: "Reply-to email address for the template.",
				Optional:    true,
			},
			"subject": schema.StringAttribute{
				Description: "Subject line of the email template.",
				Required:    true,
			},
			"preheader": schema.StringAttribute{
				Description: "Preheader text shown in email previews (optional).",
				Optional:    true,
			},
			"html": schema.StringAttribute{
				Description: "HTML content of the email template.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					htmlWhitespaceInsensitiveModifier{},
				},
			},
			"is_html_editable": schema.BoolAttribute{
				Description: "Indicates whether the HTML content can be edited in Infobip UI.",
				Computed:    true,
			},
			"landing_page": schema.StringAttribute{
				Description: "Associated landing page ID, if any.",
				Optional:    true,
				Computed:    true,
			},
			"image_preview_url": schema.StringAttribute{
				Description: "URL of the email template’s image preview.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the email template was created (RFC3339 format).",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the email template was last updated (RFC3339 format).",
				Computed:    true,
			},
		},
	}
}

func (r *EmailTemplateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Infobip client")
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// Expect provider to pass *providerClient
	pd, ok := req.ProviderData.(*providerClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *providerClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.infobipClient = pd.client
	r.apiKey = pd.apiKey
	tflog.Info(ctx, "Finish Infobip client configuration")
}

func (r *EmailTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan EmailTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make API call to create resource
	auth := context.WithValue(
		context.Background(),
		infobip.ContextAPIKeys,
		map[string]infobip.APIKey{"APIKeyHeader": {Key: r.apiKey, Prefix: "App"}},
	)
	emailTemplate, httpResponse, err := r.infobipClient.
		EmailAPI.
		CreateEmailTemplate(auth).
		Name(plan.Name.ValueString()).
		From(plan.From.ValueString()).
		ReplyTo(plan.ReplyTo.ValueString()).
		Subject(plan.Subject.ValueString()).
		Preheader(plan.Preheader.ValueString()).
		Html(plan.Html.ValueString()).
		LandingPage(plan.LandingPage.ValueString()).
		Execute()

	tflog.Info(ctx, fmt.Sprintf("HTTP Response Details: %+v\n", httpResponse))
	// Check for errors
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Email Template",
			"An error was encountered while creating the email template: "+err.Error(),
		)
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(fmt.Sprintf("%d", emailTemplate.ID))
	plan.Name = types.StringValue(emailTemplate.Name)
	plan.From = types.StringValue(emailTemplate.From)
	plan.ReplyTo = types.StringValue(emailTemplate.ReplyTo)
	plan.Subject = types.StringValue(emailTemplate.Subject)
	plan.Preheader = types.StringValue(emailTemplate.Preheader)
	// Format stored HTML as well
	plan.Html = types.StringValue(normalizeHTML(emailTemplate.HTML))
	plan.IsHtmlEditable = types.BoolValue(emailTemplate.IsHTMLEditable)
	plan.LandingPage = types.StringValue(emailTemplate.LandingPageID)
	plan.ImagePreviewUrl = types.StringValue(emailTemplate.ImagePreviewURL)
	plan.CreatedAt = types.StringValue(time.Now().Format(time.RFC850))
	plan.UpdatedAt = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *EmailTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state EmailTemplateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	auth := context.WithValue(
		context.Background(),
		infobip.ContextAPIKeys,
		map[string]infobip.APIKey{"APIKeyHeader": {Key: r.apiKey, Prefix: "App"}},
	)

	var idInt int64
	fmt.Sscanf(state.ID.ValueString(), "%d", &idInt)
	emailTemplate, httpResponse, err := r.infobipClient.
		EmailAPI.
		GetEmailTemplate(auth).
		ID(idInt).
		Execute()

	tflog.Info(ctx, fmt.Sprintf("HTTP Response Details: %+v\n", httpResponse))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HashiCups Order",
			"Could not read HashiCups order ID "+state.ID.String()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(fmt.Sprintf("%d", emailTemplate.ID))
	state.Name = types.StringValue(emailTemplate.Name)
	state.From = types.StringValue(emailTemplate.From)
	state.ReplyTo = types.StringValue(emailTemplate.ReplyTo)
	state.Subject = types.StringValue(emailTemplate.Subject)
	state.Preheader = types.StringValue(emailTemplate.Preheader)
	// Format HTML when mapping back to state
	state.Html = types.StringValue(normalizeHTML(emailTemplate.HTML))
	state.IsHtmlEditable = types.BoolValue(emailTemplate.IsHTMLEditable)
	state.LandingPage = types.StringValue(emailTemplate.LandingPageID)
	state.ImagePreviewUrl = types.StringValue(emailTemplate.ImagePreviewURL)
	state.CreatedAt = types.StringValue(emailTemplate.CreatedAt)
	state.UpdatedAt = types.StringValue(emailTemplate.UpdatedAt)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *EmailTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read plan and prior state
	var plan EmailTemplateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state EmailTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare auth context
	auth := context.WithValue(
		context.Background(),
		infobip.ContextAPIKeys,
		map[string]infobip.APIKey{"APIKeyHeader": {Key: r.apiKey, Prefix: "App"}},
	)

	// Call update API
	var idInt int64
	fmt.Sscanf(state.ID.ValueString(), "%d", &idInt)
	emailTemplate, httpResponse, err := r.infobipClient.
		EmailAPI.
		UpdateEmailTemplate(auth).
		ID(idInt).
		Name(plan.Name.ValueString()).
		From(plan.From.ValueString()).
		ReplyTo(plan.ReplyTo.ValueString()).
		Subject(plan.Subject.ValueString()).
		Preheader(plan.Preheader.ValueString()).
		Html(plan.Html.ValueString()).
		LandingPage(plan.LandingPage.ValueString()).
		Execute()

	tflog.Info(ctx, fmt.Sprintf("HTTP Response Details: %+v\n", httpResponse))

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Email Template",
			"An error was encountered while updating the email template: "+err.Error(),
		)
		return
	}

	// Map response back to state (preserve created_at if not returned)
	plan.ID = types.StringValue(fmt.Sprintf("%d", emailTemplate.ID))
	plan.Name = types.StringValue(emailTemplate.Name)
	plan.From = types.StringValue(emailTemplate.From)
	plan.ReplyTo = types.StringValue(emailTemplate.ReplyTo)
	plan.Subject = types.StringValue(emailTemplate.Subject)
	plan.Preheader = types.StringValue(emailTemplate.Preheader)
	plan.Html = types.StringValue(normalizeHTML(emailTemplate.HTML))
	plan.IsHtmlEditable = types.BoolValue(emailTemplate.IsHTMLEditable)
	plan.LandingPage = types.StringValue(emailTemplate.LandingPageID)
	plan.ImagePreviewUrl = types.StringValue(emailTemplate.ImagePreviewURL)

	// Preserve created_at from prior state if API doesn't return it
	if state.CreatedAt.ValueString() != "" {
		plan.CreatedAt = state.CreatedAt
	} else {
		plan.CreatedAt = types.StringValue(time.Now().Format(time.RFC850))
	}
	plan.UpdatedAt = types.StringValue(time.Now().Format(time.RFC850))

	// Set updated state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *EmailTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EmailTemplateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare auth context
	auth := context.WithValue(
		context.Background(),
		infobip.ContextAPIKeys,
		map[string]infobip.APIKey{"APIKeyHeader": {Key: r.apiKey, Prefix: "App"}},
	)

	// Call delete API
	var idInt int64
	fmt.Sscanf(data.ID.ValueString(), "%d", &idInt)
	httpResponse, err := r.infobipClient.
		EmailAPI.
		RemoveEmailTemplate(auth).
		ID(idInt).
		Execute()

	tflog.Info(ctx, fmt.Sprintf("HTTP Response Details: %+v\n", httpResponse))

	if err != nil {
		// If resource is already gone, treat as success and remove state.
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			tflog.Info(ctx, "Email template already deleted; removing from state", map[string]any{"id": data.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Email Template",
			"An error was encountered while deleting the email template: "+err.Error(),
		)
		return
	}

	// Remove resource from state to indicate deletion succeeded.
	resp.State.RemoveResource(ctx)
}

func (r *EmailTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func normalizeHTML(raw string) string {
	// Normalize line endings, trim edges, collapse multiple spaces
	s := strings.ReplaceAll(raw, "\r\n", "\n")
	s = strings.TrimSpace(s)
	s = strings.Join(strings.Fields(s), " ")
	re := regexp.MustCompile(`>[\s]*<`)
	s = re.ReplaceAllString(s, "><")
	return s
}

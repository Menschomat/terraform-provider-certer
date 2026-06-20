package provider

import (
	"context"
	"fmt"

	"github.com/menscho/terraform-provider-certcentral/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &APIKeyResource{}
var _ resource.ResourceWithConfigure = &APIKeyResource{}

type APIKeyResource struct {
	client *client.Client
}

type APIKeyResourceModel struct {
	ID             types.String   `tfsdk:"id"`
	Description    types.String   `tfsdk:"description"`
	CleartextToken types.String   `tfsdk:"cleartext_token"`
	AllowedDomains []types.String `tfsdk:"allowed_domains"`
	AllowedTeams   []types.String `tfsdk:"allowed_teams"`
	Admin          types.Bool     `tfsdk:"admin"`
}

func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{}
}

func (r *APIKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *APIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an API key token configuration in cert-central.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique UUID identifier of the API key configuration.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the API key configuration.",
				Optional:            true,
			},
			"cleartext_token": schema.StringAttribute{
				MarkdownDescription: "The generated plaintext token (sensitive). Only available on creation.",
				Computed:            true,
				Sensitive:           true,
			},
			"allowed_domains": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The list of domains this standard token is authorized to fetch certs for.",
				Optional:            true,
			},
			"allowed_teams": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The list of team UUIDs this token is scoped to (for configuration management if admin=true, or certificate retrieval if admin=false).",
				Optional:            true,
			},
			"admin": schema.BoolAttribute{
				MarkdownDescription: "Indicates if this is an administrative token with access to control plane APIs.",
				Required:            true,
			},
		},
	}
}

func (r *APIKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}

	r.client = c
}

func (r *APIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	allowed := []string{}
	for _, ad := range data.AllowedDomains {
		allowed = append(allowed, ad.ValueString())
	}

	allowedTeams := []string{}
	for _, at := range data.AllowedTeams {
		allowedTeams = append(allowedTeams, at.ValueString())
	}

	key := client.APIKeyConfig{
		Description:    data.Description.ValueString(),
		AllowedDomains: allowed,
		AllowedTeams:   allowedTeams,
		Admin:          data.Admin.ValueBool(),
	}

	createdKey, err := r.client.CreateAPIKey(ctx, key)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create API key: %s", err))
		return
	}

	data.ID = types.StringValue(createdKey.ID)
	data.CleartextToken = types.StringValue(createdKey.CleartextToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keys, err := r.client.GetAPIKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read API keys: %s", err))
		return
	}

	found := false
	for _, k := range keys {
		if k.ID == data.ID.ValueString() {
			found = true
			data.Description = types.StringValue(k.Description)
			data.Admin = types.BoolValue(k.Admin)
			allowedVal := []types.String{}
			for _, ad := range k.AllowedDomains {
				allowedVal = append(allowedVal, types.StringValue(ad))
			}
			data.AllowedDomains = allowedVal
			allowedTeamsVal := []types.String{}
			for _, at := range k.AllowedTeams {
				allowedTeamsVal = append(allowedTeamsVal, types.StringValue(at))
			}
			data.AllowedTeams = allowedTeamsVal
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data APIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the cleartext_token from state
	data.CleartextToken = state.CleartextToken

	allowed := []string{}
	for _, ad := range data.AllowedDomains {
		allowed = append(allowed, ad.ValueString())
	}

	allowedTeams := []string{}
	for _, at := range data.AllowedTeams {
		allowedTeams = append(allowedTeams, at.ValueString())
	}

	key := client.APIKeyConfig{
		ID:             data.ID.ValueString(),
		Description:    data.Description.ValueString(),
		AllowedDomains: allowed,
		AllowedTeams:   allowedTeams,
		Admin:          data.Admin.ValueBool(),
	}

	err := r.client.UpdateAPIKey(ctx, key)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update API key: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *APIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data APIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAPIKey(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete API key: %s", err))
		return
	}
}

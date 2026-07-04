package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/terraform-coop/terraform-provider-foreman/generated"
)

// autosign entries have no numeric primary key of their own (the entity's
// "id" is the pattern string itself), live under a specific smart proxy,
// and Foreman exposes no update or show endpoint for them - none of which
// fits the generic apidoc-driven pipeline's assumptions, so this resource
// is hand-written. See tools/gen/overrides.yaml's skip_resources entry.

var _ resource.Resource = &autosignResource{}
var _ resource.ResourceWithImportState = &autosignResource{}

func NewForemanAutosignResource() resource.Resource {
	return &autosignResource{}
}

type autosignResource struct {
	client *generated.ForemanClient
}

type autosignResourceModel struct {
	ID           types.String `tfsdk:"id"`
	SmartProxyID types.Int64  `tfsdk:"smart_proxy_id"`
}

func (r *autosignResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_autosign"
}

func (r *autosignResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The hostname or wildcard pattern to allow autosigning for (e.g. \"host.example.com\" or \"*.example.com\"). Foreman does not assign autosign entries a separate numeric id, so this value is both the input and the id.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"smart_proxy_id": schema.Int64Attribute{
			Description: "ID of the smart proxy this autosign entry applies to.",
			Required:    true,
			PlanModifiers: []planmodifier.Int64{
				int64planmodifier.RequiresReplace(),
			},
		},
	}}
}

func (r *autosignResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*generated.ForemanClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *generated.ForemanClient")
		return
	}
	r.client = client
}

func (r *autosignResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan autosignResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smartProxyID := int(plan.SmartProxyID.ValueInt64())
	result, err := r.client.CreateForemanAutosign(ctx, smartProxyID, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create autosign entry, got error: %s", err))
		return
	}

	plan.ID = types.StringValue(result.ID)

	tflog.Trace(ctx, "created autosign entry", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *autosignResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state autosignResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smartProxyID := int(state.SmartProxyID.ValueInt64())
	result, err := r.client.ReadForemanAutosign(ctx, smartProxyID, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read autosign entry, got error: %s", err))
		return
	}
	if result == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(result.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is unreachable: both attributes require replacement, so Terraform
// core always does delete+create instead of calling Update. Foreman exposes
// no update endpoint for autosign entries anyway.
func (r *autosignResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Not Supported", "Update is not supported for this resource")
}

func (r *autosignResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state autosignResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	smartProxyID := int(state.SmartProxyID.ValueInt64())
	err := r.client.DeleteForemanAutosign(ctx, smartProxyID, state.ID.ValueString())
	if err != nil && !generated.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete autosign entry, got error: %s", err))
		return
	}
}

// ImportState expects "<smart_proxy_id>/<pattern>", since both fields are
// required to read or delete an autosign entry and neither can be derived
// from the other.
func (r *autosignResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier of the form \"smart_proxy_id/pattern\", got: %s", req.ID),
		)
		return
	}
	smartProxyID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Import Identifier", fmt.Sprintf("smart_proxy_id must be numeric, got: %s", parts[0]))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("smart_proxy_id"), smartProxyID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

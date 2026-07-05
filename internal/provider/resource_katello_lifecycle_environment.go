package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &katelloLifecycleEnvironmentResource{}
	_ resource.ResourceWithImportState = &katelloLifecycleEnvironmentResource{}
)

func NewKatelloLifecycleEnvironmentResource() resource.Resource {
	return &katelloLifecycleEnvironmentResource{}
}

type katelloLifecycleEnvironmentResource struct {
	client *goforeman.Client
}

type katelloLifecycleEnvironmentResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Label          types.String `tfsdk:"label"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Library        types.Bool   `tfsdk:"library"`
	PriorID        types.Int64  `tfsdk:"prior_id"`
	SuccessorID    types.Int64  `tfsdk:"successor_id"`
}

func (r *katelloLifecycleEnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_katello_lifecycle_environment"
}

func (r *katelloLifecycleEnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"label": schema.StringAttribute{
				Optional: true,
			},
			"organization_id": schema.Int64Attribute{
				Optional: true,
			},
			"library": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"prior_id": schema.Int64Attribute{
				Optional: true,
			},
			"successor_id": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *katelloLifecycleEnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*goforeman.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *goforeman.Client")
		return
	}
	r.client = client
}

func (r *katelloLifecycleEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan katelloLifecycleEnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &goforeman.KatelloLifecycleEnvironmentRequest{
		Name:           plan.Name.ValueString(),
		Description:    plan.Description.ValueString(),
		Label:          plan.Label.ValueString(),
		OrganizationID: int(plan.OrganizationID.ValueInt64()),
		PriorID:        int(plan.PriorID.ValueInt64()),
	}

	result, err := r.client.CreateKatelloLifecycleEnvironment(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create katello lifecycle environment, got error: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	plan.Name = types.StringValue(result.Name)
	plan.Description = types.StringValue(result.Description)
	plan.Label = types.StringValue(result.Label)
	plan.OrganizationID = types.Int64Value(int64(result.OrganizationID))
	plan.Library = types.BoolValue(result.Library)
	plan.PriorID = types.Int64Value(int64(result.PriorID))
	plan.SuccessorID = types.Int64Value(int64(result.SuccessorID))

	tflog.Trace(ctx, "created katello lifecycle environment", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *katelloLifecycleEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state katelloLifecycleEnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	result, err := r.client.ReadKatelloLifecycleEnvironment(ctx, id)
	if err != nil {
		if errors.Is(err, goforeman.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read katello lifecycle environment, got error: %s", err))
		return
	}
	state.Name = types.StringValue(result.Name)
	state.Description = types.StringValue(result.Description)
	state.Label = types.StringValue(result.Label)
	state.OrganizationID = types.Int64Value(int64(result.OrganizationID))
	state.Library = types.BoolValue(result.Library)
	state.PriorID = types.Int64Value(int64(result.PriorID))
	state.SuccessorID = types.Int64Value(int64(result.SuccessorID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *katelloLifecycleEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan katelloLifecycleEnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	body := &goforeman.KatelloLifecycleEnvironmentRequest{
		Name:           plan.Name.ValueString(),
		Description:    plan.Description.ValueString(),
		Label:          plan.Label.ValueString(),
		OrganizationID: int(plan.OrganizationID.ValueInt64()),
		PriorID:        int(plan.PriorID.ValueInt64()),
	}

	result, err := r.client.UpdateKatelloLifecycleEnvironment(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update katello lifecycle environment, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(result.Name)
	plan.Description = types.StringValue(result.Description)
	plan.Label = types.StringValue(result.Label)
	plan.OrganizationID = types.Int64Value(int64(result.OrganizationID))
	plan.Library = types.BoolValue(result.Library)
	plan.PriorID = types.Int64Value(int64(result.PriorID))
	plan.SuccessorID = types.Int64Value(int64(result.SuccessorID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *katelloLifecycleEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state katelloLifecycleEnvironmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteKatelloLifecycleEnvironment(ctx, id)
	if err != nil && !errors.Is(err, goforeman.ErrNotFound) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete katello lifecycle environment, got error: %s", err))
		return
	}
}

func (r *katelloLifecycleEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

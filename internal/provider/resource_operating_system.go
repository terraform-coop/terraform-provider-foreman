package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/terraform-coop/terraform-provider-foreman/generated"
)

var (
	_ resource.Resource                = &operatingsystemResource{}
	_ resource.ResourceWithImportState = &operatingsystemResource{}
)

func NewForemanOperatingSystemResource() resource.Resource {
	return &operatingsystemResource{}
}

type operatingsystemResource struct {
	client *generated.ForemanClient
}

type operatingsystemResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Major                   types.String `tfsdk:"major"`
	Name                    types.String `tfsdk:"name"`
	ArchitectureIDs         types.List   `tfsdk:"architecture_ids"`
	Description             types.String `tfsdk:"description"`
	Family                  types.String `tfsdk:"family"`
	MediumIDs               types.List   `tfsdk:"medium_ids"`
	Minor                   types.String `tfsdk:"minor"`
	PasswordHash            types.String `tfsdk:"password_hash"`
	ProvisioningTemplateIDs types.List   `tfsdk:"provisioning_template_ids"`
	PtableIDs               types.List   `tfsdk:"ptable_ids"`
	ReleaseName             types.String `tfsdk:"release_name"`
	OsParametersAttributes  types.Map    `tfsdk:"os_parameters_attributes"`
}

func (r *operatingsystemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_operatingsystem"
}

func (r *operatingsystemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"major": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"architecture_ids": schema.ListAttribute{
				Required:    false,
				Optional:    true,
				Description: "IDs of associated architectures",
				ElementType: types.Int64Type,
			},
			"description": schema.StringAttribute{
				Required: false,
				Optional: true,
			},
			"family": schema.StringAttribute{
				Required: false,
				Optional: true,
			},
			"medium_ids": schema.ListAttribute{
				Required:    false,
				Optional:    true,
				Description: "IDs of associated media",
				ElementType: types.Int64Type,
			},
			"minor": schema.StringAttribute{
				Required: false,
				Optional: true,
			},
			"password_hash": schema.StringAttribute{
				Required:    false,
				Optional:    true,
				Description: "Root password hash function to use",
			},
			"provisioning_template_ids": schema.ListAttribute{
				Required:    false,
				Optional:    true,
				Description: "IDs of associated provisioning templates",
				ElementType: types.Int64Type,
			},
			"ptable_ids": schema.ListAttribute{
				Required:    false,
				Optional:    true,
				Description: "IDs of associated partition tables",
				ElementType: types.Int64Type,
			},
			"release_name": schema.StringAttribute{
				Required: false,
				Optional: true,
			},
			"os_parameters_attributes": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Operating system parameters as key-value pairs.",
			},
		},
	}
}

func (r *operatingsystemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type foremanOperatingSystemFullRequest struct {
	generated.ForemanOperatingSystemRequest
	OsParametersAttributes []map[string]interface{} `json:"os_parameters_attributes,omitempty"`
}

type foremanOperatingSystemFullResponse struct {
	generated.ForemanOperatingSystem
	OsParametersAttributes json.RawMessage `json:"os_parameters_attributes"`
}

func buildOperatingSystemRequest(plan operatingsystemResourceModel, diags *diag.Diagnostics) *foremanOperatingSystemFullRequest {
	return &foremanOperatingSystemFullRequest{
		ForemanOperatingSystemRequest: generated.ForemanOperatingSystemRequest{
			Major:                   plan.Major.ValueString(),
			Name:                    plan.Name.ValueString(),
			ArchitectureIDs:         listToInt64Slice(plan.ArchitectureIDs),
			Description:             plan.Description.ValueString(),
			Family:                  plan.Family.ValueString(),
			MediumIDs:               listToInt64Slice(plan.MediumIDs),
			Minor:                   plan.Minor.ValueString(),
			PasswordHash:            plan.PasswordHash.ValueString(),
			ProvisioningTemplateIDs: listToInt64Slice(plan.ProvisioningTemplateIDs),
			PtableIDs:               listToInt64Slice(plan.PtableIDs),
			ReleaseName:             plan.ReleaseName.ValueString(),
		},
		OsParametersAttributes: flattenParameters(plan.OsParametersAttributes),
	}
}

func (r *operatingsystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan operatingsystemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildOperatingSystemRequest(plan, &resp.Diagnostics)

	var result foremanOperatingSystemFullResponse
	err := r.client.Post(ctx, "operatingsystems", "operatingsystem", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create operatingsystem, got error: %s", err))
		return
	}

	marshalOperatingSystemResultToState(&result, &plan, &resp.Diagnostics)

	tflog.Trace(ctx, "created operatingsystem", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *operatingsystemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state operatingsystemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}
	var result foremanOperatingSystemFullResponse
	err = r.client.Get(ctx, fmt.Sprintf("operatingsystems/%d", id), &result)
	if err != nil {
		if generated.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read operatingsystem, got error: %s", err))
		return
	}

	marshalOperatingSystemResultToState(&result, &state, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *operatingsystemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan operatingsystemResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	body := buildOperatingSystemRequest(plan, &resp.Diagnostics)

	var result foremanOperatingSystemFullResponse
	err = r.client.Put(ctx, fmt.Sprintf("operatingsystems/%d", id), "operatingsystem", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update operatingsystem, got error: %s", err))
		return
	}

	marshalOperatingSystemResultToState(&result, &plan, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *operatingsystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state operatingsystemResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}
	err = r.client.DeleteForemanOperatingSystem(ctx, id)
	if err != nil && !generated.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete operatingsystem, got error: %s", err))
		return
	}
}

func marshalOperatingSystemResultToState(result *foremanOperatingSystemFullResponse, state *operatingsystemResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	state.Major = types.StringValue(result.Major)
	state.Name = types.StringValue(result.Name)
	state.ArchitectureIDs = int64SliceToList(result.ArchitectureIDs, diags)
	state.Description = types.StringValue(result.Description)
	state.Family = types.StringValue(result.Family)
	state.MediumIDs = int64SliceToList(result.MediumIDs, diags)
	state.Minor = types.StringValue(result.Minor)
	state.PasswordHash = types.StringValue(result.PasswordHash)
	state.ProvisioningTemplateIDs = int64SliceToList(result.ProvisioningTemplateIDs, diags)
	state.PtableIDs = int64SliceToList(result.PtableIDs, diags)
	state.ReleaseName = types.StringValue(result.ReleaseName)
	state.OsParametersAttributes = expandParameters(result.OsParametersAttributes)
}

func (r *operatingsystemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

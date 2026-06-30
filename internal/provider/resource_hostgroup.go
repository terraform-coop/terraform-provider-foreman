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
	_ resource.Resource                = &hostgroupResource{}
	_ resource.ResourceWithImportState = &hostgroupResource{}
)

func NewForemanHostgroupResource() resource.Resource {
	return &hostgroupResource{}
}

type hostgroupResource struct {
	client *generated.ForemanClient
}

// foremanHostgroupWithParams wraps the generated hostgroup response so we can
// capture the raw parameters array for bridging.go conversion.
type foremanHostgroupWithParams struct {
	generated.ForemanHostgroup
	Parameters json.RawMessage `json:"parameters"`
}

// foremanHostgroupFullRequest wraps the generated request so we can send the
// parameters array that the generator excluded as a RawMessage field.
type foremanHostgroupFullRequest struct {
	generated.ForemanHostgroupRequest
	GroupParametersAttributes []map[string]interface{} `json:"group_parameters_attributes,omitempty"`
}

type hostgroupResourceModel struct {
	ID                types.String `tfsdk:"id"`
	ArchitectureID    types.Int64  `tfsdk:"architecture_id"`
	ComputeProfileID  types.Int64  `tfsdk:"compute_profile_id"`
	Description       types.String `tfsdk:"description"`
	DomainID          types.Int64  `tfsdk:"domain_id"`
	MediumID          types.Int64  `tfsdk:"medium_id"`
	OperatingsystemID types.Int64  `tfsdk:"operatingsystem_id"`
	ParentID          types.String `tfsdk:"parent_id"`
	PtableID          types.Int64  `tfsdk:"ptable_id"`
	PXELoader         types.String `tfsdk:"pxe_loader"`
	RealmID           types.String `tfsdk:"realm_id"`
	ComputeResourceID types.Int64  `tfsdk:"compute_resource_id"`
	PuppetProxyID     types.Int64  `tfsdk:"puppet_proxy_id"`
	PuppetCaProxyID   types.Int64  `tfsdk:"puppet_ca_proxy_id"`
	RootPass          types.String `tfsdk:"root_pass"`
	Subnet6ID         types.Int64  `tfsdk:"subnet6_id"`
	SubnetID          types.Int64  `tfsdk:"subnet_id"`
	Parameters        types.Map    `tfsdk:"parameters"`
}

func (r *hostgroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hostgroup"
}

func (r *hostgroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"architecture_id": schema.Int64Attribute{
				Optional: true,
			},
			"compute_profile_id": schema.Int64Attribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"domain_id": schema.Int64Attribute{
				Optional: true,
			},
			"medium_id": schema.Int64Attribute{
				Optional: true,
			},
			"operatingsystem_id": schema.Int64Attribute{
				Optional: true,
			},
			"parent_id": schema.StringAttribute{
				Optional: true,
			},
			"ptable_id": schema.Int64Attribute{
				Optional: true,
			},
			"pxe_loader": schema.StringAttribute{
				Optional: true,
			},
			"realm_id": schema.StringAttribute{
				Optional: true,
			},
			"compute_resource_id": schema.Int64Attribute{
				Optional: true,
			},
			"puppet_proxy_id": schema.Int64Attribute{
				Optional: true,
			},
			"puppet_ca_proxy_id": schema.Int64Attribute{
				Optional: true,
			},
			"root_pass": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"subnet6_id": schema.Int64Attribute{
				Optional: true,
			},
			"subnet_id": schema.Int64Attribute{
				Optional: true,
			},
			"parameters": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *hostgroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *hostgroupResource) createHostgroup(ctx context.Context, req *foremanHostgroupFullRequest) (*foremanHostgroupWithParams, error) {
	var resp foremanHostgroupWithParams
	err := r.client.Post(ctx, "hostgroups", "hostgroup", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *hostgroupResource) readHostgroup(ctx context.Context, id int) (*foremanHostgroupWithParams, error) {
	var resp foremanHostgroupWithParams
	err := r.client.Get(ctx, fmt.Sprintf("hostgroups/%d", id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *hostgroupResource) updateHostgroup(ctx context.Context, id int, req *foremanHostgroupFullRequest) (*foremanHostgroupWithParams, error) {
	var resp foremanHostgroupWithParams
	err := r.client.Put(ctx, fmt.Sprintf("hostgroups/%d", id), "hostgroup", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func buildHostgroupRequest(plan hostgroupResourceModel, diags *diag.Diagnostics) *foremanHostgroupFullRequest {
	return &foremanHostgroupFullRequest{
		ForemanHostgroupRequest: generated.ForemanHostgroupRequest{
			ArchitectureID:    plan.ArchitectureID.ValueInt64(),
			ComputeProfileID:  plan.ComputeProfileID.ValueInt64(),
			ComputeResourceID: plan.ComputeResourceID.ValueInt64(),
			Description:       plan.Description.ValueString(),
			DomainID:          plan.DomainID.ValueInt64(),
			MediumID:          plan.MediumID.ValueInt64(),
			OperatingsystemID: plan.OperatingsystemID.ValueInt64(),
			ParentID:          parseStringToInt64(plan.ParentID.ValueString(), "parent_id", diags),
			PtableID:          plan.PtableID.ValueInt64(),
			PuppetCaProxyID:   plan.PuppetCaProxyID.ValueInt64(),
			PuppetProxyID:     plan.PuppetProxyID.ValueInt64(),
			PXELoader:         plan.PXELoader.ValueString(),
			RealmID:           parseStringToInt64(plan.RealmID.ValueString(), "realm_id", diags),
			RootPass:          plan.RootPass.ValueString(),
			Subnet6ID:         plan.Subnet6ID.ValueInt64(),
			SubnetID:          plan.SubnetID.ValueInt64(),
		},
		GroupParametersAttributes: flattenParameters(plan.Parameters),
	}
}

func marshalHostgroupResultToState(result *foremanHostgroupWithParams, state *hostgroupResourceModel, diags *diag.Diagnostics) {
	state.ArchitectureID = types.Int64Value(int64(result.ArchitectureID))
	state.ComputeProfileID = types.Int64Value(int64(result.ComputeProfileID))
	state.ComputeResourceID = types.Int64Value(parseStringToInt64(result.ComputeResourceID, "compute_resource_id", diags))
	state.Description = types.StringValue(result.Description)
	state.DomainID = types.Int64Value(int64(result.DomainID))
	state.MediumID = types.Int64Value(int64(result.MediumID))
	state.OperatingsystemID = types.Int64Value(int64(result.OperatingsystemID))
	state.ParentID = types.StringValue(result.ParentID)
	state.PtableID = types.Int64Value(int64(result.PtableID))
	state.PuppetCaProxyID = types.Int64Value(int64(result.PuppetCaProxyID))
	state.PuppetProxyID = types.Int64Value(int64(result.PuppetProxyID))
	state.PXELoader = types.StringValue(result.PXELoader)
	state.RealmID = types.StringValue(result.RealmID)
	// RootPass is not returned by the API; preserve the value already in state.
	state.Subnet6ID = types.Int64Value(int64(result.Subnet6ID))
	state.SubnetID = types.Int64Value(int64(result.SubnetID))
	state.Parameters = expandParameters(result.Parameters)
}

func (r *hostgroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostgroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildHostgroupRequest(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.createHostgroup(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create hostgroup, got error: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	marshalHostgroupResultToState(result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created hostgroup", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostgroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostgroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	result, err := r.readHostgroup(ctx, id)
	if err != nil {
		if generated.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read hostgroup, got error: %s", err))
		return
	}

	marshalHostgroupResultToState(result, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hostgroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hostgroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	body := buildHostgroupRequest(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.updateHostgroup(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update hostgroup, got error: %s", err))
		return
	}

	marshalHostgroupResultToState(result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostgroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostgroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteForemanHostgroup(ctx, id)
	if err != nil && !generated.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete hostgroup, got error: %s", err))
		return
	}
}

func (r *hostgroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

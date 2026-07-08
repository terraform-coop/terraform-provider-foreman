package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                   = &parameterResource{}
	_ resource.ResourceWithImportState    = &parameterResource{}
	_ resource.ResourceWithValidateConfig = &parameterResource{}
)

func NewParameterResource() resource.Resource {
	return &parameterResource{}
}

type parameterResource struct {
	client *goforeman.Client
}

type parameterResourceModel struct {
	ID                types.String `tfsdk:"id"`
	HostID            types.Int64  `tfsdk:"host_id"`
	HostgroupID       types.Int64  `tfsdk:"hostgroup_id"`
	DomainID          types.Int64  `tfsdk:"domain_id"`
	OperatingsystemID types.Int64  `tfsdk:"operatingsystem_id"`
	SubnetID          types.Int64  `tfsdk:"subnet_id"`
	LocationID        types.Int64  `tfsdk:"location_id"`
	OrganizationID    types.Int64  `tfsdk:"organization_id"`
	Name              types.String `tfsdk:"name"`
	Value             types.String `tfsdk:"value"`
	ParameterType     types.String `tfsdk:"parameter_type"`
	HiddenValue       types.Bool   `tfsdk:"hidden_value"`
}

func (r *parameterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_parameter"
}

func (r *parameterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a parameter scoped to exactly one of host_id, hostgroup_id, " +
			"domain_id, operatingsystem_id, subnet_id, location_id, or organization_id - " +
			"set exactly one of these.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host_id":            schema.Int64Attribute{Optional: true},
			"hostgroup_id":       schema.Int64Attribute{Optional: true},
			"domain_id":          schema.Int64Attribute{Optional: true},
			"operatingsystem_id": schema.Int64Attribute{Optional: true},
			"subnet_id":          schema.Int64Attribute{Optional: true},
			"location_id":        schema.Int64Attribute{Optional: true},
			"organization_id":    schema.Int64Attribute{Optional: true},
			"name": schema.StringAttribute{
				Required: true,
			},
			"value": schema.StringAttribute{
				Required: true,
			},
			"parameter_type": schema.StringAttribute{
				Required:    true,
				Description: "One of string, boolean, integer, real, array, hash, yaml, json.",
			},
			"hidden_value": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (r *parameterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ValidateConfig ensures exactly one parent-scoping attribute is set, since
// Foreman requires the parameter to be scoped under exactly one parent.
func (r *parameterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data parameterResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// A parent ID referencing a not-yet-created resource is unknown at plan
	// time - the exactly-one-of rule can't be evaluated until apply, so
	// don't reject what may become valid (Create re-checks with real
	// values).
	for _, v := range []types.Int64{
		data.HostID, data.HostgroupID, data.DomainID, data.OperatingsystemID,
		data.SubnetID, data.LocationID, data.OrganizationID,
	} {
		if v.IsUnknown() {
			return
		}
	}
	_, _, diags := parameterParent(&data)
	resp.Diagnostics.Append(diags...)
}

// parameterParent converts the model's parent-scoping attributes to plain
// IDs and delegates the exactly-one-of resolution (a Foreman API
// requirement) to goforeman.ResolveParameterParent.
func parameterParent(m *parameterResourceModel) (parentType string, parentID int64, diags diag.Diagnostics) {
	setIDs := map[string]int64{}
	for field, value := range map[string]types.Int64{
		"host_id":            m.HostID,
		"hostgroup_id":       m.HostgroupID,
		"domain_id":          m.DomainID,
		"operatingsystem_id": m.OperatingsystemID,
		"subnet_id":          m.SubnetID,
		"location_id":        m.LocationID,
		"organization_id":    m.OrganizationID,
	} {
		if !value.IsNull() && !value.IsUnknown() {
			setIDs[field] = value.ValueInt64()
		}
	}
	parentType, parentID, err := goforeman.ResolveParameterParent(setIDs)
	if err != nil {
		diags.AddError("Invalid parameter scoping", err.Error())
		return "", 0, diags
	}
	return parentType, parentID, diags
}

func (r *parameterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan parameterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType, parentID, diags := parameterParent(&plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &goforeman.ParameterRequest{
		Name:          plan.Name.ValueString(),
		Value:         plan.Value.ValueString(),
		ParameterType: plan.ParameterType.ValueString(),
	}
	if !plan.HiddenValue.IsNull() && !plan.HiddenValue.IsUnknown() {
		v := plan.HiddenValue.ValueBool()
		body.HiddenValue = &v
	}

	result, err := r.client.CreateParameter(ctx, parentType, int(parentID), body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create parameter, got error: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	plan.Name = types.StringValue(result.Name)
	plan.Value = types.StringValue(goforeman.RawValueString(result.Value))
	plan.ParameterType = types.StringValue(result.ParameterType)
	plan.HiddenValue = types.BoolValue(result.HiddenValue)

	tflog.Trace(ctx, "created parameter", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *parameterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state parameterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType, parentID, diags := parameterParent(&state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	result, err := r.client.ReadParameter(ctx, parentType, int(parentID), id)
	if err != nil {
		if errors.Is(err, goforeman.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read parameter, got error: %s", err))
		return
	}
	state.Name = types.StringValue(result.Name)
	state.Value = types.StringValue(goforeman.RawValueString(result.Value))
	state.ParameterType = types.StringValue(result.ParameterType)
	state.HiddenValue = types.BoolValue(result.HiddenValue)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *parameterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan parameterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType, parentID, diags := parameterParent(&plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	body := &goforeman.ParameterRequest{
		Name:          plan.Name.ValueString(),
		Value:         plan.Value.ValueString(),
		ParameterType: plan.ParameterType.ValueString(),
	}
	if !plan.HiddenValue.IsNull() && !plan.HiddenValue.IsUnknown() {
		v := plan.HiddenValue.ValueBool()
		body.HiddenValue = &v
	}

	result, err := r.client.UpdateParameter(ctx, parentType, int(parentID), id, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update parameter, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(result.Name)
	plan.Value = types.StringValue(goforeman.RawValueString(result.Value))
	plan.ParameterType = types.StringValue(result.ParameterType)
	plan.HiddenValue = types.BoolValue(result.HiddenValue)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *parameterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state parameterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType, parentID, diags := parameterParent(&state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteParameter(ctx, parentType, int(parentID), id)
	if err != nil && !errors.Is(err, goforeman.ErrNotFound) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete parameter, got error: %s", err))
		return
	}
}

// parseParameterImportID parses the "<parent_field>:<parent_id>:<parameter_id>"
// import identifier (e.g. "host_id:5:42") into its three parts, or returns
// an error describing exactly what's wrong.
func parseParameterImportID(id string) (parentField string, parentID int64, paramID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", 0, "", fmt.Errorf("expected format \"<parent_field>:<parent_id>:<parameter_id>\" (e.g. \"host_id:5:42\"), got: %s", id)
	}
	parentField, parentIDStr, paramID := parts[0], parts[1], parts[2]

	if _, ok := goforeman.ParameterParentTypes[parentField]; !ok {
		return "", 0, "", fmt.Errorf("unknown parent field %q; must be one of %s", parentField, strings.Join(goforeman.ParameterParentFields, ", "))
	}
	parentID, err = strconv.ParseInt(parentIDStr, 10, 64)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid parent ID %q: %w", parentIDStr, err)
	}
	return parentField, parentID, paramID, nil
}

// ImportState expects "<parent_field>:<parent_id>:<parameter_id>", e.g.
// "host_id:5:42" - a bare numeric ID isn't enough to know which of the 7
// possible parent-scoped endpoints to read the parameter back from.
func (r *parameterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parentField, parentID, paramID, err := parseParameterImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Import Identifier", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(parentField), parentID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), paramID)...)
}

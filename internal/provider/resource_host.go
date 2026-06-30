package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                = &hostResource{}
	_ resource.ResourceWithImportState = &hostResource{}
)

func NewForemanHostResource() resource.Resource {
	return &hostResource{}
}

type hostResource struct {
	client *generated.ForemanClient
}

type hostResourceModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	ArchitectureID           types.Int64  `tfsdk:"architecture_id"`
	BmcAvailable             types.Bool   `tfsdk:"bmc_available"`
	Build                    types.Bool   `tfsdk:"build"`
	BuildStatus              types.Int64  `tfsdk:"build_status"`
	BuildStatusLabel         types.String `tfsdk:"build_status_label"`
	Certname                 types.String `tfsdk:"certname"`
	Comment                  types.String `tfsdk:"comment"`
	ComputeProfileID         types.String `tfsdk:"compute_profile_id"`
	ComputeResourceID        types.Int64  `tfsdk:"compute_resource_id"`
	ComputeResourceProvider  types.String `tfsdk:"compute_resource_provider"`
	Creator                  types.String `tfsdk:"creator"`
	CreatorID                types.Int64  `tfsdk:"creator_id"`
	Disk                     types.String `tfsdk:"disk"`
	DisplayName              types.String `tfsdk:"display_name"`
	DomainID                 types.Int64  `tfsdk:"domain_id"`
	Enabled                  types.Bool   `tfsdk:"enabled"`
	GlobalStatus             types.Int64  `tfsdk:"global_status"`
	GlobalStatusFulltext     types.List   `tfsdk:"global_status_fulltext"`
	GlobalStatusLabel        types.String `tfsdk:"global_status_label"`
	HostgroupID              types.Int64  `tfsdk:"hostgroup_id"`
	ImageID                  types.String `tfsdk:"image_id"`
	InitiatedAt              types.String `tfsdk:"initiated_at"`
	InstalledAt              types.String `tfsdk:"installed_at"`
	IP                       types.String `tfsdk:"ip"`
	IP6                      types.String `tfsdk:"ip6"`
	LastCompile              types.String `tfsdk:"last_compile"`
	LastReport               types.String `tfsdk:"last_report"`
	MAC                      types.String `tfsdk:"mac"`
	Managed                  types.Bool   `tfsdk:"managed"`
	MediumID                 types.Int64  `tfsdk:"medium_id"`
	ModelID                  types.String `tfsdk:"model_id"`
	OperatingsystemIcon      types.String `tfsdk:"operatingsystem_icon"`
	OperatingsystemID        types.Int64  `tfsdk:"operatingsystem_id"`
	OwnerID                  types.Int64  `tfsdk:"owner_id"`
	OwnerType                types.String `tfsdk:"owner_type"`
	Permissions              types.Map    `tfsdk:"permissions"`
	ProvisionMethod          types.String `tfsdk:"provision_method"`
	PtableID                 types.Int64  `tfsdk:"ptable_id"`
	PuppetCaProxyID          types.String `tfsdk:"puppet_ca_proxy_id"`
	PuppetProxyID            types.String `tfsdk:"puppet_proxy_id"`
	PXELoader                types.String `tfsdk:"pxe_loader"`
	RealmID                  types.String `tfsdk:"realm_id"`
	RebuildRequiresPoweroff  types.Bool   `tfsdk:"rebuild_requires_poweroff"`
	SpIP                     types.String `tfsdk:"sp_ip"`
	SpMAC                    types.String `tfsdk:"sp_mac"`
	SpName                   types.String `tfsdk:"sp_name"`
	SpSubnetID               types.String `tfsdk:"sp_subnet_id"`
	Subnet6ID                types.String `tfsdk:"subnet6_id"`
	SubnetID                 types.String `tfsdk:"subnet_id"`
	UseImage                 types.String `tfsdk:"use_image"`
	InterfacesAttributes     types.List   `tfsdk:"interfaces_attributes"`
	HostParametersAttributes types.Map    `tfsdk:"host_parameters_attributes"`
	ComputeAttributes        types.String `tfsdk:"compute_attributes"`
}

func (r *hostResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_host"
}

func (r *hostResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"architecture_id": schema.Int64Attribute{
				Optional: true,
			},
			"bmc_available": schema.BoolAttribute{
				Computed: true,
			},
			"build": schema.BoolAttribute{
				Optional: true,
			},
			"build_status": schema.Int64Attribute{
				Computed: true,
			},
			"build_status_label": schema.StringAttribute{
				Computed: true,
			},
			"certname": schema.StringAttribute{
				Computed: true,
			},
			"comment": schema.StringAttribute{
				Optional: true,
			},
			"compute_profile_id": schema.StringAttribute{
				Optional: true,
			},
			"compute_resource_id": schema.Int64Attribute{
				Optional: true,
			},
			"compute_resource_provider": schema.StringAttribute{
				Computed: true,
			},
			"creator": schema.StringAttribute{
				Computed: true,
			},
			"creator_id": schema.Int64Attribute{
				Computed: true,
			},
			"disk": schema.StringAttribute{
				Computed: true,
			},
			"display_name": schema.StringAttribute{
				Computed: true,
			},
			"domain_id": schema.Int64Attribute{
				Optional: true,
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
			},
			"global_status": schema.Int64Attribute{
				Computed: true,
			},
			"global_status_fulltext": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"global_status_label": schema.StringAttribute{
				Computed: true,
			},
			"hostgroup_id": schema.Int64Attribute{
				Optional: true,
			},
			"image_id": schema.StringAttribute{
				Optional: true,
			},
			"initiated_at": schema.StringAttribute{
				Computed: true,
			},
			"installed_at": schema.StringAttribute{
				Computed: true,
			},
			"ip": schema.StringAttribute{
				Optional: true,
			},
			"ip6": schema.StringAttribute{
				Computed: true,
			},
			"last_compile": schema.StringAttribute{
				Computed: true,
			},
			"last_report": schema.StringAttribute{
				Computed: true,
			},
			"mac": schema.StringAttribute{
				Optional: true,
			},
			"managed": schema.BoolAttribute{
				Optional: true,
			},
			"medium_id": schema.Int64Attribute{
				Optional: true,
			},
			"model_id": schema.StringAttribute{
				Optional: true,
			},
			"operatingsystem_icon": schema.StringAttribute{
				Computed: true,
			},
			"operatingsystem_id": schema.Int64Attribute{
				Optional: true,
			},
			"owner_id": schema.Int64Attribute{
				Optional: true,
			},
			"owner_type": schema.StringAttribute{
				Optional: true,
			},
			"permissions": schema.MapAttribute{
				Computed: true,
			},
			"provision_method": schema.StringAttribute{
				Optional: true,
			},
			"ptable_id": schema.Int64Attribute{
				Optional: true,
			},
			"puppet_ca_proxy_id": schema.StringAttribute{
				Optional: true,
			},
			"puppet_proxy_id": schema.StringAttribute{
				Optional: true,
			},
			"pxe_loader": schema.StringAttribute{
				Optional: true,
			},
			"realm_id": schema.StringAttribute{
				Optional: true,
			},
			"rebuild_requires_poweroff": schema.BoolAttribute{
				Computed: true,
			},
			"sp_ip": schema.StringAttribute{
				Computed: true,
			},
			"sp_mac": schema.StringAttribute{
				Computed: true,
			},
			"sp_name": schema.StringAttribute{
				Computed: true,
			},
			"sp_subnet_id": schema.StringAttribute{
				Computed: true,
			},
			"subnet6_id": schema.StringAttribute{
				Computed: true,
			},
			"subnet_id": schema.StringAttribute{
				Optional: true,
			},
			"use_image": schema.StringAttribute{
				Computed: true,
			},
			"interfaces_attributes": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Host interface information. One block per interface.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                 schema.Int64Attribute{Computed: true},
						"primary":            schema.BoolAttribute{Optional: true},
						"ip":                 schema.StringAttribute{Optional: true, Computed: true},
						"mac":                schema.StringAttribute{Optional: true, Computed: true},
						"name":               schema.StringAttribute{Optional: true, Computed: true},
						"subnet_id":          schema.Int64Attribute{Optional: true, Computed: true},
						"identifier":         schema.StringAttribute{Optional: true},
						"managed":            schema.BoolAttribute{Optional: true},
						"provision":          schema.BoolAttribute{Optional: true},
						"virtual":            schema.BoolAttribute{Optional: true},
						"type":               schema.StringAttribute{Optional: true},
						"bmc_provider":       schema.StringAttribute{Optional: true},
						"username":           schema.StringAttribute{Optional: true},
						"password":           schema.StringAttribute{Optional: true, Sensitive: true},
						"domain_id":          schema.Int64Attribute{Optional: true, Computed: true},
						"attached_to":        schema.StringAttribute{Optional: true},
						"attached_devices":   schema.StringAttribute{Optional: true},
						"compute_attributes": schema.StringAttribute{Optional: true, Description: "Hypervisor specific interface options (JSON)"},
					},
				},
			},
			"host_parameters_attributes": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Host parameters as key-value pairs.",
			},
			"compute_attributes": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Hypervisor specific compute options (JSON string).",
			},
		},
	}
}

func (r *hostResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *hostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan hostResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildHostRequest(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var result foremanHostFullResponse
	err := r.client.Post(ctx, "hosts", "host", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create host, got error: %s", err))
		return
	}

	marshalHostResultToState(&result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created host", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state hostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}
	var result foremanHostFullResponse
	err = r.client.Get(ctx, fmt.Sprintf("hosts/%d", id), &result)
	if err != nil {
		if generated.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read host, got error: %s", err))
		return
	}

	marshalHostResultToState(&result, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *hostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan hostResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	body := buildHostRequest(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var result foremanHostFullResponse
	err = r.client.Put(ctx, fmt.Sprintf("hosts/%d", id), "host", body, &result)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update host, got error: %s", err))
		return
	}

	marshalHostResultToState(&result, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *hostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state hostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}
	err = r.client.DeleteForemanHost(ctx, id)
	if err != nil && !generated.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete host, got error: %s", err))
		return
	}
}

func (r *hostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildHostRequest builds the full API request body from the Terraform plan.
func buildHostRequest(plan hostResourceModel, diags *diag.Diagnostics) *foremanHostFullRequest {
	return &foremanHostFullRequest{
		ForemanHostRequest: generated.ForemanHostRequest{
			Name:              plan.Name.ValueString(),
			ArchitectureID:    plan.ArchitectureID.ValueInt64(),
			Build:             plan.Build.ValueBool(),
			Comment:           plan.Comment.ValueString(),
			ComputeProfileID:  parseStringToInt64(plan.ComputeProfileID.ValueString(), "compute_profile_id", diags),
			ComputeResourceID: plan.ComputeResourceID.ValueInt64(),
			DomainID:          plan.DomainID.ValueInt64(),
			Enabled:           plan.Enabled.ValueBool(),
			HostgroupID:       plan.HostgroupID.ValueInt64(),
			ImageID:           parseStringToInt64(plan.ImageID.ValueString(), "image_id", diags),
			IP:                plan.IP.ValueString(),
			MAC:               plan.MAC.ValueString(),
			Managed:           plan.Managed.ValueBool(),
			MediumID:          plan.MediumID.ValueInt64(),
			ModelID:           parseStringToInt64(plan.ModelID.ValueString(), "model_id", diags),
			OperatingsystemID: plan.OperatingsystemID.ValueInt64(),
			OwnerID:           plan.OwnerID.ValueInt64(),
			OwnerType:         plan.OwnerType.ValueString(),
			ProvisionMethod:   plan.ProvisionMethod.ValueString(),
			PtableID:          plan.PtableID.ValueInt64(),
			PuppetCaProxyID:   parseStringToInt64(plan.PuppetCaProxyID.ValueString(), "puppet_ca_proxy_id", diags),
			PuppetProxyID:     parseStringToInt64(plan.PuppetProxyID.ValueString(), "puppet_proxy_id", diags),
			PXELoader:         plan.PXELoader.ValueString(),
			RealmID:           parseStringToInt64(plan.RealmID.ValueString(), "realm_id", diags),
			SubnetID:          parseStringToInt64(plan.SubnetID.ValueString(), "subnet_id", diags),
		},
		InterfacesAttributes:     flattenInterfacesAttributes(plan.InterfacesAttributes),
		HostParametersAttributes: flattenParameters(plan.HostParametersAttributes),
		ComputeAttributes:        flattenComputeAttributes(plan.ComputeAttributes),
	}
}

// marshalHostResultToState maps a bridged API response back into the Terraform state model.
func marshalHostResultToState(result *foremanHostFullResponse, state *hostResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(strconv.Itoa(result.ID))
	state.Name = types.StringValue(result.Name)
	state.ArchitectureID = types.Int64Value(result.ArchitectureID)
	state.BmcAvailable = types.BoolValue(result.BmcAvailable)
	state.Build = types.BoolValue(result.Build)
	state.BuildStatus = types.Int64Value(result.BuildStatus)
	state.BuildStatusLabel = types.StringValue(result.BuildStatusLabel)
	state.Certname = types.StringValue(result.Certname)
	state.Comment = types.StringValue(result.Comment)
	state.ComputeProfileID = types.StringValue(result.ComputeProfileID)
	state.ComputeResourceID = types.Int64Value(result.ComputeResourceID)
	state.ComputeResourceProvider = types.StringValue(result.ComputeResourceProvider)
	state.Creator = types.StringValue(result.Creator)
	state.CreatorID = types.Int64Value(result.CreatorID)
	state.Disk = types.StringValue(result.Disk)
	state.DisplayName = types.StringValue(result.DisplayName)
	state.DomainID = types.Int64Value(result.DomainID)
	state.Enabled = types.BoolValue(result.Enabled)
	state.GlobalStatus = types.Int64Value(result.GlobalStatus)
	state.GlobalStatusLabel = types.StringValue(result.GlobalStatusLabel)
	state.HostgroupID = types.Int64Value(result.HostgroupID)
	state.ImageID = types.StringValue(result.ImageID)
	state.InitiatedAt = types.StringValue(result.InitiatedAt)
	state.InstalledAt = types.StringValue(result.InstalledAt)
	state.IP = types.StringValue(result.IP)
	state.IP6 = types.StringValue(result.IP6)
	state.LastCompile = types.StringValue(result.LastCompile)
	state.LastReport = types.StringValue(result.LastReport)
	state.MAC = types.StringValue(result.MAC)
	state.Managed = types.BoolValue(result.Managed)
	state.MediumID = types.Int64Value(result.MediumID)
	state.ModelID = types.StringValue(result.ModelID)
	state.OperatingsystemIcon = types.StringValue(result.OperatingsystemIcon)
	state.OperatingsystemID = types.Int64Value(result.OperatingsystemID)
	state.OwnerID = types.Int64Value(result.OwnerID)
	state.OwnerType = types.StringValue(result.OwnerType)
	state.ProvisionMethod = types.StringValue(result.ProvisionMethod)
	state.PtableID = types.Int64Value(result.PtableID)
	state.PuppetCaProxyID = types.StringValue(result.PuppetCaProxyID)
	state.PuppetProxyID = types.StringValue(result.PuppetProxyID)
	state.PXELoader = types.StringValue(result.PXELoader)
	state.RealmID = types.StringValue(result.RealmID)
	state.RebuildRequiresPoweroff = types.BoolValue(result.RebuildRequiresPoweroff)
	state.SpIP = types.StringValue(result.SpIP)
	state.SpMAC = types.StringValue(result.SpMAC)
	state.SpName = types.StringValue(result.SpName)
	state.SpSubnetID = types.StringValue(result.SpSubnetID)
	state.Subnet6ID = types.StringValue(result.Subnet6ID)
	state.SubnetID = types.StringValue(result.SubnetID)
	state.UseImage = types.StringValue(result.UseImage)

	if len(result.GlobalStatusFulltext) > 0 {
		elems := make([]attr.Value, len(result.GlobalStatusFulltext))
		for i, s := range result.GlobalStatusFulltext {
			elems[i] = types.StringValue(s)
		}
		listVal, d := types.ListValue(types.StringType, elems)
		diags.Append(d...)
		state.GlobalStatusFulltext = listVal
	} else {
		state.GlobalStatusFulltext = types.ListNull(types.StringType)
	}

	if len(result.Permissions) > 0 {
		elems := make(map[string]attr.Value, len(result.Permissions))
		for k, v := range result.Permissions {
			elems[k] = types.StringValue(fmt.Sprint(v))
		}
		mapVal, d := types.MapValue(types.StringType, elems)
		diags.Append(d...)
		state.Permissions = mapVal
	} else {
		state.Permissions = types.MapNull(types.StringType)
	}

	state.InterfacesAttributes = expandInterfacesAttributes(result.InterfacesAttributes, diags)
	state.HostParametersAttributes = expandParameters(result.HostParametersAttributes)
	state.ComputeAttributes = expandComputeAttributes(result.ComputeAttributes)
}

// ---------------------------------------------------------------------------
// Bridged request/response types
// ---------------------------------------------------------------------------

// foremanHostFullRequest extends the generated request with fields the
// generated code does not marshal (interfaces_attributes,
// host_parameters_attributes, compute_attributes).
type foremanHostFullRequest struct {
	generated.ForemanHostRequest
	InterfacesAttributes     []map[string]interface{} `json:"interfaces_attributes,omitempty"`
	HostParametersAttributes []map[string]interface{} `json:"host_parameters_attributes,omitempty"`
	ComputeAttributes        map[string]interface{}   `json:"compute_attributes,omitempty"`
}

// foremanHostFullResponse captures the full JSON response including the
// three bridged fields that the generated ForemanHost struct does not contain.
type foremanHostFullResponse struct {
	generated.ForemanHost
	InterfacesAttributes     json.RawMessage `json:"interfaces_attributes"`
	HostParametersAttributes json.RawMessage `json:"host_parameters_attributes"`
	ComputeAttributes        json.RawMessage `json:"compute_attributes"`
}

// ---------------------------------------------------------------------------
// Interfaces bridging — host-specific (18-attribute nested object)
// ---------------------------------------------------------------------------

var interfaceAttrTypes = map[string]attr.Type{
	"id":                 types.Int64Type,
	"primary":            types.BoolType,
	"ip":                 types.StringType,
	"mac":                types.StringType,
	"name":               types.StringType,
	"subnet_id":          types.Int64Type,
	"identifier":         types.StringType,
	"managed":            types.BoolType,
	"provision":          types.BoolType,
	"virtual":            types.BoolType,
	"type":               types.StringType,
	"bmc_provider":       types.StringType,
	"username":           types.StringType,
	"password":           types.StringType,
	"domain_id":          types.Int64Type,
	"attached_to":        types.StringType,
	"attached_devices":   types.StringType,
	"compute_attributes": types.StringType,
}

func flattenInterfacesAttributes(l types.List) []map[string]interface{} {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	elements := l.Elements()
	out := make([]map[string]interface{}, 0, len(elements))
	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		a := obj.Attributes()
		m := map[string]interface{}{
			"id":                 maybeInt64(a["id"]),
			"primary":            maybeBool(a["primary"]),
			"ip":                 maybeString(a["ip"]),
			"mac":                maybeString(a["mac"]),
			"name":               maybeString(a["name"]),
			"subnet_id":          maybeInt64(a["subnet_id"]),
			"identifier":         maybeString(a["identifier"]),
			"managed":            maybeBool(a["managed"]),
			"provision":          maybeBool(a["provision"]),
			"virtual":            maybeBool(a["virtual"]),
			"type":               maybeString(a["type"]),
			"provider":           maybeString(a["bmc_provider"]),
			"username":           maybeString(a["username"]),
			"password":           maybeString(a["password"]),
			"domain_id":          maybeInt64(a["domain_id"]),
			"attached_to":        maybeString(a["attached_to"]),
			"attached_devices":   maybeString(a["attached_devices"]),
			"compute_attributes": maybeJSON(a["compute_attributes"]),
		}
		out = append(out, m)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func expandInterfacesAttributes(raw json.RawMessage, diags *diag.Diagnostics) types.List {
	if len(raw) == 0 || string(raw) == "null" {
		return types.ListNull(types.ObjectType{AttrTypes: interfaceAttrTypes})
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(raw, &items); err != nil {
		return types.ListNull(types.ObjectType{AttrTypes: interfaceAttrTypes})
	}
	elems := make([]attr.Value, 0, len(items))
	for _, item := range items {
		obj := map[string]attr.Value{
			"id":                 int64Value(item["id"]),
			"primary":            boolValue(item["primary"]),
			"ip":                 stringValue(item["ip"]),
			"mac":                stringValue(item["mac"]),
			"name":               stringValue(item["name"]),
			"subnet_id":          int64Value(item["subnet_id"]),
			"identifier":         stringValue(item["identifier"]),
			"managed":            boolValue(item["managed"]),
			"provision":          boolValue(item["provision"]),
			"virtual":            boolValue(item["virtual"]),
			"type":               stringValue(item["type"]),
			"bmc_provider":       stringValue(item["bmc_provider"]),
			"username":           stringValue(item["username"]),
			"password":           stringValue(item["password"]),
			"domain_id":          int64Value(item["domain_id"]),
			"attached_to":        stringValue(item["attached_to"]),
			"attached_devices":   stringValue(item["attached_devices"]),
			"compute_attributes": stringValue(item["compute_attributes"]),
		}
		val, objDiags := types.ObjectValue(interfaceAttrTypes, obj)
		if objDiags.HasError() {
			diags.Append(objDiags...)
			continue
		}
		elems = append(elems, val)
	}
	listVal, listDiags := types.ListValue(types.ObjectType{AttrTypes: interfaceAttrTypes}, elems)
	if listDiags.HasError() {
		diags.Append(listDiags...)
		return types.ListNull(types.ObjectType{AttrTypes: interfaceAttrTypes})
	}
	return listVal
}

// maybeInt64 extracts int64 from attr.Value, returns 0 if null.
func maybeInt64(v attr.Value) int64 {
	if iv, ok := v.(types.Int64); ok && !iv.IsNull() {
		return iv.ValueInt64()
	}
	return 0
}

// maybeBool extracts bool from attr.Value, returns false if null.
func maybeBool(v attr.Value) bool {
	if bv, ok := v.(types.Bool); ok && !bv.IsNull() {
		return bv.ValueBool()
	}
	return false
}

// maybeString extracts string from attr.Value, returns "" if null.
func maybeString(v attr.Value) string {
	if sv, ok := v.(types.String); ok && !sv.IsNull() {
		return sv.ValueString()
	}
	return ""
}

// maybeJSON extracts a JSON string from attr.Value, unmarshals to map.
func maybeJSON(v attr.Value) map[string]interface{} {
	sv, ok := v.(types.String)
	if !ok || sv.IsNull() {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(sv.ValueString()), &m); err != nil {
		return nil
	}
	return m
}

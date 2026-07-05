package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"
)

// A compute profile's own create/update body only ever has "name"; its
// per-compute-resource VM sizing attributes are a separate nested resource
// scoped by BOTH compute_profile_id and compute_resource_id at once, which
// the generic apidoc-driven pipeline's single-parent mechanism can't
// express. Hand-written to restore the old provider's behavior of managing
// them as a list on this resource. See generated/compute_profile.go and
// tools/gen/overrides.yaml's skip_resources entry.

var _ resource.Resource = &computeprofileResource{}
var _ resource.ResourceWithImportState = &computeprofileResource{}

func NewForemanComputeProfileResource() resource.Resource {
	return &computeprofileResource{}
}

type computeprofileResource struct {
	client *goforeman.ForemanClient
}

type computeAttributeModel struct {
	ID                types.Int64  `tfsdk:"id"`
	ComputeResourceID types.Int64  `tfsdk:"compute_resource_id"`
	VMAttrs           types.String `tfsdk:"vm_attrs"`
}

type computeprofileResourceModel struct {
	ID                types.String            `tfsdk:"id"`
	Name              types.String            `tfsdk:"name"`
	ComputeAttributes []computeAttributeModel `tfsdk:"compute_attributes"`
}

func (r *computeprofileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_computeprofile"
}

func (r *computeprofileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed:      true,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		},
		"name": schema.StringAttribute{Required: true},
		"compute_attributes": schema.ListNestedAttribute{
			Description: "Per-compute-resource VM sizing attributes (cpus, memory, disks, ...). One entry per compute resource this profile is configured for.",
			Optional:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"id": schema.Int64Attribute{
					Computed:      true,
					PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				},
				"compute_resource_id": schema.Int64Attribute{
					Description: "ID of the compute resource this VM sizing applies to.",
					Required:    true,
				},
				"vm_attrs": schema.StringAttribute{
					Description: "Compute-resource-provider-specific VM sizing attributes (e.g. cpus, memory_mb, volumes_attributes for VMware) as a JSON-encoded string. Use jsonencode({...}).",
					Optional:    true,
				},
			}},
		},
	}}
}

func (r *computeprofileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*goforeman.ForemanClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *goforeman.ForemanClient")
		return
	}
	r.client = client
}

func vmAttrsToRaw(v types.String) json.RawMessage {
	if v.IsNull() || v.IsUnknown() || v.ValueString() == "" {
		return nil
	}
	return json.RawMessage(v.ValueString())
}

func vmAttrsFromRaw(raw json.RawMessage) types.String {
	if len(raw) == 0 || string(raw) == "null" {
		return types.StringNull()
	}
	return types.StringValue(string(raw))
}

func computeAttributeFromResult(ca *goforeman.ForemanComputeAttribute) computeAttributeModel {
	return computeAttributeModel{
		ID:                types.Int64Value(int64(ca.ID)),
		ComputeResourceID: types.Int64Value(int64(ca.ComputeResourceID)),
		VMAttrs:           vmAttrsFromRaw(ca.VMAttrs),
	}
}

func (r *computeprofileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan computeprofileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &goforeman.ForemanComputeProfileRequest{Name: plan.Name.ValueString()}

	result, err := r.client.CreateForemanComputeProfile(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create computeprofile, got error: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	plan.Name = types.StringValue(result.Name)

	profileID := int(result.ID)
	attrs := make([]computeAttributeModel, 0, len(plan.ComputeAttributes))
	for _, ca := range plan.ComputeAttributes {
		crID := int(ca.ComputeResourceID.ValueInt64())
		created, err := r.client.CreateForemanComputeAttribute(ctx, profileID, crID, vmAttrsToRaw(ca.VMAttrs))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create compute attribute for compute_resource_id %d, got error: %s", crID, err))
			return
		}
		attrs = append(attrs, computeAttributeFromResult(created))
	}
	plan.ComputeAttributes = attrs

	tflog.Trace(ctx, "created computeprofile", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *computeprofileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state computeprofileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	result, err := r.client.ReadForemanComputeProfile(ctx, id)
	if err != nil {
		if goforeman.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read computeprofile, got error: %s", err))
		return
	}

	state.Name = types.StringValue(result.Name)
	attrs := make([]computeAttributeModel, 0, len(result.ComputeAttributes))
	for _, ca := range result.ComputeAttributes {
		attrs = append(attrs, computeAttributeFromResult(ca))
	}
	state.ComputeAttributes = attrs

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *computeprofileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan computeprofileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state computeprofileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	body := &goforeman.ForemanComputeProfileRequest{Name: plan.Name.ValueString()}
	result, err := r.client.UpdateForemanComputeProfile(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update computeprofile, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(result.Name)

	// Reconcile compute_attributes against prior state, matched by
	// compute_resource_id (Foreman treats that as the effective key: one
	// attribute set per compute resource per profile).
	byComputeResource := make(map[int64]computeAttributeModel, len(state.ComputeAttributes))
	for _, ca := range state.ComputeAttributes {
		byComputeResource[ca.ComputeResourceID.ValueInt64()] = ca
	}
	seen := make(map[int64]bool, len(plan.ComputeAttributes))

	attrs := make([]computeAttributeModel, 0, len(plan.ComputeAttributes))
	for _, ca := range plan.ComputeAttributes {
		crID := ca.ComputeResourceID.ValueInt64()
		seen[crID] = true
		if existing, ok := byComputeResource[crID]; ok {
			updated, err := r.client.UpdateForemanComputeAttribute(ctx, id, int(crID), int(existing.ID.ValueInt64()), vmAttrsToRaw(ca.VMAttrs))
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update compute attribute for compute_resource_id %d, got error: %s", crID, err))
				return
			}
			attrs = append(attrs, computeAttributeFromResult(updated))
		} else {
			created, err := r.client.CreateForemanComputeAttribute(ctx, id, int(crID), vmAttrsToRaw(ca.VMAttrs))
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create compute attribute for compute_resource_id %d, got error: %s", crID, err))
				return
			}
			attrs = append(attrs, computeAttributeFromResult(created))
		}
	}
	for crID, existing := range byComputeResource {
		if seen[crID] {
			continue
		}
		if err := r.client.DeleteForemanComputeAttribute(ctx, id, int(crID), int(existing.ID.ValueInt64())); err != nil && !goforeman.IsNotFoundError(err) {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete compute attribute for compute_resource_id %d, got error: %s", crID, err))
			return
		}
	}
	plan.ComputeAttributes = attrs

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *computeprofileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state computeprofileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteForemanComputeProfile(ctx, id)
	if err != nil && !goforeman.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete computeprofile, got error: %s", err))
		return
	}
}

func (r *computeprofileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

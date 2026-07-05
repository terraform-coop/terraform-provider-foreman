package provider

import (
	"context"
	"encoding/json"
	"errors"
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

func NewComputeProfileResource() resource.Resource {
	return &computeprofileResource{}
}

type computeprofileResource struct {
	client *goforeman.Client
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
	client, ok := req.ProviderData.(*goforeman.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *goforeman.Client")
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

func computeAttributeFromResult(ca *goforeman.ComputeAttribute) computeAttributeModel {
	return computeAttributeModel{
		ID:                types.Int64Value(int64(ca.ID)),
		ComputeResourceID: types.Int64Value(int64(ca.ComputeResourceID)),
		VMAttrs:           vmAttrsFromRaw(ca.VMAttrs),
	}
}

func desiredComputeAttributes(models []computeAttributeModel) []goforeman.ComputeAttribute {
	out := make([]goforeman.ComputeAttribute, 0, len(models))
	for _, ca := range models {
		out = append(out, goforeman.ComputeAttribute{
			ComputeResourceID: int(ca.ComputeResourceID.ValueInt64()),
			VMAttrs:           vmAttrsToRaw(ca.VMAttrs),
		})
	}
	return out
}

func computeAttributesFromResults(results []*goforeman.ComputeAttribute) []computeAttributeModel {
	out := make([]computeAttributeModel, 0, len(results))
	for _, ca := range results {
		out = append(out, computeAttributeFromResult(ca))
	}
	return out
}

func (r *computeprofileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan computeprofileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &goforeman.ComputeProfileRequest{Name: plan.Name.ValueString()}

	result, err := r.client.CreateComputeProfile(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create computeprofile, got error: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	plan.Name = types.StringValue(result.Name)

	synced, err := r.client.SyncComputeAttributes(ctx, int(result.ID), desiredComputeAttributes(plan.ComputeAttributes))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to sync compute attributes, got error: %s", err))
		return
	}
	plan.ComputeAttributes = computeAttributesFromResults(synced)

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

	result, err := r.client.ReadComputeProfile(ctx, id)
	if err != nil {
		if errors.Is(err, goforeman.ErrNotFound) {
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

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	body := &goforeman.ComputeProfileRequest{Name: plan.Name.ValueString()}
	result, err := r.client.UpdateComputeProfile(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update computeprofile, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(result.Name)

	synced, err := r.client.SyncComputeAttributes(ctx, id, desiredComputeAttributes(plan.ComputeAttributes))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to sync compute attributes, got error: %s", err))
		return
	}
	plan.ComputeAttributes = computeAttributesFromResults(synced)

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

	err = r.client.DeleteComputeProfile(ctx, id)
	if err != nil && !errors.Is(err, goforeman.ErrNotFound) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete computeprofile, got error: %s", err))
		return
	}
}

func (r *computeprofileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

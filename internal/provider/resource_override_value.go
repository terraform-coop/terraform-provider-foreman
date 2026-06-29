// Hand-written override_value resource (nested endpoint, custom marshaling).

package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/terraform-coop/terraform-provider-foreman/generated"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &overrideValueResource{}
	_ resource.ResourceWithImportState = &overrideValueResource{}
)

func NewForemanOverrideValueResource() resource.Resource {
	return &overrideValueResource{}
}

type overrideValueResource struct {
	client *generated.ForemanClient
}

type overrideValueResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	SmartClassParameterID types.Int64  `tfsdk:"smart_class_parameter_id"`
	MatchType             types.String `tfsdk:"match_type"`
	MatchValue            types.String `tfsdk:"match_value"`
	Value                 types.String `tfsdk:"value"`
	Omit                  types.Bool   `tfsdk:"omit"`
}

func (r *overrideValueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_override_value"
}

func (r *overrideValueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"smart_class_parameter_id": schema.Int64Attribute{
				Required:    true,
				Description: "ID of the smart class parameter to override.",
			},
			"match_type": schema.StringAttribute{
				Required:    true,
				Description: "The type of match: fqdn, hostgroup, domain, or os.",
			},
			"match_value": schema.StringAttribute{
				Required:    true,
				Description: "The value of the match (e.g. hostname, hostgroup name).",
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "The override value. Hashes and arrays must be JSON encoded.",
			},
			"omit": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, Foreman will not send this parameter in classification output.",
			},
		},
	}
}

func (r *overrideValueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *overrideValueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan overrideValueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scpID := int(plan.SmartClassParameterID.ValueInt64())
	matchStr := plan.MatchType.ValueString() + "=" + plan.MatchValue.ValueString()

	body := &generated.ForemanOverrideValueRequest{
		Match: matchStr,
		Value: plan.Value.ValueString(),
		Omit:  plan.Omit.ValueBool(),
	}

	result, err := r.client.CreateForemanOverrideValue(ctx, scpID, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create override_value, got error: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(int(result.ID)) + "@" + strconv.Itoa(scpID))
	plan.Omit = types.BoolValue(result.Omit)

	tflog.Trace(ctx, "created override_value", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *overrideValueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state overrideValueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, scpID, err := parseOverrideValueID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	result, err := r.client.ReadForemanOverrideValue(ctx, scpID, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read override_value, got error: %s", err))
		return
	}

	state.Omit = types.BoolValue(result.Omit)
	state.MatchType = types.StringValue(result.MatchType)
	state.MatchValue = types.StringValue(result.MatchValue)
	state.Value = types.StringValue(result.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *overrideValueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan overrideValueResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, scpID, err := parseOverrideValueID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	matchStr := plan.MatchType.ValueString() + "=" + plan.MatchValue.ValueString()
	body := &generated.ForemanOverrideValueRequest{
		Match: matchStr,
		Value: plan.Value.ValueString(),
		Omit:  plan.Omit.ValueBool(),
	}

	result, err := r.client.UpdateForemanOverrideValue(ctx, scpID, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update override_value, got error: %s", err))
		return
	}

	plan.Omit = types.BoolValue(result.Omit)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *overrideValueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state overrideValueResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, scpID, err := parseOverrideValueID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteForemanOverrideValue(ctx, scpID, id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete override_value, got error: %s", err))
		return
	}
}

func (r *overrideValueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// parseOverrideValueID parses "id@scp_id" format.
func parseOverrideValueID(raw string) (int, int, error) {
	parts := strings.SplitN(raw, "@", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected 'id@scp_id' format, got %q", raw)
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid id: %w", err)
	}
	scpID, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid scp_id: %w", err)
	}
	return id, scpID, nil
}

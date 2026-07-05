package provider

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &settingResource{}
	_ resource.ResourceWithImportState = &settingResource{}
)

func NewForemanSettingResource() resource.Resource {
	return &settingResource{}
}

type settingResource struct {
	client *goforeman.ForemanClient
}

type settingResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Value types.String `tfsdk:"value"`
}

func (r *settingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_setting"
}

func (r *settingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"value": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
		},
		Description:         "Foreman does not support creating or deleting setting entries via the API; only its existing attributes can be updated. Use `terraform import` to bring an existing one under management.",
		MarkdownDescription: "Foreman does not support creating or deleting setting entries via the API; only its existing attributes can be updated. Use `terraform import` to bring an existing one under management.",
	}
}

func (r *settingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *settingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Warn(ctx, "Create is not supported for setting")
	resp.Diagnostics.AddError("Not Supported", "Foreman does not support creating setting entries via the API; import an existing one instead (see this resource's Import Statement docs).")
}

func (r *settingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state settingResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.ReadForemanSetting(ctx, state.ID.ValueString())
	if err != nil {
		if goforeman.IsNotFoundError(err) {
			tflog.Warn(ctx, "setting not found, removing from state", map[string]interface{}{"id": state.ID.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read setting, got error: %s", err))
		return
	}

	state.Value = types.StringValue(parameterValueToString(result.Value))

	tflog.Trace(ctx, "read setting", map[string]interface{}{"id": state.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *settingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan settingResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &goforeman.ForemanSettingRequest{Value: plan.Value.ValueString()}

	result, err := r.client.UpdateForemanSetting(ctx, plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update setting, got error: %s", err))
		return
	}

	plan.Value = types.StringValue(parameterValueToString(result.Value))

	tflog.Trace(ctx, "updated setting", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *settingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Warn(ctx, "Delete is not supported for setting")
	resp.Diagnostics.AddError("Not Supported", "Foreman does not support deleting setting entries via the API; run 'terraform state rm' to stop managing it instead.")
}

func (r *settingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

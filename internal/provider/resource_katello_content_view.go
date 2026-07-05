package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/terraform-coop/terraform-provider-foreman/generated"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                = &katelloContentViewResource{}
	_ resource.ResourceWithImportState = &katelloContentViewResource{}
)

func NewKatelloContentViewResource() resource.Resource {
	return &katelloContentViewResource{}
}

type katelloContentViewResource struct {
	client *generated.ForemanClient
}

type katelloContentViewResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Label             types.String `tfsdk:"label"`
	OrganizationID    types.Int64  `tfsdk:"organization_id"`
	Composite         types.Bool   `tfsdk:"composite"`
	AutoPublish       types.Bool   `tfsdk:"auto_publish"`
	SolveDependencies types.Bool   `tfsdk:"solve_dependencies"`
	Filtered          types.Bool   `tfsdk:"filtered"`
	Filters           types.List   `tfsdk:"filters"`
	LatestVersionID   types.Int64  `tfsdk:"latest_version_id"`
	LatestVersion     types.String `tfsdk:"latest_version"`
	ContentHostCount  types.Int64  `tfsdk:"content_host_count"`
	VersionCount      types.Int64  `tfsdk:"version_count"`
	RepositoryIDs     types.List   `tfsdk:"repository_ids"`
	ComponentIDs      types.List   `tfsdk:"component_ids"`
}

func (r *katelloContentViewResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_katello_content_view"
}

func (r *katelloContentViewResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"composite": schema.BoolAttribute{
				Optional: true,
			},
			"auto_publish": schema.BoolAttribute{
				Optional: true,
			},
			"solve_dependencies": schema.BoolAttribute{
				Optional: true,
			},
			"filtered": schema.BoolAttribute{
				Optional: true,
			},
			"filters": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Content view filters. Each filter can include rules for package, package group, erratum, or docker manifest filtering.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Filter type: package, package_group, erratum, docker_manifest, or deb.",
						},
						"inclusion": schema.BoolAttribute{
							Optional:    true,
							Description: "If true, the filter includes matching content. If false, it excludes matching content. Defaults to false.",
						},
						"description": schema.StringAttribute{
							Optional: true,
						},
						"rule": schema.ListNestedAttribute{
							Optional:    true,
							Description: "Rules for this filter.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.Int64Attribute{
										Computed: true,
									},
									"name": schema.StringAttribute{
										Required: true,
									},
									"architecture": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"latest_version_id": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"latest_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"content_host_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"version_count": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"repository_ids": schema.ListAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
			},
			"component_ids": schema.ListAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
			},
		},
	}
}

func (r *katelloContentViewResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *katelloContentViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan katelloContentViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &generated.ForemanKatelloContentViewRequest{
		Name:              plan.Name.ValueString(),
		Description:       plan.Description.ValueString(),
		Label:             plan.Label.ValueString(),
		OrganizationID:    int(plan.OrganizationID.ValueInt64()),
		Composite:         boolPointerOrNil(plan.Composite),
		AutoPublish:       boolPointerOrNil(plan.AutoPublish),
		SolveDependencies: boolPointerOrNil(plan.SolveDependencies),
		Filtered:          boolPointerOrNil(plan.Filtered),
	}
	if !plan.RepositoryIDs.IsUnknown() {
		var repoIDs []int64
		diags := plan.RepositoryIDs.ElementsAs(ctx, &repoIDs, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		for _, id := range repoIDs {
			body.RepositoryIDs = append(body.RepositoryIDs, int(id))
		}
	}
	if !plan.ComponentIDs.IsUnknown() {
		var compIDs []int64
		diags := plan.ComponentIDs.ElementsAs(ctx, &compIDs, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		for _, id := range compIDs {
			body.ComponentIDs = append(body.ComponentIDs, int(id))
		}
	}

	result, err := r.client.CreateForemanKatelloContentView(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create katello content view, got error: %s", err))
		return
	}

	// Publish an initial version after creation (matches old provider behavior)
	published, err := r.client.PublishContentView(ctx, int(result.ID))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to publish initial content view version, got error: %s", err))
		return
	}
	if published != nil {
		result = published
	}

	plan.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	plan.Name = types.StringValue(result.Name)
	plan.Description = types.StringValue(result.Description)
	plan.Label = types.StringValue(result.Label)
	plan.OrganizationID = types.Int64Value(int64(result.OrganizationID))
	plan.Composite = types.BoolValue(result.Composite)
	plan.AutoPublish = types.BoolValue(result.AutoPublish)
	plan.SolveDependencies = types.BoolValue(result.SolveDependencies)
	plan.Filtered = types.BoolValue(result.Filtered)
	plan.LatestVersionID = types.Int64Value(int64(result.LatestVersionID))
	plan.LatestVersion = types.StringValue(result.LatestVersion)
	plan.ContentHostCount = types.Int64Value(int64(result.ContentHostCount))
	plan.VersionCount = types.Int64Value(int64(result.VersionCount))
	var diags diag.Diagnostics
	repoIDVals := make([]attr.Value, len(result.RepositoryIDs))
	for i, v := range result.RepositoryIDs {
		repoIDVals[i] = types.Int64Value(int64(v))
	}
	plan.RepositoryIDs, diags = types.ListValue(types.Int64Type, repoIDVals)
	resp.Diagnostics.Append(diags...)
	compIDVals := make([]attr.Value, len(result.ComponentIDs))
	for i, v := range result.ComponentIDs {
		compIDVals[i] = types.Int64Value(int64(v))
	}
	plan.ComponentIDs, diags = types.ListValue(types.Int64Type, compIDVals)
	resp.Diagnostics.Append(diags...)

	// Sync filters if provided
	if !plan.Filters.IsNull() && !plan.Filters.IsUnknown() {
		filters, filterDiags := expandContentViewFilters(ctx, plan.Filters)
		if filterDiags.HasError() {
			resp.Diagnostics.Append(filterDiags...)
			return
		}
		if err := r.client.SyncContentViewFilters(ctx, int(result.ID), filters); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to sync content view filters, got error: %s", err))
			return
		}
		// Re-read filters to get computed IDs
		readFilters, err := r.client.ReadContentViewFilters(ctx, int(result.ID))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read content view filters after sync, got error: %s", err))
			return
		}
		plan.Filters, diags = flattenContentViewFilters(ctx, readFilters)
		resp.Diagnostics.Append(diags...)
	} else {
		plan.Filters = types.ListNull(filterObjType)
	}

	tflog.Trace(ctx, "created katello content view", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *katelloContentViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state katelloContentViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	result, err := r.client.ReadForemanKatelloContentView(ctx, id)
	if err != nil {
		if generated.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read katello content view, got error: %s", err))
		return
	}
	var diags diag.Diagnostics
	state.Name = types.StringValue(result.Name)
	state.Description = types.StringValue(result.Description)
	state.Label = types.StringValue(result.Label)
	state.OrganizationID = types.Int64Value(int64(result.OrganizationID))
	state.Composite = types.BoolValue(result.Composite)
	state.AutoPublish = types.BoolValue(result.AutoPublish)
	state.SolveDependencies = types.BoolValue(result.SolveDependencies)
	state.Filtered = types.BoolValue(result.Filtered)
	state.LatestVersionID = types.Int64Value(int64(result.LatestVersionID))
	state.LatestVersion = types.StringValue(result.LatestVersion)
	state.ContentHostCount = types.Int64Value(int64(result.ContentHostCount))
	state.VersionCount = types.Int64Value(int64(result.VersionCount))
	repoIDVals := make([]attr.Value, len(result.RepositoryIDs))
	for i, v := range result.RepositoryIDs {
		repoIDVals[i] = types.Int64Value(int64(v))
	}
	state.RepositoryIDs, diags = types.ListValue(types.Int64Type, repoIDVals)
	resp.Diagnostics.Append(diags...)
	compIDVals := make([]attr.Value, len(result.ComponentIDs))
	for i, v := range result.ComponentIDs {
		compIDVals[i] = types.Int64Value(int64(v))
	}
	state.ComponentIDs, diags = types.ListValue(types.Int64Type, compIDVals)
	resp.Diagnostics.Append(diags...)

	// Read filters
	readFilters, err := r.client.ReadContentViewFilters(ctx, id)
	if err != nil {
		tflog.Warn(ctx, "Unable to read content view filters", map[string]interface{}{"error": err.Error()})
		state.Filters = types.ListNull(filterObjType)
	} else {
		state.Filters, diags = flattenContentViewFilters(ctx, readFilters)
		resp.Diagnostics.Append(diags...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *katelloContentViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan katelloContentViewResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	body := &generated.ForemanKatelloContentViewRequest{
		Name:              plan.Name.ValueString(),
		Description:       plan.Description.ValueString(),
		Label:             plan.Label.ValueString(),
		OrganizationID:    int(plan.OrganizationID.ValueInt64()),
		Composite:         boolPointerOrNil(plan.Composite),
		AutoPublish:       boolPointerOrNil(plan.AutoPublish),
		SolveDependencies: boolPointerOrNil(plan.SolveDependencies),
		Filtered:          boolPointerOrNil(plan.Filtered),
	}
	if !plan.RepositoryIDs.IsUnknown() {
		var repoIDs []int64
		diags := plan.RepositoryIDs.ElementsAs(ctx, &repoIDs, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		for _, id := range repoIDs {
			body.RepositoryIDs = append(body.RepositoryIDs, int(id))
		}
	}
	if !plan.ComponentIDs.IsUnknown() {
		var compIDs []int64
		diags := plan.ComponentIDs.ElementsAs(ctx, &compIDs, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		for _, id := range compIDs {
			body.ComponentIDs = append(body.ComponentIDs, int(id))
		}
	}

	result, err := r.client.UpdateForemanKatelloContentView(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update katello content view, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(result.Name)
	plan.Description = types.StringValue(result.Description)
	plan.Label = types.StringValue(result.Label)
	plan.OrganizationID = types.Int64Value(int64(result.OrganizationID))
	plan.Composite = types.BoolValue(result.Composite)
	plan.AutoPublish = types.BoolValue(result.AutoPublish)
	plan.SolveDependencies = types.BoolValue(result.SolveDependencies)
	plan.Filtered = types.BoolValue(result.Filtered)
	plan.LatestVersionID = types.Int64Value(int64(result.LatestVersionID))
	plan.LatestVersion = types.StringValue(result.LatestVersion)
	plan.ContentHostCount = types.Int64Value(int64(result.ContentHostCount))
	plan.VersionCount = types.Int64Value(int64(result.VersionCount))
	var diags diag.Diagnostics
	repoIDVals := make([]attr.Value, len(result.RepositoryIDs))
	for i, v := range result.RepositoryIDs {
		repoIDVals[i] = types.Int64Value(int64(v))
	}
	plan.RepositoryIDs, diags = types.ListValue(types.Int64Type, repoIDVals)
	resp.Diagnostics.Append(diags...)
	compIDVals := make([]attr.Value, len(result.ComponentIDs))
	for i, v := range result.ComponentIDs {
		compIDVals[i] = types.Int64Value(int64(v))
	}
	plan.ComponentIDs, diags = types.ListValue(types.Int64Type, compIDVals)
	resp.Diagnostics.Append(diags...)

	// Sync filters
	if !plan.Filters.IsNull() && !plan.Filters.IsUnknown() {
		filters, filterDiags := expandContentViewFilters(ctx, plan.Filters)
		if filterDiags.HasError() {
			resp.Diagnostics.Append(filterDiags...)
			return
		}
		if err := r.client.SyncContentViewFilters(ctx, id, filters); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to sync content view filters, got error: %s", err))
			return
		}
		// Re-read filters to get computed IDs
		readFilters, err := r.client.ReadContentViewFilters(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read content view filters after sync, got error: %s", err))
			return
		}
		plan.Filters, diags = flattenContentViewFilters(ctx, readFilters)
		resp.Diagnostics.Append(diags...)
	} else {
		// If filters removed, delete all existing
		readFilters, _ := r.client.ReadContentViewFilters(ctx, id)
		if len(readFilters) > 0 {
			var empty []generated.ForemanKatelloContentViewFilter
			if err := r.client.SyncContentViewFilters(ctx, id, empty); err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete content view filters, got error: %s", err))
				return
			}
		}
		plan.Filters = types.ListNull(filterObjType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *katelloContentViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state katelloContentViewResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteForemanKatelloContentView(ctx, id)
	if err != nil && !generated.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete katello content view, got error: %s", err))
		return
	}
}

func (r *katelloContentViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ---------------------------------------------------------------------------
// Helper functions for content view filters
// ---------------------------------------------------------------------------

var filterRuleAttrTypes = map[string]attr.Type{
	"id":           types.Int64Type,
	"name":         types.StringType,
	"architecture": types.StringType,
}

var filterRuleObjType = types.ObjectType{AttrTypes: filterRuleAttrTypes}

var filterAttrTypeMap = map[string]attr.Type{
	"id":          types.Int64Type,
	"name":        types.StringType,
	"type":        types.StringType,
	"inclusion":   types.BoolType,
	"description": types.StringType,
	"rule":        types.ListType{ElemType: filterRuleObjType},
}

var filterObjType = types.ObjectType{AttrTypes: filterAttrTypeMap}

func flattenContentViewFilters(ctx context.Context, filters []generated.ForemanKatelloContentViewFilter) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(filters) == 0 {
		return types.ListNull(filterObjType), diags
	}

	elements := make([]attr.Value, 0, len(filters))
	for _, f := range filters {
		ruleElements := make([]attr.Value, 0, len(f.Rules))
		for _, r := range f.Rules {
			ruleObj, d := types.ObjectValue(filterRuleAttrTypes, map[string]attr.Value{
				"id":           types.Int64Value(int64(r.ID)),
				"name":         types.StringValue(r.Name),
				"architecture": types.StringValue(r.Architecture),
			})
			diags.Append(d...)
			ruleElements = append(ruleElements, ruleObj)
		}

		rulesList, d := types.ListValue(types.ObjectType{AttrTypes: filterRuleAttrTypes}, ruleElements)
		diags.Append(d...)

		filterObj, d := types.ObjectValue(filterAttrTypeMap, map[string]attr.Value{
			"id":          types.Int64Value(int64(f.ID)),
			"name":        types.StringValue(f.Name),
			"type":        types.StringValue(f.Type),
			"inclusion":   types.BoolValue(f.Inclusion),
			"description": types.StringValue(f.Description),
			"rule":        rulesList,
		})
		diags.Append(d...)
		elements = append(elements, filterObj)
	}

	result, d := types.ListValue(filterObjType, elements)
	diags.Append(d...)
	return result, diags
}

func expandContentViewFilters(ctx context.Context, filtersList types.List) ([]generated.ForemanKatelloContentViewFilter, diag.Diagnostics) {
	var diags diag.Diagnostics
	if filtersList.IsNull() || filtersList.IsUnknown() {
		return nil, diags
	}

	elements := make([]types.Object, 0, len(filtersList.Elements()))
	diags.Append(filtersList.ElementsAs(ctx, &elements, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]generated.ForemanKatelloContentViewFilter, 0, len(elements))
	for _, elem := range elements {
		attrs := elem.Attributes()

		var f generated.ForemanKatelloContentViewFilter
		if v, ok := attrs["id"]; ok && !v.(types.Int64).IsNull() && !v.(types.Int64).IsUnknown() {
			f.ID = int(v.(types.Int64).ValueInt64())
		}
		if v, ok := attrs["name"]; ok && !v.(types.String).IsNull() && !v.(types.String).IsUnknown() {
			f.Name = v.(types.String).ValueString()
		}
		if v, ok := attrs["type"]; ok && !v.(types.String).IsNull() && !v.(types.String).IsUnknown() {
			f.Type = v.(types.String).ValueString()
		}
		if v, ok := attrs["inclusion"]; ok && !v.(types.Bool).IsNull() && !v.(types.Bool).IsUnknown() {
			f.Inclusion = v.(types.Bool).ValueBool()
		}
		if v, ok := attrs["description"]; ok && !v.(types.String).IsNull() && !v.(types.String).IsUnknown() {
			f.Description = v.(types.String).ValueString()
		}

		// Expand rules
		if v, ok := attrs["rule"]; ok && !v.(types.List).IsNull() && !v.(types.List).IsUnknown() {
			ruleElements := make([]types.Object, 0, len(v.(types.List).Elements()))
			diags.Append(v.(types.List).ElementsAs(ctx, &ruleElements, false)...)
			if diags.HasError() {
				return nil, diags
			}
			for _, re := range ruleElements {
				rAttrs := re.Attributes()
				var rule generated.ForemanKatelloContentViewFilterRule
				if rv, ok := rAttrs["id"]; ok && !rv.(types.Int64).IsNull() && !rv.(types.Int64).IsUnknown() {
					rule.ID = int(rv.(types.Int64).ValueInt64())
				}
				if rv, ok := rAttrs["name"]; ok && !rv.(types.String).IsNull() && !rv.(types.String).IsUnknown() {
					rule.Name = rv.(types.String).ValueString()
				}
				if rv, ok := rAttrs["architecture"]; ok && !rv.(types.String).IsNull() && !rv.(types.String).IsUnknown() {
					rule.Architecture = rv.(types.String).ValueString()
				}
				f.Rules = append(f.Rules, rule)
			}
		}

		result = append(result, f)
	}

	return result, diags
}

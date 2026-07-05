package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &parameterDataSource{}

func NewForemanParameterDataSource() datasource.DataSource {
	return &parameterDataSource{}
}

type parameterDataSource struct {
	client *goforeman.ForemanClient
}

type parameterDataSourceModel struct {
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

func (d *parameterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_parameter"
}

func (d *parameterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a parameter scoped to exactly one of host_id, hostgroup_id, " +
			"domain_id, operatingsystem_id, subnet_id, location_id, or organization_id - " +
			"set exactly one of these.",
		Attributes: map[string]schema.Attribute{
			"id":                 schema.StringAttribute{Computed: true},
			"host_id":            schema.Int64Attribute{Optional: true},
			"hostgroup_id":       schema.Int64Attribute{Optional: true},
			"domain_id":          schema.Int64Attribute{Optional: true},
			"operatingsystem_id": schema.Int64Attribute{Optional: true},
			"subnet_id":          schema.Int64Attribute{Optional: true},
			"location_id":        schema.Int64Attribute{Optional: true},
			"organization_id":    schema.Int64Attribute{Optional: true},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the parameter to look up.",
			},
			"value":          schema.StringAttribute{Computed: true},
			"parameter_type": schema.StringAttribute{Computed: true},
			"hidden_value":   schema.BoolAttribute{Computed: true},
		},
	}
}

func (d *parameterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*goforeman.ForemanClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *goforeman.ForemanClient")
		return
	}
	d.client = client
}

func (d *parameterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data parameterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentType, parentID, diags := parameterParent(&parameterResourceModel{
		HostID: data.HostID, HostgroupID: data.HostgroupID, DomainID: data.DomainID,
		OperatingsystemID: data.OperatingsystemID, SubnetID: data.SubnetID,
		LocationID: data.LocationID, OrganizationID: data.OrganizationID,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	result, err := d.client.QueryForemanParameter(ctx, parentType, int(parentID), name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read parameter, got error: %s", err))
		return
	}
	if result == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("Parameter %q not found", name))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	data.Value = types.StringValue(parameterValueToString(result.Value))
	data.ParameterType = types.StringValue(result.ParameterType)
	data.HiddenValue = types.BoolValue(result.HiddenValue)

	tflog.Trace(ctx, "read parameter data source", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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

var (
	_ datasource.DataSource = &katelloLifecycleEnvironmentDataSource{}
)

func NewKatelloLifecycleEnvironmentDataSource() datasource.DataSource {
	return &katelloLifecycleEnvironmentDataSource{}
}

type katelloLifecycleEnvironmentDataSource struct {
	client *goforeman.Client
}

type katelloLifecycleEnvironmentDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Label          types.String `tfsdk:"label"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Library        types.Bool   `tfsdk:"library"`
}

func (d *katelloLifecycleEnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_katello_lifecycle_environment"
}

func (d *katelloLifecycleEnvironmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the katello lifecycle environment to look up.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the lifecycle environment",
			},
			"label": schema.StringAttribute{
				Computed:    true,
				Description: "Label of the lifecycle environment",
			},
			"organization_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Organization ID",
			},
			"library": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this is the library lifecycle environment",
			},
		},
	}
}

func (d *katelloLifecycleEnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*goforeman.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *goforeman.Client")
		return
	}
	d.client = client
}

func (d *katelloLifecycleEnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data katelloLifecycleEnvironmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	result, err := d.client.FindKatelloLifecycleEnvironmentByName(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read katello lifecycle environment, got error: %s", err))
		return
	}
	if result == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("KatelloLifecycleEnvironment %q not found", name))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	data.Description = types.StringValue(result.Description)
	data.Label = types.StringValue(result.Label)
	data.OrganizationID = types.Int64Value(int64(result.OrganizationID))
	data.Library = types.BoolValue(result.Library)

	tflog.Trace(ctx, "read katello lifecycle environment data source", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

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
	_ datasource.DataSource = &katelloSyncPlanDataSource{}
)

func NewKatelloSyncPlanDataSource() datasource.DataSource {
	return &katelloSyncPlanDataSource{}
}

type katelloSyncPlanDataSource struct {
	client *goforeman.ForemanClient
}

type katelloSyncPlanDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Interval       types.String `tfsdk:"interval"`
	SyncDate       types.String `tfsdk:"sync_date"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	CronExpression types.String `tfsdk:"cron_expression"`
}

func (d *katelloSyncPlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_katello_sync_plan"
}

func (d *katelloSyncPlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the katello sync plan to look up.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the sync plan",
			},
			"interval": schema.StringAttribute{
				Computed:    true,
				Description: "Sync interval",
			},
			"sync_date": schema.StringAttribute{
				Computed:    true,
				Description: "Sync date",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the sync plan is enabled",
			},
			"cron_expression": schema.StringAttribute{
				Computed:    true,
				Description: "Cron expression",
			},
		},
	}
}

func (d *katelloSyncPlanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *katelloSyncPlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data katelloSyncPlanDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	result, err := d.client.QueryForemanKatelloSyncPlan(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read katello sync plan, got error: %s", err))
		return
	}
	if result == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("ForemanKatelloSyncPlan %q not found", name))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	data.Description = types.StringValue(result.Description)
	data.Interval = types.StringValue(result.Interval)
	data.SyncDate = types.StringValue(result.SyncDate)
	data.Enabled = types.BoolValue(result.Enabled)
	data.CronExpression = types.StringValue(result.CronExpression)

	tflog.Trace(ctx, "read katello sync plan data source", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/terraform-coop/terraform-provider-foreman/generated"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource = &katelloProductDataSource{}
)

func NewKatelloProductDataSource() datasource.DataSource {
	return &katelloProductDataSource{}
}

type katelloProductDataSource struct {
	client *generated.ForemanClient
}

type katelloProductDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Label           types.String `tfsdk:"label"`
	GpgKeyID        types.Int64  `tfsdk:"gpg_key_id"`
	SslCaCertID     types.Int64  `tfsdk:"ssl_ca_cert_id"`
	SslClientCertID types.Int64  `tfsdk:"ssl_client_cert_id"`
	SslClientKeyID  types.Int64  `tfsdk:"ssl_client_key_id"`
	SyncPlanID      types.Int64  `tfsdk:"sync_plan_id"`
}

func (d *katelloProductDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_katello_product"
}

func (d *katelloProductDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the katello product to look up.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the product",
			},
			"label": schema.StringAttribute{
				Computed:    true,
				Description: "Label of the product",
			},
			"gpg_key_id": schema.Int64Attribute{
				Computed:    true,
				Description: "GPG key ID",
			},
			"ssl_ca_cert_id": schema.Int64Attribute{
				Computed:    true,
				Description: "SSL CA cert ID",
			},
			"ssl_client_cert_id": schema.Int64Attribute{
				Computed:    true,
				Description: "SSL client cert ID",
			},
			"ssl_client_key_id": schema.Int64Attribute{
				Computed:    true,
				Description: "SSL client key ID",
			},
			"sync_plan_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Sync plan ID",
			},
		},
	}
}

func (d *katelloProductDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*generated.ForemanClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *generated.ForemanClient")
		return
	}
	d.client = client
}

func (d *katelloProductDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data katelloProductDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	result, err := d.client.QueryForemanKatelloProduct(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read katello product, got error: %s", err))
		return
	}
	if result == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("ForemanKatelloProduct %q not found", name))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	data.Description = types.StringValue(result.Description)
	data.Label = types.StringValue(result.Label)
	data.GpgKeyID = types.Int64Value(int64(result.GpgKeyID))
	data.SslCaCertID = types.Int64Value(int64(result.SslCaCertID))
	data.SslClientCertID = types.Int64Value(int64(result.SslClientCertID))
	data.SslClientKeyID = types.Int64Value(int64(result.SslClientKeyID))
	data.SyncPlanID = types.Int64Value(int64(result.SyncPlanID))

	tflog.Trace(ctx, "read katello product data source", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"
)

// Hand-written: see resource_autosign.go's header comment for why autosign
// doesn't fit the generic apidoc-driven pipeline.

var _ datasource.DataSource = &autosignDataSource{}

func NewAutosignDataSource() datasource.DataSource {
	return &autosignDataSource{}
}

type autosignDataSource struct {
	client *goforeman.Client
}

type autosignDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	SmartProxyID types.Int64  `tfsdk:"smart_proxy_id"`
}

func (d *autosignDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_autosign"
}

func (d *autosignDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The hostname or wildcard pattern to look up (e.g. \"host.example.com\" or \"*.example.com\").",
			Required:    true,
		},
		"smart_proxy_id": schema.Int64Attribute{
			Description: "ID of the smart proxy this autosign entry applies to.",
			Required:    true,
		},
	}}
}

func (d *autosignDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *autosignDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data autosignDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pattern := data.ID.ValueString()
	smartProxyID := int(data.SmartProxyID.ValueInt64())
	result, err := d.client.ReadAutosign(ctx, smartProxyID, pattern)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read autosign entry, got error: %s", err))
		return
	}
	if result == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("autosign entry %q not found on smart proxy %d", pattern, smartProxyID))
		return
	}

	data.ID = types.StringValue(result.ID)

	tflog.Trace(ctx, "read autosign data source", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

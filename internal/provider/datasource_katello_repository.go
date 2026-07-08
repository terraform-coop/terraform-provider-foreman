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
	_ datasource.DataSource = &katelloRepositoryDataSource{}
)

func NewKatelloRepositoryDataSource() datasource.DataSource {
	return &katelloRepositoryDataSource{}
}

type katelloRepositoryDataSource struct {
	client *goforeman.Client
}

type katelloRepositoryDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Label               types.String `tfsdk:"label"`
	ProductID           types.Int64  `tfsdk:"product_id"`
	ContentType         types.String `tfsdk:"content_type"`
	URL                 types.String `tfsdk:"url"`
	GpgKeyID            types.Int64  `tfsdk:"gpg_key_id"`
	Unprotected         types.Bool   `tfsdk:"unprotected"`
	ChecksumType        types.String `tfsdk:"checksum_type"`
	DownloadPolicy      types.String `tfsdk:"download_policy"`
	DownloadConcurrency types.Int64  `tfsdk:"download_concurrency"`
	MirrorOnSync        types.Bool   `tfsdk:"mirror_on_sync"`
	MirroringPolicy     types.String `tfsdk:"mirroring_policy"`
	HttpProxyPolicy     types.String `tfsdk:"http_proxy_policy"`
	HttpProxyID         types.Int64  `tfsdk:"http_proxy_id"`
}

func (d *katelloRepositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_katello_repository"
}

func (d *katelloRepositoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the katello repository to look up.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the repository",
			},
			"label": schema.StringAttribute{
				Computed:    true,
				Description: "Label of the repository",
			},
			"product_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Product ID",
			},
			"content_type": schema.StringAttribute{
				Computed:    true,
				Description: "Content type of the repository",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "URL of the repository",
			},
			"gpg_key_id": schema.Int64Attribute{
				Computed:    true,
				Description: "GPG key ID",
			},
			"unprotected": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the repository is unprotected",
			},
			"checksum_type": schema.StringAttribute{
				Computed:    true,
				Description: "Checksum type",
			},
			"download_policy": schema.StringAttribute{
				Computed:    true,
				Description: "Download policy",
			},
			"download_concurrency": schema.Int64Attribute{
				Computed:    true,
				Description: "Download concurrency",
			},
			"mirror_on_sync": schema.BoolAttribute{
				Computed:    true,
				Description: "Mirror on sync",
			},
			"mirroring_policy": schema.StringAttribute{
				Computed:    true,
				Description: "Mirroring policy",
			},
			"http_proxy_policy": schema.StringAttribute{
				Computed:    true,
				Description: "HTTP proxy policy",
			},
			"http_proxy_id": schema.Int64Attribute{
				Computed:    true,
				Description: "HTTP proxy ID",
			},
		},
	}
}

func (d *katelloRepositoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *katelloRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data katelloRepositoryDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	result, err := d.client.FindKatelloRepositoryByName(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read katello repository, got error: %s", err))
		return
	}
	if result == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("KatelloRepository %q not found", name))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	data.Description = types.StringValue(result.Description)
	data.Label = types.StringValue(result.Label)
	data.ProductID = types.Int64Value(int64(result.ProductID))
	data.ContentType = types.StringValue(result.ContentType)
	data.URL = types.StringValue(result.URL)
	data.GpgKeyID = types.Int64Value(int64(result.GpgKeyID))
	data.Unprotected = types.BoolValue(result.Unprotected)
	data.ChecksumType = types.StringValue(result.ChecksumType)
	data.DownloadPolicy = types.StringValue(result.DownloadPolicy)
	data.DownloadConcurrency = types.Int64Value(int64(result.DownloadConcurrency))
	data.MirrorOnSync = types.BoolValue(result.MirrorOnSync)
	data.MirroringPolicy = types.StringValue(result.MirroringPolicy)
	data.HttpProxyPolicy = types.StringValue(result.HttpProxyPolicy)
	data.HttpProxyID = types.Int64Value(int64(result.HttpProxyID))

	tflog.Trace(ctx, "read katello repository data source", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

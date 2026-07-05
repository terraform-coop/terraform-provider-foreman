package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"
)

// Hand-written: see resource_compute_profile.go's header comment.

var _ datasource.DataSource = &computeprofileDataSource{}

func NewForemanComputeProfileDataSource() datasource.DataSource {
	return &computeprofileDataSource{}
}

type computeprofileDataSource struct {
	client *goforeman.ForemanClient
}

type computeprofileDataSourceModel struct {
	ID                types.String            `tfsdk:"id"`
	Name              types.String            `tfsdk:"name"`
	ComputeAttributes []computeAttributeModel `tfsdk:"compute_attributes"`
}

func (d *computeprofileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_computeprofile"
}

func (d *computeprofileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{
			Description: "The name of the computeprofile to look up.",
			Required:    true,
		},
		"compute_attributes": schema.ListNestedAttribute{
			Description: "Per-compute-resource VM sizing attributes for this profile.",
			Computed:    true,
			NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"id":                  schema.Int64Attribute{Computed: true},
				"compute_resource_id": schema.Int64Attribute{Computed: true},
				"vm_attrs":            schema.StringAttribute{Computed: true},
			}},
		},
	}}
}

func (d *computeprofileDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *computeprofileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data computeprofileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	result, err := d.client.QueryForemanComputeProfile(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read computeprofile, got error: %s", err))
		return
	}
	if result == nil {
		resp.Diagnostics.AddError("Not Found", fmt.Sprintf("ForemanComputeProfile %q not found", name))
		return
	}

	data.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	attrs := make([]computeAttributeModel, 0, len(result.ComputeAttributes))
	for _, ca := range result.ComputeAttributes {
		attrs = append(attrs, computeAttributeFromResult(ca))
	}
	data.ComputeAttributes = attrs

	tflog.Trace(ctx, "read computeprofile data source", map[string]interface{}{"id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

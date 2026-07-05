package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/terraform-coop/terraform-provider-foreman/goforeman"

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
	_ resource.Resource                = &katelloRepositoryResource{}
	_ resource.ResourceWithImportState = &katelloRepositoryResource{}
)

func NewKatelloRepositoryResource() resource.Resource {
	return &katelloRepositoryResource{}
}

type katelloRepositoryResource struct {
	client *goforeman.Client
}

type katelloRepositoryResourceModel struct {
	ID                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	Description                   types.String `tfsdk:"description"`
	Label                         types.String `tfsdk:"label"`
	ProductID                     types.Int64  `tfsdk:"product_id"`
	ContentType                   types.String `tfsdk:"content_type"`
	URL                           types.String `tfsdk:"url"`
	GpgKeyID                      types.Int64  `tfsdk:"gpg_key_id"`
	Unprotected                   types.Bool   `tfsdk:"unprotected"`
	ChecksumType                  types.String `tfsdk:"checksum_type"`
	DownloadPolicy                types.String `tfsdk:"download_policy"`
	DownloadConcurrency           types.Int64  `tfsdk:"download_concurrency"`
	MirrorOnSync                  types.Bool   `tfsdk:"mirror_on_sync"`
	MirroringPolicy               types.String `tfsdk:"mirroring_policy"`
	HttpProxyPolicy               types.String `tfsdk:"http_proxy_policy"`
	HttpProxyID                   types.Int64  `tfsdk:"http_proxy_id"`
	IgnoreGlobalProxy             types.Bool   `tfsdk:"ignore_global_proxy"`
	IgnorableContent              types.List   `tfsdk:"ignorable_content"`
	VerifySslOnSync               types.Bool   `tfsdk:"verify_ssl_on_sync"`
	UpstreamUsername              types.String `tfsdk:"upstream_username"`
	UpstreamPassword              types.String `tfsdk:"upstream_password"`
	DebReleases                   types.String `tfsdk:"deb_releases"`
	DebComponents                 types.String `tfsdk:"deb_components"`
	DebArchitectures              types.String `tfsdk:"deb_architectures"`
	DockerUpstreamName            types.String `tfsdk:"docker_upstream_name"`
	DockerTagsWhitelist           types.String `tfsdk:"docker_tags_whitelist"`
	AnsibleCollectionRequirements types.String `tfsdk:"ansible_collection_requirements"`
}

func (r *katelloRepositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_katello_repository"
}

func (r *katelloRepositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"product_id": schema.Int64Attribute{
				Optional: true,
			},
			"content_type": schema.StringAttribute{
				Optional: true,
			},
			"url": schema.StringAttribute{
				Optional: true,
			},
			"gpg_key_id": schema.Int64Attribute{
				Optional: true,
			},
			"unprotected": schema.BoolAttribute{
				Optional: true,
			},
			"checksum_type": schema.StringAttribute{
				Optional: true,
			},
			"download_policy": schema.StringAttribute{
				Optional: true,
			},
			"download_concurrency": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					suppressDownloadConcurrencyDiff{},
				},
			},
			"mirror_on_sync": schema.BoolAttribute{
				Optional: true,
			},
			"mirroring_policy": schema.StringAttribute{
				Optional: true,
			},
			"http_proxy_policy": schema.StringAttribute{
				Optional: true,
			},
			"http_proxy_id": schema.Int64Attribute{
				Optional: true,
			},
			"ignore_global_proxy": schema.BoolAttribute{
				Optional: true,
			},
			"ignorable_content": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of content units to ignore while syncing a yum repository. " +
					"Must be subset of rpm,drpm,srpm,distribution,erratum",
			},
			"verify_ssl_on_sync": schema.BoolAttribute{
				Optional: true,
			},
			"upstream_username": schema.StringAttribute{
				Optional: true,
			},
			"upstream_password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"deb_releases": schema.StringAttribute{
				Optional: true,
			},
			"deb_components": schema.StringAttribute{
				Optional: true,
			},
			"deb_architectures": schema.StringAttribute{
				Optional: true,
			},
			"docker_upstream_name": schema.StringAttribute{
				Optional: true,
			},
			"docker_tags_whitelist": schema.StringAttribute{
				Optional: true,
			},
			"ansible_collection_requirements": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *katelloRepositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*goforeman.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", "Expected *goforeman.Client")
		return
	}
	r.client = client
}

func (r *katelloRepositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan katelloRepositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &goforeman.KatelloRepositoryRequest{
		Name:                          plan.Name.ValueString(),
		Description:                   plan.Description.ValueString(),
		Label:                         plan.Label.ValueString(),
		ProductID:                     int(plan.ProductID.ValueInt64()),
		ContentType:                   plan.ContentType.ValueString(),
		URL:                           plan.URL.ValueString(),
		GpgKeyID:                      int(plan.GpgKeyID.ValueInt64()),
		Unprotected:                   boolPointerOrNil(plan.Unprotected),
		ChecksumType:                  plan.ChecksumType.ValueString(),
		DownloadPolicy:                plan.DownloadPolicy.ValueString(),
		DownloadConcurrency:           int(plan.DownloadConcurrency.ValueInt64()),
		MirrorOnSync:                  boolPointerOrNil(plan.MirrorOnSync),
		MirroringPolicy:               plan.MirroringPolicy.ValueString(),
		HttpProxyPolicy:               plan.HttpProxyPolicy.ValueString(),
		HttpProxyID:                   int(plan.HttpProxyID.ValueInt64()),
		IgnoreGlobalProxy:             boolPointerOrNil(plan.IgnoreGlobalProxy),
		IgnorableContent:              stringListToSlice(plan.IgnorableContent),
		VerifySslOnSync:               boolPointerOrNil(plan.VerifySslOnSync),
		UpstreamUsername:              plan.UpstreamUsername.ValueString(),
		UpstreamPassword:              plan.UpstreamPassword.ValueString(),
		DebReleases:                   plan.DebReleases.ValueString(),
		DebComponents:                 plan.DebComponents.ValueString(),
		DebArchitectures:              plan.DebArchitectures.ValueString(),
		DockerUpstreamName:            plan.DockerUpstreamName.ValueString(),
		DockerTagsWhitelist:           plan.DockerTagsWhitelist.ValueString(),
		AnsibleCollectionRequirements: plan.AnsibleCollectionRequirements.ValueString(),
	}

	result, err := r.client.CreateKatelloRepository(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create katello repository, got error: %s", err))
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(int(result.ID)))
	plan.Name = types.StringValue(result.Name)
	plan.Description = types.StringValue(result.Description)
	plan.Label = types.StringValue(result.Label)
	plan.ProductID = types.Int64Value(int64(result.ProductID))
	plan.ContentType = types.StringValue(result.ContentType)
	plan.URL = types.StringValue(result.URL)
	plan.GpgKeyID = types.Int64Value(int64(result.GpgKeyID))
	plan.Unprotected = types.BoolValue(result.Unprotected)
	plan.ChecksumType = types.StringValue(result.ChecksumType)
	plan.DownloadPolicy = types.StringValue(result.DownloadPolicy)
	plan.DownloadConcurrency = types.Int64Value(int64(result.DownloadConcurrency))
	plan.MirrorOnSync = types.BoolValue(result.MirrorOnSync)
	plan.MirroringPolicy = types.StringValue(result.MirroringPolicy)
	plan.HttpProxyPolicy = types.StringValue(result.HttpProxyPolicy)
	plan.HttpProxyID = types.Int64Value(int64(result.HttpProxyID))
	plan.IgnoreGlobalProxy = types.BoolValue(result.IgnoreGlobalProxy)
	plan.IgnorableContent = stringSliceToList(result.IgnorableContent, &resp.Diagnostics)
	plan.VerifySslOnSync = types.BoolValue(result.VerifySslOnSync)
	plan.UpstreamUsername = types.StringValue(result.UpstreamUsername)
	plan.UpstreamPassword = types.StringValue(result.UpstreamPassword)
	plan.DebReleases = types.StringValue(result.DebReleases)
	plan.DebComponents = types.StringValue(result.DebComponents)
	plan.DebArchitectures = types.StringValue(result.DebArchitectures)
	plan.DockerUpstreamName = types.StringValue(result.DockerUpstreamName)
	plan.DockerTagsWhitelist = types.StringValue(result.DockerTagsWhitelist)
	plan.AnsibleCollectionRequirements = types.StringValue(result.AnsibleCollectionRequirements)

	tflog.Trace(ctx, "created katello repository", map[string]interface{}{"id": plan.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *katelloRepositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state katelloRepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	result, err := r.client.ReadKatelloRepository(ctx, id)
	if err != nil {
		if goforeman.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read katello repository, got error: %s", err))
		return
	}
	state.Name = types.StringValue(result.Name)
	state.Description = types.StringValue(result.Description)
	state.Label = types.StringValue(result.Label)
	state.ProductID = types.Int64Value(int64(result.ProductID))
	state.ContentType = types.StringValue(result.ContentType)
	state.URL = types.StringValue(result.URL)
	state.GpgKeyID = types.Int64Value(int64(result.GpgKeyID))
	state.Unprotected = types.BoolValue(result.Unprotected)
	state.ChecksumType = types.StringValue(result.ChecksumType)
	state.DownloadPolicy = types.StringValue(result.DownloadPolicy)
	state.DownloadConcurrency = types.Int64Value(int64(result.DownloadConcurrency))
	state.MirrorOnSync = types.BoolValue(result.MirrorOnSync)
	state.MirroringPolicy = types.StringValue(result.MirroringPolicy)
	state.HttpProxyPolicy = types.StringValue(result.HttpProxyPolicy)
	state.HttpProxyID = types.Int64Value(int64(result.HttpProxyID))
	state.IgnoreGlobalProxy = types.BoolValue(result.IgnoreGlobalProxy)
	state.IgnorableContent = stringSliceToList(result.IgnorableContent, &resp.Diagnostics)
	state.VerifySslOnSync = types.BoolValue(result.VerifySslOnSync)
	state.UpstreamUsername = types.StringValue(result.UpstreamUsername)
	state.UpstreamPassword = types.StringValue(result.UpstreamPassword)
	state.DebReleases = types.StringValue(result.DebReleases)
	state.DebComponents = types.StringValue(result.DebComponents)
	state.DebArchitectures = types.StringValue(result.DebArchitectures)
	state.DockerUpstreamName = types.StringValue(result.DockerUpstreamName)
	state.DockerTagsWhitelist = types.StringValue(result.DockerTagsWhitelist)
	state.AnsibleCollectionRequirements = types.StringValue(result.AnsibleCollectionRequirements)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *katelloRepositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan katelloRepositoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	body := &goforeman.KatelloRepositoryRequest{
		Name:                          plan.Name.ValueString(),
		Description:                   plan.Description.ValueString(),
		Label:                         plan.Label.ValueString(),
		ProductID:                     int(plan.ProductID.ValueInt64()),
		ContentType:                   plan.ContentType.ValueString(),
		URL:                           plan.URL.ValueString(),
		GpgKeyID:                      int(plan.GpgKeyID.ValueInt64()),
		Unprotected:                   boolPointerOrNil(plan.Unprotected),
		ChecksumType:                  plan.ChecksumType.ValueString(),
		DownloadPolicy:                plan.DownloadPolicy.ValueString(),
		DownloadConcurrency:           int(plan.DownloadConcurrency.ValueInt64()),
		MirrorOnSync:                  boolPointerOrNil(plan.MirrorOnSync),
		MirroringPolicy:               plan.MirroringPolicy.ValueString(),
		HttpProxyPolicy:               plan.HttpProxyPolicy.ValueString(),
		HttpProxyID:                   int(plan.HttpProxyID.ValueInt64()),
		IgnoreGlobalProxy:             boolPointerOrNil(plan.IgnoreGlobalProxy),
		IgnorableContent:              stringListToSlice(plan.IgnorableContent),
		VerifySslOnSync:               boolPointerOrNil(plan.VerifySslOnSync),
		UpstreamUsername:              plan.UpstreamUsername.ValueString(),
		UpstreamPassword:              plan.UpstreamPassword.ValueString(),
		DebReleases:                   plan.DebReleases.ValueString(),
		DebComponents:                 plan.DebComponents.ValueString(),
		DebArchitectures:              plan.DebArchitectures.ValueString(),
		DockerUpstreamName:            plan.DockerUpstreamName.ValueString(),
		DockerTagsWhitelist:           plan.DockerTagsWhitelist.ValueString(),
		AnsibleCollectionRequirements: plan.AnsibleCollectionRequirements.ValueString(),
	}

	result, err := r.client.UpdateKatelloRepository(ctx, id, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update katello repository, got error: %s", err))
		return
	}
	plan.Name = types.StringValue(result.Name)
	plan.Description = types.StringValue(result.Description)
	plan.Label = types.StringValue(result.Label)
	plan.ProductID = types.Int64Value(int64(result.ProductID))
	plan.ContentType = types.StringValue(result.ContentType)
	plan.URL = types.StringValue(result.URL)
	plan.GpgKeyID = types.Int64Value(int64(result.GpgKeyID))
	plan.Unprotected = types.BoolValue(result.Unprotected)
	plan.ChecksumType = types.StringValue(result.ChecksumType)
	plan.DownloadPolicy = types.StringValue(result.DownloadPolicy)
	plan.DownloadConcurrency = types.Int64Value(int64(result.DownloadConcurrency))
	plan.MirrorOnSync = types.BoolValue(result.MirrorOnSync)
	plan.MirroringPolicy = types.StringValue(result.MirroringPolicy)
	plan.HttpProxyPolicy = types.StringValue(result.HttpProxyPolicy)
	plan.HttpProxyID = types.Int64Value(int64(result.HttpProxyID))
	plan.IgnoreGlobalProxy = types.BoolValue(result.IgnoreGlobalProxy)
	plan.IgnorableContent = stringSliceToList(result.IgnorableContent, &resp.Diagnostics)
	plan.VerifySslOnSync = types.BoolValue(result.VerifySslOnSync)
	plan.UpstreamUsername = types.StringValue(result.UpstreamUsername)
	plan.UpstreamPassword = types.StringValue(result.UpstreamPassword)
	plan.DebReleases = types.StringValue(result.DebReleases)
	plan.DebComponents = types.StringValue(result.DebComponents)
	plan.DebArchitectures = types.StringValue(result.DebArchitectures)
	plan.DockerUpstreamName = types.StringValue(result.DockerUpstreamName)
	plan.DockerTagsWhitelist = types.StringValue(result.DockerTagsWhitelist)
	plan.AnsibleCollectionRequirements = types.StringValue(result.AnsibleCollectionRequirements)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *katelloRepositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state katelloRepositoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("Unable to parse ID: %s", err))
		return
	}

	err = r.client.DeleteKatelloRepository(ctx, id)
	if err != nil && !goforeman.IsNotFoundError(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete katello repository, got error: %s", err))
		return
	}
}

func (r *katelloRepositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// stringListToSlice converts a types.List of strings to a []string for the API
// request body. See issue #164: ignorable_content must be a list of content
// unit types (e.g. rpm, drpm, srpm), not a single string.
func stringListToSlice(l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	elements := l.Elements()
	out := make([]string, 0, len(elements))
	for _, e := range elements {
		if s, ok := e.(types.String); ok {
			out = append(out, s.ValueString())
		}
	}
	return out
}

// stringSliceToList converts a []string from the API response to a types.List
// of strings.
func stringSliceToList(ss []string, diags *diag.Diagnostics) types.List {
	if ss == nil {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, len(ss))
	for i, s := range ss {
		elems[i] = types.StringValue(s)
	}
	list, d := types.ListValue(types.StringType, elems)
	diags.Append(d...)
	return list
}

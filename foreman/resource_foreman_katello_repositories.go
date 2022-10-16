package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanKatelloRepository() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanKatelloRepositoryCreate,
		ReadContext:   resourceForemanKatelloRepositoryRead,
		UpdateContext: resourceForemanKatelloRepositoryUpdate,
		DeleteContext: resourceForemanKatelloRepositoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Repository", // todo
					autodoc.MetaSummary,
				),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Repository name."+
						"%s \"My Repository\"", //todo
					autodoc.MetaExample,
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Repository description."+
						"%s \"A repository description\"",
					autodoc.MetaExample,
				),
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"%s",
					autodoc.MetaExample,
				),
			},
			"product_id": {
				Type:     schema.TypeInt,
				Required: true,
				Description: fmt.Sprintf(
					"Product the repository belongs to."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"content_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"deb",
					"docker",
					"file",
					"puppet",
					"yum",
				}, false),
				Description: fmt.Sprintf(
					"Product the repository belongs to. Valid values include:"+
						"`\"deb\"`, \"docker\"`, \"file\"`, \"puppet\"`, \"yum\"`."+
						"%s \"yum\"",
					autodoc.MetaExample,
				),
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Repository source url."+
						"%s \"http://mirror.centos.org/centos/7/os/x86_64/\"",
					autodoc.MetaExample,
				),
			},
			"gpg_key_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Identifier of the GPG key."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"unprotected": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: fmt.Sprintf(
					"true if this repository can be published via HTTP."+
						"%s true",
					autodoc.MetaExample,
				),
			},
			"checksum_type": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Checksum of the repository, currently 'sha1' & 'sha256' are supported"+
						"%s \"sha256\"",
					autodoc.MetaExample,
				),
			},
			"docker_upstream_name": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Name of the upstream docker repository"+
						"%s",
					autodoc.MetaExample,
				),
			},
			"docker_tags_whitelist": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Comma separated list of tags to sync for Container Image repository."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"download_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"immediate",
					"on_demand",
					"background",
				}, false),
				Description: fmt.Sprintf(
					"Product the repository belongs to. Valid values include:"+
						"`\"immediate\"`, \"on_demand\"`, \"background\"`."+
						"%s \"immediate\"",
					autodoc.MetaExample,
				),
			},
			"download_concurrency": {
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Used to determine download concurrency of the repository in pulp3. "+
						"Use value less than 20. Defaults to 10"+
						"%s",
					autodoc.MetaExample,
				),
			},
			"mirror_on_sync": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: fmt.Sprintf(
					"true if this repository when synced has to be mirrored from the source and stale rpms removed."+
						"%s true",
					autodoc.MetaExample,
				),
			},
			"verify_ssl_on_sync": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: fmt.Sprintf(
					"If true, Katello will verify the upstream url's SSL certifcates are signed by a trusted CA."+
						"%s true",
					autodoc.MetaExample,
				),
			},
			"upstream_username": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Username of the upstream repository user used for authentication."+
						"%s \"admin\"",
					autodoc.MetaExample,
				),
			},
			"upstream_password": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Password of the upstream repository user used for authentication."+
						"%s \"S3cr3t123!\"",
					autodoc.MetaExample,
				),
			},
			"deb_releases": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Comma separated list of releases to be synched from deb-archive."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"deb_components": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Comma separated list of repo components to be synched from deb-archive."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"deb_architectures": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Comma separated list of architectures to be synched from deb-archive."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"ignore_global_proxy": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: fmt.Sprintf(
					"If true, will ignore the globally configured proxy when syncing."+
						"%s true",
					autodoc.MetaExample,
				),
			},
			"ignorable_content": { //array
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"List of content units to ignore while syncing a yum repository. "+
						"Must be subset of rpm,drpm,srpm,distribution,erratum"+
						"%s",
					autodoc.MetaExample,
				),
			},
			"ansible_collection_requirements": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Contents of requirement yaml file to sync from URL."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"http_proxy_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"global_default_http_proxy",
					"none",
					"use_selected_http_proxy",
				}, false),
				Description: fmt.Sprintf(
					"Policies for HTTP proxy for content sync. Valid values include:"+
						"`\"global_default_http_proxy\"`, \"none\"`, \"use_selected_http_proxy\"`."+
						"%s \"global_default_http_proxy\"",
					autodoc.MetaExample,
				),
			},
			"http_proxy_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"ID of a HTTP Proxy."+
						"%s",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanKatelloRepository constructs a ForemanKatelloRepository struct from a resource
// data reference. The struct's members are populated from the data populated
// in the resource data. Missing members will be left to the zero value for
// that member's type.
func buildForemanKatelloRepository(d *schema.ResourceData) *api.ForemanKatelloRepository {
	log.Tracef("resource_foreman_katello_repository.go#buildForemanKatelloRepository")

	Repository := api.ForemanKatelloRepository{}

	obj := buildForemanObject(d)
	Repository.ForemanObject = *obj

	Repository.Description = d.Get("description").(string)
	Repository.Label = d.Get("label").(string)
	Repository.ProductId = d.Get("product_id").(int)
	Repository.ContentType = d.Get("content_type").(string)
	Repository.Url = d.Get("url").(string)
	Repository.GpgKeyId = d.Get("gpg_key_id").(int)
	Repository.Unprotected = d.Get("unprotected").(bool)
	Repository.ChecksumType = d.Get("checksum_type").(string)
	Repository.DockerUpstreamName = d.Get("docker_upstream_name").(string)
	Repository.DockerTagsWhitelist = d.Get("docker_tags_whitelist").(string)
	Repository.DownloadPolicy = d.Get("download_policy").(string)
	Repository.DownloadConcurrency = d.Get("download_concurrency").(int)
	Repository.MirrorOnSync = d.Get("mirror_on_sync").(bool)
	Repository.VerifySslOnSync = d.Get("verify_ssl_on_sync").(bool)
	Repository.UpstreamUsername = d.Get("upstream_username").(string)
	Repository.UpstreamPassword = d.Get("upstream_password").(string)
	Repository.DebReleases = d.Get("deb_releases").(string)
	Repository.DebComponents = d.Get("deb_components").(string)
	Repository.DebArchitectures = d.Get("deb_architectures").(string)
	Repository.IgnoreGlobalProxy = d.Get("ignore_global_proxy").(bool)
	Repository.IgnorableContent = d.Get("ignorable_content").(string)
	Repository.AnsibleCollectionRequirements = d.Get("ansible_collection_requirements").(string)
	Repository.HttpProxyPolicy = d.Get("http_proxy_policy").(string)
	Repository.HttpProxyId = d.Get("http_proxy_id").(int)

	return &Repository
}

// setResourceDataFromForemanKatelloRepository sets a ResourceData's attributes from
// the attributes of the supplied ForemanKatelloRepository struct
func setResourceDataFromForemanKatelloRepository(d *schema.ResourceData, Repository *api.ForemanKatelloRepository) {
	log.Tracef("resource_foreman_katello_repository.go#setResourceDataFromForemanKatelloRepository")

	d.SetId(strconv.Itoa(Repository.Id))
	d.Set("name", Repository.Name)
	d.Set("description", Repository.Description)
	d.Set("label", Repository.Label)
	d.Set("product_id", Repository.ProductId)
	d.Set("content_type", Repository.ContentType)
	d.Set("url", Repository.Url)
	d.Set("gpg_key_id", Repository.GpgKeyId)
	d.Set("unprotected", Repository.Unprotected)
	d.Set("checksum_type", Repository.ChecksumType)
	d.Set("docker_upstream_name", Repository.DockerUpstreamName)
	d.Set("docker_tags_whitelist", Repository.DockerTagsWhitelist)
	d.Set("download_policy", Repository.DownloadPolicy)
	d.Set("download_concurrency", Repository.DownloadConcurrency)
	d.Set("mirror_on_sync", Repository.MirrorOnSync)
	d.Set("verify_ssl_on_sync", Repository.VerifySslOnSync)
	d.Set("upstream_username", Repository.UpstreamUsername)
	d.Set("upstream_password", Repository.UpstreamPassword)
	d.Set("deb_releases", Repository.DebReleases)
	d.Set("deb_components", Repository.DebComponents)
	d.Set("deb_architectures", Repository.DebArchitectures)
	d.Set("ignore_global_proxy", Repository.IgnoreGlobalProxy)
	d.Set("ignorable_content", Repository.IgnorableContent)
	d.Set("ansible_collection_requirements", Repository.AnsibleCollectionRequirements)
	d.Set("http_proxy_policy", Repository.HttpProxyPolicy)
	d.Set("http_proxy_id", Repository.HttpProxyId)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanKatelloRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_katello_repository.go#Create")

	client := meta.(*api.Client)
	repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", repository)

	createdKatelloRepository, createErr := client.CreateKatelloRepository(ctx, repository)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanKatelloRepository: [%+v]", createdKatelloRepository)

	setResourceDataFromForemanKatelloRepository(d, createdKatelloRepository)

	return nil
}

func resourceForemanKatelloRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_katello_repository.go#Read")

	client := meta.(*api.Client)
	repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", repository)

	readKatelloRepository, readErr := client.ReadKatelloRepository(ctx, repository.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanKatelloRepository: [%+v]", readKatelloRepository)

	setResourceDataFromForemanKatelloRepository(d, readKatelloRepository)

	return nil
}

func resourceForemanKatelloRepositoryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_katello_repository.go#Update")

	client := meta.(*api.Client)
	repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", repository)

	updatedKatelloRepository, updateErr := client.UpdateKatelloRepository(ctx, repository)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("ForemanKatelloRepository: [%+v]", updatedKatelloRepository)

	setResourceDataFromForemanKatelloRepository(d, updatedKatelloRepository)

	return nil
}

func resourceForemanKatelloRepositoryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_katello_repository.go#Delete")

	client := meta.(*api.Client)
	repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", repository)

	return diag.FromErr(api.CheckDeleted(d, client.DeleteKatelloRepository(ctx, repository.Id)))
}

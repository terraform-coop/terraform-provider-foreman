package foreman

import (
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceForemanKatelloRepository() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanKatelloRepositoryCreate,
		Read:   resourceForemanKatelloRepositoryRead,
		Update: resourceForemanKatelloRepositoryUpdate,
		Delete: resourceForemanKatelloRepositoryDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Repository", // todo
					autodoc.MetaSummary,
				),
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Repository name."+
						"%s \"My Repository\"", //todo
					autodoc.MetaExample,
				),
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Repository description."+
						"%s \"A repository description\"",
					autodoc.MetaExample,
				),
			},
			"label": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"%s",
					autodoc.MetaExample,
				),
			},
			"product_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				Description: fmt.Sprintf(
					"Product the repository belongs to."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"content_type": &schema.Schema{
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
						"%s \"deb\"",
					autodoc.MetaExample,
				),
			},
			"url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Repository source url."+
						"%s \"http://mirror.centos.org/centos/7/os/x86_64/\"",
					autodoc.MetaExample,
				),
			},
			"gpg_key_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Identifier of the GPG key."+
						"%s",
					autodoc.MetaExample,
				),
			},
			/* "ssl_ca_cert_id": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Description: fmt.Sprintf(
								"Idenifier of the SSL CA Cert."+
									"%s",
								autodoc.MetaExample,
							),
						},
			            "ssl_client_cert_id": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Description: fmt.Sprintf(
								"Identifier of the SSL Client Cert."+
									"%s",
								autodoc.MetaExample,
							),
						},
			            "ssl_client_key_id": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Description: fmt.Sprintf(
								"Identifier of the SSL Client Key."+
									"%s",
								autodoc.MetaExample,
							),
						}, */
			"unprotected": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Description: fmt.Sprintf(
					"true if this repository can be published via HTTP."+
						"%s true",
					autodoc.MetaExample,
				),
			},
			"checksum_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Checksum of the repository, currently 'sha1' & 'sha256' are supported"+
						"%s \"sha256\"",
					autodoc.MetaExample,
				),
			},
			"docker_upstream_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Name of the upstream docker repository"+
						"%s",
					autodoc.MetaExample,
				),
			},
			"docker_tags_whitelist": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Comma separated list of tags to sync for Container Image repository."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"download_policy": &schema.Schema{
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
			"download_concurrency": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Used to determine download concurrency of the repository in pulp3. "+
						"Use value less than 20. Defaults to 10"+
						"%s",
					autodoc.MetaExample,
				),
			},
			"mirror_on_sync": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Description: fmt.Sprintf(
					"true if this repository when synced has to be mirrored from the source and stale rpms removed."+
						"%s true",
					autodoc.MetaExample,
				),
			},
			"verify_ssl_on_sync": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Description: fmt.Sprintf(
					"If true, Katello will verify the upstream url's SSL certifcates are signed by a trusted CA."+
						"%s true",
					autodoc.MetaExample,
				),
			},
			"upstream_username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Username of the upstream repository user used for authentication."+
						"%s \"admin\"",
					autodoc.MetaExample,
				),
			},
			"upstream_password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Password of the upstream repository user used for authentication."+
						"%s \"S3cr3t123!\"",
					autodoc.MetaExample,
				),
			},
			"ostree_upstream_sync_policy": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"latest",
					"all",
					"custom",
				}, false),
				Description: fmt.Sprintf(
					"Policies for syncing upstream ostree repositories. Valid values include:"+
						"`\"latest\"`, \"all\"`, \"custom\"`."+
						"%s \"latest\"",
					autodoc.MetaExample,
				),
			},
			"ostree_upstream_sync_depth": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"If a custom sync policy is chosen for ostree repositories then a 'depth' value must be provided."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"deb_releases": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Comma separated list of releases to be synched from deb-archive."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"deb_components": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Comma separated list of repo components to be synched from deb-archive."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"deb_architectures": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Comma separated list of architectures to be synched from deb-archive."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"ignore_global_proxy": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Description: fmt.Sprintf(
					"If true, will ignore the globally configured proxy when syncing."+
						"%s true",
					autodoc.MetaExample,
				),
			},
			"ignorable_content": &schema.Schema{ //array
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"List of content units to ignore while syncing a yum repository. "+
						"Must be subset of rpm,drpm,srpm,distribution,erratum"+
						"%s",
					autodoc.MetaExample,
				),
			},
			"ansible_collection_requirements": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Contents of requirement yaml file to sync from URL."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"http_proxy_policy": &schema.Schema{
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
			"http_proxy_id": &schema.Schema{
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
	/* 	Repository.SslCaCertId = d.Get("ssl_ca_cert_id").(int)
	   	Repository.SslClientCertId = d.Get("ssl_client_cert_id").(int)
	   	Repository.SslClientKeyId = d.Get("ssl_client_key_id").(int) */
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
	Repository.OstreeUpstreamSyncPolicy = d.Get("ostree_upstream_sync_policy").(string)
	Repository.OstreeUpstreamSyncDepth = d.Get("ostree_upstream_sync_depth").(int)
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
	/* 	d.Set("ssl_ca_cert_id", Repository.SslCaCertId)
	   	d.Set("ssl_client_cert_id", Repository.SslClientCertId)
	   	d.Set("ssl_client_key_id", Repository.SslClientKeyId) */
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
	d.Set("ostree_upstream_sync_policy", Repository.OstreeUpstreamSyncPolicy)
	d.Set("ostree_upstream_sync_depth", Repository.OstreeUpstreamSyncDepth)
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

func resourceForemanKatelloRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_katello_repository.go#Create")

	client := meta.(*api.Client)
	Repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", Repository)

	createdKatelloRepository, createErr := client.CreateKatelloRepository(Repository)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanKatelloRepository: [%+v]", createdKatelloRepository)

	setResourceDataFromForemanKatelloRepository(d, createdKatelloRepository)

	return nil
}

func resourceForemanKatelloRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_katello_repository.go#Read")

	client := meta.(*api.Client)
	Repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", Repository)

	readKatelloRepository, readErr := client.ReadKatelloRepository(Repository.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanKatelloRepository: [%+v]", readKatelloRepository)

	setResourceDataFromForemanKatelloRepository(d, readKatelloRepository)

	return nil
}

func resourceForemanKatelloRepositoryUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_katello_repository.go#Update")

	client := meta.(*api.Client)
	Repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", Repository)

	updatedKatelloRepository, updateErr := client.UpdateKatelloRepository(Repository)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("ForemanKatelloRepository: [%+v]", updatedKatelloRepository)

	setResourceDataFromForemanKatelloRepository(d, updatedKatelloRepository)

	return nil
}

func resourceForemanKatelloRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_katello_repository.go#Delete")

	client := meta.(*api.Client)
	Repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", Repository)

	return client.DeleteKatelloRepository(Repository.Id)
}

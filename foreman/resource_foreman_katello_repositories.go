package foreman

import (
	"context"
	"errors"
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
				Computed: true,
				Description: "Label of the repository. Cannot be changed after creation. " +
					"Is auto generated from name if not specified.",
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
					"ansible_collection",
				}, false),
				Description: fmt.Sprintf(
					"Content type of the repository. Valid values include:"+
						"`\"deb\"`, \"docker\"`, \"file\"`, \"puppet\"`, \"yum\"`, `\"ansible_collection\"`."+
						"%s \"yum\"",
					autodoc.MetaExample,
				),
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Repository source URL or Docker registry URL"+
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
					"Used to determine download concurrency of the repository in pulp3. " +
						"Use value less than 20. Defaults to 10. Warning: the value is not returned from the API and " +
						"is therefore handled by a DiffSuppressFunc.",
				),
				ValidateFunc: validation.IntBetween(1, 20),
				DiffSuppressFunc: func(key, oldValue, newValue string, d *schema.ResourceData) bool {
					// "download_concurrency" is not returned from the Katello API, but still exists in the
					// source code at https://github.com/Katello/katello/blob/6d8d3ca36e1469d1f7c2c8e180e42467176ac1a4/app/controllers/katello/api/v2/repositories_controller.rb#L56.
					// So we use a diffsuppression if the value is defined in the .tf manifest, but
					// would be reset to 0 every time an "apply" is executed.
					newAsInt, err := strconv.Atoi(newValue)
					if err != nil {
						log.Fatalf("download_concurrency value was not an int!")
					}

					if oldValue == "0" && newAsInt > 0 {
						return true
					}
					return false
				},
			},
			"mirror_on_sync": {
				Type:       schema.TypeBool,
				Optional:   true,
				Deprecated: "Deprecated and removed in Katello 4.9 in favor of mirroring_policy",
				Description: fmt.Sprintf(
					"'True' if this repository when synced has to be mirrored from the source and stale rpms removed.",
				),
			},
			"mirroring_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf("Mirroring policy for this repo. Values: \"mirror_content_only\" "+
					"or \"additive\". %s \"mirror_content_only\"", autodoc.MetaExample),
				ValidateFunc: validation.StringInSlice([]string{
					"additive",
					"mirror_content_only",
				}, false),
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
				Default:  "global_default_http_proxy",
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

	repo := api.ForemanKatelloRepository{}

	obj := buildForemanObject(d)
	repo.ForemanObject = *obj

	repo.Description = d.Get("description").(string)
	repo.Label = d.Get("label").(string)
	repo.ProductId = d.Get("product_id").(int)
	repo.ContentType = d.Get("content_type").(string)
	repo.Url = d.Get("url").(string)
	repo.GpgKeyId = d.Get("gpg_key_id").(int)
	repo.Unprotected = d.Get("unprotected").(bool)
	repo.ChecksumType = d.Get("checksum_type").(string)
	repo.DockerUpstreamName = d.Get("docker_upstream_name").(string)
	repo.DockerTagsWhitelist = d.Get("docker_tags_whitelist").(string)
	repo.DownloadPolicy = d.Get("download_policy").(string)
	repo.DownloadConcurrency = d.Get("download_concurrency").(int)
	repo.MirrorOnSync = d.Get("mirror_on_sync").(bool)
	repo.MirroringPolicy = d.Get("mirroring_policy").(string)
	repo.VerifySslOnSync = d.Get("verify_ssl_on_sync").(bool)
	repo.UpstreamUsername = d.Get("upstream_username").(string)
	repo.UpstreamPassword = d.Get("upstream_password").(string)
	repo.DebReleases = d.Get("deb_releases").(string)
	repo.DebComponents = d.Get("deb_components").(string)
	repo.DebArchitectures = d.Get("deb_architectures").(string)
	repo.IgnoreGlobalProxy = d.Get("ignore_global_proxy").(bool)
	repo.IgnorableContent = d.Get("ignorable_content").(string)
	repo.AnsibleCollectionRequirements = d.Get("ansible_collection_requirements").(string)
	repo.HttpProxyPolicy = d.Get("http_proxy_policy").(string)
	repo.HttpProxyId = d.Get("http_proxy_id").(int)

	return &repo
}

// setResourceDataFromForemanKatelloRepository sets a ResourceData's attributes from
// the attributes of the supplied ForemanKatelloRepository struct
func setResourceDataFromForemanKatelloRepository(d *schema.ResourceData, repo *api.ForemanKatelloRepository) {
	log.Tracef("resource_foreman_katello_repository.go#setResourceDataFromForemanKatelloRepository")

	d.SetId(strconv.Itoa(repo.Id))
	d.Set("name", repo.Name)
	d.Set("description", repo.Description)
	d.Set("label", repo.Label)
	d.Set("product_id", repo.ProductId)
	d.Set("content_type", repo.ContentType)
	d.Set("url", repo.Url)
	d.Set("gpg_key_id", repo.GpgKeyId)
	d.Set("unprotected", repo.Unprotected)
	d.Set("checksum_type", repo.ChecksumType)
	d.Set("docker_upstream_name", repo.DockerUpstreamName)
	d.Set("docker_tags_whitelist", repo.DockerTagsWhitelist)
	d.Set("download_policy", repo.DownloadPolicy)

	if repo.DownloadConcurrency > 0 {
		// In case it is 0 and unset, the value will default
		d.Set("download_concurrency", repo.DownloadConcurrency)
	}

	d.Set("mirror_on_sync", repo.MirrorOnSync)
	d.Set("mirroring_policy", repo.MirroringPolicy)
	d.Set("verify_ssl_on_sync", repo.VerifySslOnSync)
	d.Set("upstream_username", repo.UpstreamUsername)
	d.Set("upstream_password", repo.UpstreamPassword)
	d.Set("deb_releases", repo.DebReleases)
	d.Set("deb_components", repo.DebComponents)
	d.Set("deb_architectures", repo.DebArchitectures)
	d.Set("ignore_global_proxy", repo.IgnoreGlobalProxy)
	d.Set("ignorable_content", repo.IgnorableContent)
	d.Set("ansible_collection_requirements", repo.AnsibleCollectionRequirements)
	d.Set("http_proxy_policy", repo.HttpProxyPolicy)
	d.Set("http_proxy_id", repo.HttpProxyId)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func handleDownloadConcurrencyBetweenTerraformAndKatello(resData *schema.ResourceData, repo *api.ForemanKatelloRepository) error {
	log.Tracef("handleDownloadConcurrencyBetweenTerraformAndKatello")

	// Handle missing download_concurrency attribute in API response
	originalDLC, ok := resData.GetOk("download_concurrency")
	if ok {
		originalDLCint, ok := originalDLC.(int)

		if !ok {
			return errors.New("unable to convert 'download_concurrency' state value to int")
		}

		if originalDLCint > 0 && repo.DownloadConcurrency == 0 {
			// We passed in a value > 0, but the Katello API did not return this value and therefore the default 0 was applied to the struct
			log.Debugf("State has download_concurrency of %d, but repo has 0. Setting repo object parameter to state", originalDLCint)
			repo.DownloadConcurrency = originalDLCint
		}
	}
	return nil
}

func resourceForemanKatelloRepositoryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_katello_repository.go#Create")

	client := meta.(*api.Client)
	repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", repository)

	createdKatelloRepository, createErr := client.CreateKatelloRepository(ctx, repository)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	err := handleDownloadConcurrencyBetweenTerraformAndKatello(d, createdKatelloRepository)
	if err != nil {
		return diag.FromErr(err)
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

	err := handleDownloadConcurrencyBetweenTerraformAndKatello(d, readKatelloRepository)
	if err != nil {
		return diag.FromErr(err)
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

	err := handleDownloadConcurrencyBetweenTerraformAndKatello(d, updatedKatelloRepository)
	if err != nil {
		return diag.FromErr(err)
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

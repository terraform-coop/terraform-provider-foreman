package foreman

import (
	"context"
	"fmt"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"strconv"
)

func resourceForemanKatelloLifecycleEnvironment() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanKatelloLifecycleEnvironmentCreate,
		ReadContext:   resourceForemanKatelloLifecycleEnvironmentRead,
		UpdateContext: resourceForemanKatelloLifecycleEnvironmentUpdate,
		DeleteContext: resourceForemanKatelloLifecycleEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		/*
			Left over, not implemented yet:

			RegistryNamePattern         string `json:"registry_name_pattern"`
			RegistryUnauthenticatedPull bool   `json:"registry_unauthenticated_pull"`

			Counts struct {
				ContentHosts int `json:"content_hosts"`
				ContentViews int `json:"content_views"`
			} `json:"counts"`

			ContentViews []ContentViews `json:"content_views"`
		*/

		Schema: map[string]*schema.Schema{
			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Lifecycle environments group hosts into logical stages, example dev/test/prod.",
					autodoc.MetaSummary,
				),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the lifecycle environment. %s \"My new env\"", autodoc.MetaExample),
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description for the lifecycle environment",
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Created from name if not passed in
				ForceNew: true,
				Description: fmt.Sprintf(
					"Label for the lifecycle environment. Cannot be changed after creation. "+
						"By default set to the name, with underscores as spaces replacement. %s",
					autodoc.MetaExample,
				),
			},
			"organization_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: fmt.Sprintf("%s 1", autodoc.MetaExample),
			},
			"library": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Specifies if this environment is the special 'Library' root environment.",
			},
			"prior_id": {
				Type:     schema.TypeInt,
				Required: true,
				Description: fmt.Sprintf("ID of the prior lifecycle environment. Use '1' to refer to "+
					"the built-in 'Library' root environment. "+
					"%s data.foreman_katello_lifecycle_environment.library.id", autodoc.MetaExample),
			},
			"successor_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func buildForemanKatelloLifecycleEnvironment(d *schema.ResourceData) *api.LifecycleEnvironment {
	utils.TraceFunctionCall()

	lce := api.LifecycleEnvironment{}
	lce.ForemanObject = *buildForemanObject(d)

	lce.Description = d.Get("description").(string)
	lce.Label = d.Get("label").(string)
	lce.OrganizationId = d.Get("organization_id").(int)
	lce.Library = d.Get("library").(bool)
	lce.Prior.Id = d.Get("prior_id").(int)
	lce.Successor.Id = d.Get("successor_id").(int)

	return &lce
}

func setResourceDataFromForemanKatelloLifecycleEnvironment(d *schema.ResourceData, lce *api.LifecycleEnvironment) {
	utils.TraceFunctionCall()

	d.SetId(strconv.Itoa(lce.Id))
	d.Set("name", lce.Name)
	d.Set("description", lce.Description)
	d.Set("label", lce.Label)
	d.Set("organization_id", lce.OrganizationId)
	d.Set("library", lce.Library)
	d.Set("prior_id", lce.Prior.Id)
	d.Set("successor_id", lce.Successor.Id)
}

func resourceForemanKatelloLifecycleEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	lce := buildForemanKatelloLifecycleEnvironment(d)
	utils.Debugf("lce: %+v", lce)

	createdLce, err := client.CreateKatelloLifecycleEnvironment(ctx, lce)
	if err != nil {
		return diag.FromErr(err)
	}
	utils.Debugf("Created lce: %+v", createdLce)

	setResourceDataFromForemanKatelloLifecycleEnvironment(d, createdLce)
	return nil
}

func resourceForemanKatelloLifecycleEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	lce := buildForemanKatelloLifecycleEnvironment(d)

	readLce, readErr := client.ReadKatelloLifecycleEnvironment(ctx, lce)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}
	utils.Debugf("Read lifecycle env: %+v", readLce)

	setResourceDataFromForemanKatelloLifecycleEnvironment(d, readLce)
	return nil
}

func resourceForemanKatelloLifecycleEnvironmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	lce := buildForemanKatelloLifecycleEnvironment(d)
	utils.Debugf("lce: [%+v]", lce)

	updatedLce, err := client.UpdateKatelloLifecycleEnvironment(ctx, lce)
	if err != nil {
		return diag.FromErr(err)
	}
	utils.Debugf("updatedLce: %+v", updatedLce)

	setResourceDataFromForemanKatelloLifecycleEnvironment(d, updatedLce)
	return nil
}

func resourceForemanKatelloLifecycleEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	lce := buildForemanKatelloLifecycleEnvironment(d)

	utils.Debugf("lce to be deleted: %+v", lce)

	return diag.FromErr(api.CheckDeleted(d, client.DeleteKatelloLifecycleEnvironment(ctx, lce.Id)))
}

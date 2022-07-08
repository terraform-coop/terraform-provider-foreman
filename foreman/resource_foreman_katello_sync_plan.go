package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanKatelloSyncPlan() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanKatelloSyncPlanCreate,
		ReadContext:   resourceForemanKatelloSyncPlanRead,
		UpdateContext: resourceForemanKatelloSyncPlanUpdate,
		DeleteContext: resourceForemanKatelloSyncPlanDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s A sync plan is used to schedule a synchronization of a product in katello",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Sync plan name."+
						"%s \"daily\"",
					autodoc.MetaExample,
				),
			},

			"interval": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"hourly",
					"daily",
					"weekly",
					"custom cron",
				}, false),
				Description: fmt.Sprintf(
					"How often synchronization should run. Valid "+
						"values include: `\"hourly\"`, `\"daily\"`, `\"weekly\"`,`\"custom cron\"`."+
						"%s \"daily\"",
					autodoc.MetaExample,
				),
			},
			"sync_date": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Start datetime of synchronization."+
						"%s \"1970-01-01 00:00:00 UTC\"",
					autodoc.MetaExample,
				),
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Sync plan description."+
						"%s \"A sync plan description\"",
					autodoc.MetaExample,
				),
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
				Description: fmt.Sprintf(
					"Enables or disables synchronization."+
						"%s true",
					autodoc.MetaExample,
				),
			},
			"cron_expression": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Custom cron logic for sync plan."+
						"%s \"*/5 * * * *\"",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanKatelloSyncPlan constructs a ForemanKatelloSyncPlan struct from a resource
// data reference.  The struct's members are populated from the data populated
// in the resource data.  Missing members will be left to the zero value for
// that member's type.
func buildForemanKatelloSyncPlan(d *schema.ResourceData) *api.ForemanKatelloSyncPlan {
	log.Tracef("resource_foreman_katello_sync_plan.go#buildForemanKatelloSyncPlan")

	syncPlan := api.ForemanKatelloSyncPlan{}

	obj := buildForemanObject(d)
	syncPlan.ForemanObject = *obj

	syncPlan.Interval = d.Get("interval").(string)
	syncPlan.SyncDate = d.Get("sync_date").(string)
	syncPlan.Description = d.Get("description").(string)
	syncPlan.Enabled = d.Get("enabled").(bool)
	syncPlan.CronExpression = d.Get("cron_expression").(string)

	return &syncPlan
}

// setResourceDataFromForemanKatelloSyncPlan sets a ResourceData's attributes from
// the attributes of the supplied ForemanKatelloSyncPlan struct
func setResourceDataFromForemanKatelloSyncPlan(d *schema.ResourceData, syncPlan *api.ForemanKatelloSyncPlan) {
	log.Tracef("resource_foreman_katello_sync_plan.go#setResourceDataFromForemanKatelloSyncPlan")

	d.SetId(strconv.Itoa(syncPlan.Id))
	d.Set("name", syncPlan.Name)
	d.Set("interval", syncPlan.Interval)
	d.Set("sync_date", syncPlan.SyncDate)
	d.Set("description", syncPlan.Description)
	d.Set("enabled", syncPlan.Enabled)
	d.Set("cron_expression", syncPlan.CronExpression)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanKatelloSyncPlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_katello_sync_plan.go#Create")

	client := meta.(*api.Client)
	syncPlan := buildForemanKatelloSyncPlan(d)

	log.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	createdKatelloSyncPlan, createErr := client.CreateKatelloSyncPlan(ctx, syncPlan)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanKatelloSyncPlan: [%+v]", createdKatelloSyncPlan)

	setResourceDataFromForemanKatelloSyncPlan(d, createdKatelloSyncPlan)

	return nil
}

func resourceForemanKatelloSyncPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_katello_sync_plan.go#Read")

	client := meta.(*api.Client)
	syncPlan := buildForemanKatelloSyncPlan(d)

	log.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	readKatelloSyncPlan, readErr := client.ReadKatelloSyncPlan(ctx, syncPlan.Id)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	log.Debugf("Read ForemanKatelloSyncPlan: [%+v]", readKatelloSyncPlan)

	setResourceDataFromForemanKatelloSyncPlan(d, readKatelloSyncPlan)

	return nil
}

func resourceForemanKatelloSyncPlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_katello_sync_plan.go#Update")

	client := meta.(*api.Client)
	syncPlan := buildForemanKatelloSyncPlan(d)

	log.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	updatedKatelloSyncPlan, updateErr := client.UpdateKatelloSyncPlan(ctx, syncPlan)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("ForemanKatelloSyncPlan: [%+v]", updatedKatelloSyncPlan)

	setResourceDataFromForemanKatelloSyncPlan(d, updatedKatelloSyncPlan)

	return nil
}

func resourceForemanKatelloSyncPlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_katello_sync_plan.go#Delete")

	client := meta.(*api.Client)
	syncPlan := buildForemanKatelloSyncPlan(d)

	log.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	return diag.FromErr(client.DeleteKatelloSyncPlan(ctx, syncPlan.Id))
}

package foreman

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"strconv"
	"strings"
	"time"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// TIMELAYOUT specifies the format of the datetime string used in sync_date
const TIMELAYOUT = "2006-01-02 15:04:05 -0700" // TZ as +-0000

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

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s A sync plan is used to schedule a synchronization of a product in katello",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Sync plan name."+
						"%s \"daily\"",
					autodoc.MetaExample,
				),
			},

			"interval": {
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

			"sync_date": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Start datetime of synchronization. Use the specified format: YYYY-MM-DD HH:MM:SS +0000, "+
						"where '+0000' is the timezone difference. A value of '+0000' means UTC. "+
						"%s \"1970-01-01 00:00:00 +0000\"",
					autodoc.MetaExample,
				),
				ValidateDiagFunc: func(obj interface{}, path cty.Path) diag.Diagnostics {
					datetimeString := obj.(string)

					if strings.Contains(datetimeString, "UTC") {
						utils.Warningf("sync_date used 'UTC' instead of '+0000'. This is internally corrected" +
							"because of historic documentation but might be changed in the future.")
						datetimeString = strings.Replace(datetimeString, "UTC", "+0000", 1)
					}

					_, err := time.Parse(TIMELAYOUT, datetimeString)
					if err != nil {
						e := fmt.Sprintf("Your 'sync_date' value is incorrectly formatted. Use the "+
							"format 'YYYY-MM-DD HH:MM:SS +0000' as documented. (Error: %s)", err)
						return diag.FromErr(errors.New(e))
					}
					return nil
				},
				DiffSuppressFunc: func(key, oldValue, newValue string, d *schema.ResourceData) bool {
					if oldValue == "" || newValue == "" {
						return false
					}

					// If someone uses "UTC" instead of +0000, replace the string first
					if strings.Contains(newValue, "UTC") {
						newValue = strings.Replace(newValue, "UTC", "+0000", 1)
					}

					// Then parse the old value
					tOld, err := time.Parse(TIMELAYOUT, oldValue)
					if err != nil {
						utils.Warningf("Error in time.Parse: %v", err)
						return false
					}

					// And the new value
					tNew, err := time.Parse(TIMELAYOUT, newValue)
					if err != nil {
						utils.Warningf("Error in time.Parse: %v", err)
						return false
					}

					// And compare the two time.Time objects
					if tOld == tNew {
						return true
					}
					return false
				},
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Sync plan description."+
						"%s \"A sync plan description\"",
					autodoc.MetaExample,
				),
			},

			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
				Description: fmt.Sprintf(
					"Enables or disables synchronization."+
						"%s true",
					autodoc.MetaExample,
				),
			},

			"cron_expression": {
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
	utils.TraceFunctionCall()

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
	utils.TraceFunctionCall()

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
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	syncPlan := buildForemanKatelloSyncPlan(d)

	utils.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	createdKatelloSyncPlan, createErr := client.CreateKatelloSyncPlan(ctx, syncPlan)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	utils.Debugf("Created ForemanKatelloSyncPlan: [%+v]", createdKatelloSyncPlan)

	setResourceDataFromForemanKatelloSyncPlan(d, createdKatelloSyncPlan)

	return nil
}

func resourceForemanKatelloSyncPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	syncPlan := buildForemanKatelloSyncPlan(d)

	utils.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	readKatelloSyncPlan, readErr := client.ReadKatelloSyncPlan(ctx, syncPlan.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	utils.Debugf("Read ForemanKatelloSyncPlan: [%+v]", readKatelloSyncPlan)

	setResourceDataFromForemanKatelloSyncPlan(d, readKatelloSyncPlan)

	return nil
}

func resourceForemanKatelloSyncPlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	syncPlan := buildForemanKatelloSyncPlan(d)

	utils.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	updatedKatelloSyncPlan, updateErr := client.UpdateKatelloSyncPlan(ctx, syncPlan)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	utils.Debugf("ForemanKatelloSyncPlan: [%+v]", updatedKatelloSyncPlan)

	setResourceDataFromForemanKatelloSyncPlan(d, updatedKatelloSyncPlan)

	return nil
}

func resourceForemanKatelloSyncPlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	syncPlan := buildForemanKatelloSyncPlan(d)

	utils.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	return diag.FromErr(api.CheckDeleted(d, client.DeleteKatelloSyncPlan(ctx, syncPlan.Id)))
}

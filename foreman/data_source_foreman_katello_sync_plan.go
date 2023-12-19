package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
)

func dataSourceForemanKatelloSyncPlan() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanKatelloSyncPlan()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"sync plan name."+
				"%s \"daily\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanKatelloSyncPlanRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanKatelloSyncPlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	syncPlan := buildForemanKatelloSyncPlan(d)

	log.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	queryResponse, queryErr := client.QueryKatelloSyncPlan(ctx, syncPlan)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("data source sync plan returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("data source sync plan returned more than 1 result")
	}

	var queryKatelloSyncPlan api.ForemanKatelloSyncPlan
	var ok bool
	if queryKatelloSyncPlan, ok = queryResponse.Results[0].(api.ForemanKatelloSyncPlan); !ok {
		return diag.Errorf(
			"data source results contain unexpected type. Expected "+
				"[api.ForemanKatelloSyncPlan], got [%T]",
			queryResponse.Results[0],
		)
	}
	syncPlan = &queryKatelloSyncPlan

	log.Debugf("ForemanKatelloSyncPlan: [%+v]", syncPlan)

	setResourceDataFromForemanKatelloSyncPlan(d, syncPlan)

	return nil
}

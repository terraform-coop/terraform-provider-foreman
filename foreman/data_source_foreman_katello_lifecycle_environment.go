package foreman

import (
	"context"
	"fmt"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
)

func dataSourceForemanKatelloLifecycleEnvironment() *schema.Resource {
	r := resourceForemanKatelloLifecycleEnvironment()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: fmt.Sprintf("Name of the lifecycle environment. %s \"Library\"", autodoc.MetaExample),
	}

	return &schema.Resource{
		ReadContext: dataSourceForemanKatelloLifecycleRead,
		Schema:      ds,
	}
}

func dataSourceForemanKatelloLifecycleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	lce := buildForemanKatelloLifecycleEnvironment(d)

	utils.Debugf("lifecycle env: %+v", lce)

	queryResponse, err := client.QueryLifecycleEnvironment(ctx, lce)
	if err != nil {
		return diag.FromErr(err)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("data source lifecycle_environment returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("data source lifecycle_environment returned more than 1 result")
	}

	if queryLce, ok := queryResponse.Results[0].(api.LifecycleEnvironment); !ok {
		return diag.Errorf(
			"data source results contain unexpected type. Expected "+
				"[api.LifecycleEnvironment], got [%T]",
			queryResponse.Results[0],
		)
	} else {
		lce = &queryLce
	}

	utils.Debugf("lifecycle env: %+v", lce)

	setResourceDataFromForemanKatelloLifecycleEnvironment(d, lce)

	return nil
}

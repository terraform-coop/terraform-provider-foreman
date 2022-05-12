package foreman

import (
	"context"
	"fmt"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanEnvironment() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanEnvironment()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the puppet branch, environment. "+
				"%s \"production\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanEnvironmentRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_environment.go#Read")

	client := meta.(*api.Client)
	e := buildForemanEnvironment(d)

	log.Debugf("ForemanEnvironment: [%+v]", e)

	queryResponse, queryErr := client.QueryEnvironment(ctx, e)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source environment returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source environment returned more than 1 result")
	}

	var queryEnvironment api.ForemanEnvironment
	var ok bool
	if queryEnvironment, ok = queryResponse.Results[0].(api.ForemanEnvironment); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanEnvironment], got [%T]",
			queryResponse.Results[0],
		)
	}
	e = &queryEnvironment

	log.Debugf("ForemanEnvironment: [%+v]", e)

	setResourceDataFromForemanEnvironment(d, e)

	return nil
}

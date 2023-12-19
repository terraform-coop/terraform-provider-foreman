package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanParameter() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanParameter()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the parameter - the full DNS parameter name. "+
				"%s \"dev.dc1.company.com\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanParameterRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	parameter := buildForemanParameter(d)

	log.Debugf("ForemanParameter: [%+v]", parameter)

	queryResponse, queryErr := client.QueryParameter(ctx, parameter)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source parameter returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source parameter returned more than 1 result")
	}

	var queryParameter api.ForemanParameter
	var ok bool
	if queryParameter, ok = queryResponse.Results[0].(api.ForemanParameter); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanParameter], got [%T]",
			queryResponse.Results[0],
		)
	}
	parameter = &queryParameter

	log.Debugf("ForemanParameter: [%+v]", parameter)

	setResourceDataFromForemanParameter(d, parameter)

	return nil
}

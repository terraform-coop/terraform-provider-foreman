package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanCommonParameter() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanCommonParameter()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the common_parameter - the full DNS common_parameter name. "+
				"%s \"dev.dc1.company.com\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanCommonParameterRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanCommonParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	commonParameter := buildForemanCommonParameter(d)

	utils.Debugf("ForemanCommonParameter: [%+v]", commonParameter)

	queryResponse, queryErr := client.QueryCommonParameter(ctx, commonParameter)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source common_parameter returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source common_parameter returned more than 1 result")
	}

	var queryCommonParameter api.ForemanCommonParameter
	var ok bool
	if queryCommonParameter, ok = queryResponse.Results[0].(api.ForemanCommonParameter); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanCommonParameter], got [%T]",
			queryResponse.Results[0],
		)
	}
	commonParameter = &queryCommonParameter

	utils.Debugf("ForemanCommonParameter: [%+v]", commonParameter)

	setResourceDataFromForemanCommonParameter(d, commonParameter)

	return nil
}

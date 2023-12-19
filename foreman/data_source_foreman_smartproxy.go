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

func dataSourceForemanSmartProxy() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanSmartProxy()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the smart proxy. "+
				"%s \"dns.dc1.company.com\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanSmartProxyRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanSmartProxyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	s := buildForemanSmartProxy(d)

	utils.Debugf("ForemanSmartProxy: [%+v]", s)

	queryResponse, queryErr := client.QuerySmartProxy(ctx, s)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source smart proxy returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source smart proxy returned more than 1 result")
	}

	var querySmartProxy api.ForemanSmartProxy
	var ok bool
	if querySmartProxy, ok = queryResponse.Results[0].(api.ForemanSmartProxy); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanSmartProxy], got [%T]",
			queryResponse.Results[0],
		)
	}
	s = &querySmartProxy

	utils.Debugf("ForemanSmartProxy: [%+v]", s)

	setResourceDataFromForemanSmartProxy(d, s)

	return nil
}

package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
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

		Read: dataSourceForemanSmartProxyRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanSmartProxyRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_smartproxy.go#Read")

	client := meta.(*api.Client)
	s := buildForemanSmartProxy(d)

	log.Debugf("ForemanSmartProxy: [%+v]", s)

	queryResponse, queryErr := client.QuerySmartProxy(s)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source smart proxy returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source smart proxy returned more than 1 result")
	}

	var querySmartProxy api.ForemanSmartProxy
	var ok bool
	if querySmartProxy, ok = queryResponse.Results[0].(api.ForemanSmartProxy); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanSmartProxy], got [%T]",
			queryResponse.Results[0],
		)
	}
	s = &querySmartProxy

	log.Debugf("ForemanSmartProxy: [%+v]", s)

	setResourceDataFromForemanSmartProxy(d, s)

	return nil
}

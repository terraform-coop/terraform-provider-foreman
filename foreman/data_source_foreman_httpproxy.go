package foreman

import (
	"fmt"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceForemanHTTPProxy() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanHTTPProxy()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the smart proxy. "+
				"%s \"proxy.company.com\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		Read: dataSourceForemanHTTPProxyRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanHTTPProxyRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_smartproxy.go#Read")

	client := meta.(*api.Client)
	s := buildForemanHTTPProxy(d)

	log.Debugf("ForemanHTTPProxy: [%+v]", s)

	queryResponse, queryErr := client.QueryHTTPProxy(s)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source smart proxy returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source smart proxy returned more than 1 result")
	}

	var queryHTTPProxy api.ForemanHTTPProxy
	var ok bool
	if queryHTTPProxy, ok = queryResponse.Results[0].(api.ForemanHTTPProxy); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanHTTPProxy], got [%T]",
			queryResponse.Results[0],
		)
	}
	s = &queryHTTPProxy

	log.Debugf("ForemanHTTPProxy: [%+v]", s)

	setResourceDataFromForemanHTTPProxy(d, s)

	return nil
}

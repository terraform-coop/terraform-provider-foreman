package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
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

		Read: dataSourceForemanEnvironmentRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_environment.go#Read")

	client := meta.(*api.Client)
	e := buildForemanEnvironment(d)

	log.Debugf("ForemanEnvironment: [%+v]", e)

	queryResponse, queryErr := client.QueryEnvironment(e)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source environment returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source environment returned more than 1 result")
	}

	var queryEnvironment api.ForemanEnvironment
	var ok bool
	if queryEnvironment, ok = queryResponse.Results[0].(api.ForemanEnvironment); !ok {
		return fmt.Errorf(
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

package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
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

		Read: dataSourceForemanParameterRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanParameterRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_parameter.go#Read")

	client := meta.(*api.Client)
	parameter := buildForemanParameter(d)

	log.Debugf("ForemanParameter: [%+v]", parameter)

	queryResponse, queryErr := client.QueryParameter(parameter)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source parameter returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source parameter returned more than 1 result")
	}

	var queryParameter api.ForemanParameter
	var ok bool
	if queryParameter, ok = queryResponse.Results[0].(api.ForemanParameter); !ok {
		return fmt.Errorf(
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

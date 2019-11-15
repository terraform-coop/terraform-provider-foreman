package foreman

import (
	"fmt"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
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

		Read: dataSourceForemanCommonParameterRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanCommonParameterRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_common_parameter.go#Read")

	client := meta.(*api.Client)
	common_parameter := buildForemanCommonParameter(d)

	log.Debugf("ForemanCommonParameter: [%+v]", common_parameter)

	queryResponse, queryErr := client.QueryCommonParameter(common_parameter)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source common_parameter returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source common_parameter returned more than 1 result")
	}

	var queryCommonParameter api.ForemanCommonParameter
	var ok bool
	if queryCommonParameter, ok = queryResponse.Results[0].(api.ForemanCommonParameter); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanCommonParameter], got [%T]",
			queryResponse.Results[0],
		)
	}
	common_parameter = &queryCommonParameter

	log.Debugf("ForemanCommonParameter: [%+v]", common_parameter)

	setResourceDataFromForemanCommonParameter(d, common_parameter)

	return nil
}

package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanModel() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanModel()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the hardware model. "+
				"%s \"PowerEdge C8220\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		Read: dataSourceForemanModelRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanModelRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_model.go#Read")

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	log.Debugf("ForemanModel: [%+v]", m)

	queryResponse, queryErr := client.QueryModel(m)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source model returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source model returned more than 1 result")
	}

	var queryModel api.ForemanModel
	var ok bool
	if queryModel, ok = queryResponse.Results[0].(api.ForemanModel); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanModel], got [%T]",
			queryResponse.Results[0],
		)
	}
	m = &queryModel

	log.Debugf("ForemanModel: [%+v]", m)

	setResourceDataFromForemanModel(d, m)

	return nil
}

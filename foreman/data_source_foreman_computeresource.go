package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanComputeResource() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanComputeResource()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: fmt.Sprintf("The name of the compute resource. %s", autodoc.MetaExample),
	}

	return &schema.Resource{

		Read: dataSourceForemanComputeResourceRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanComputeResourceRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_computeresource.go#Read")

	client := meta.(*api.Client)
	computeresource := buildForemanComputeResource(d)

	log.Debugf("ForemanComputeResource: [%+v]", computeresource)

	queryResponse, queryErr := client.QueryComputeResource(computeresource)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source computeresource returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source computeresource returned more than 1 result")
	}

	var queryComputeResource api.ForemanComputeResource
	var ok bool
	if queryComputeResource, ok = queryResponse.Results[0].(api.ForemanComputeResource); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanComputeResource], got [%T]",
			queryResponse.Results[0],
		)
	}
	computeresource = &queryComputeResource

	log.Debugf("ForemanComputeResource: [%+v]", computeresource)

	setResourceDataFromForemanComputeResource(d, computeresource)

	return nil
}

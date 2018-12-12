package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanHostgroup() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanHostgroup()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source

	ds["title"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The title is the fullname of the hostgroup.  A "+
				"hostgroup's title is a path-like string from the head "+
				"of the hostgroup tree down to this hostgroup.  The title will be "+
				"in the form of: \"<parent 1>/<parent 2>/.../<name>\". "+
				"%s \"BO1/VM/DEVP4\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		Read: dataSourceForemanHostgroupRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanHostgroupRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_hostgroup.go#Read")

	client := meta.(*api.Client)
	h := buildForemanHostgroup(d)

	log.Debugf("ForemanHostgroup: [%+v]", h)

	queryResponse, queryErr := client.QueryHostgroup(h)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source hostgroup returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source hostgroup returned more than 1 result")
	}

	var queryHostgroup api.ForemanHostgroup
	var ok bool
	if queryHostgroup, ok = queryResponse.Results[0].(api.ForemanHostgroup); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanHostgroup], got [%T]",
			queryResponse.Results[0],
		)
	}
	h = &queryHostgroup

	log.Debugf("ForemanHostgroup: [%+v]", h)

	setResourceDataFromForemanHostgroup(d, h)

	return nil
}

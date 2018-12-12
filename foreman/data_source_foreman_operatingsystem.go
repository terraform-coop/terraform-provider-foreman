package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanOperatingSystem() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanOperatingSystem()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["title"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"Title is a Foreman computed property that combines the operating "+
				"system's name, major, and minor versioning information into a single "+
				"string. "+
				"%s \"CentOS 7.5\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		Read: dataSourceForemanOperatingSystemRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanOperatingSystemRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_operatingsystem.go#Read")

	client := meta.(*api.Client)
	o := buildForemanOperatingSystem(d)

	log.Debugf("ForemanOperatingSystem: [%+v]", o)

	queryResponse, queryErr := client.QueryOperatingSystem(o)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source operating system returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source operating system returned more than 1 result")
	}

	var queryOS api.ForemanOperatingSystem
	var ok bool
	if queryOS, ok = queryResponse.Results[0].(api.ForemanOperatingSystem); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanArchitecture], got [%T]",
			queryResponse.Results[0],
		)
	}
	o = &queryOS

	log.Debugf("ForemanOperatingSystem: [%+v]", o)

	setResourceDataFromForemanOperatingSystem(d, o)

	return nil
}

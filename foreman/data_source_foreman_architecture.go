package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanArchitecture() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanArchitecture()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the architecture. "+
				"%s \"x86_64\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		Read: dataSourceForemanArchitectureRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanArchitectureRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_architecture.go#Read")

	client := meta.(*api.Client)
	arch := buildForemanArchitecture(d)

	log.Debugf("ForemanArchitecture: [%+v]", arch)

	queryResponse, queryErr := client.QueryArchitecture(arch)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source architecture returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source architecture returned more than 1 result")
	}

	var queryArch api.ForemanArchitecture
	var ok bool
	if queryArch, ok = queryResponse.Results[0].(api.ForemanArchitecture); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanArchitecture], got [%T]",
			queryResponse.Results[0],
		)
	}
	arch = &queryArch

	log.Debugf("ForemanArchitecture: [%+v]", arch)

	setResourceDataFromForemanArchitecture(d, arch)

	return nil
}

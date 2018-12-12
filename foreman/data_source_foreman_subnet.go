package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanSubnet() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanSubnet()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["network"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"Subnet network. "+
				"%s \"10.228.247.0\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		Read: dataSourceForemanSubnetRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanSubnetRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_subnet.go#Read")

	client := meta.(*api.Client)
	s := buildForemanSubnet(d)

	log.Debugf("ForemanSubnet: [%+v]", s)

	queryResponse, queryErr := client.QuerySubnet(s)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source subnet returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source subnet returned more than 1 result")
	}

	var querySubnet api.ForemanSubnet
	var ok bool
	if querySubnet, ok = queryResponse.Results[0].(api.ForemanSubnet); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanSubnet], got [%T]",
			queryResponse.Results[0],
		)
	}
	s = &querySubnet

	log.Debugf("ForemanSubnet: [%+v]", s)

	setResourceDataFromForemanSubnet(d, s)

	return nil
}

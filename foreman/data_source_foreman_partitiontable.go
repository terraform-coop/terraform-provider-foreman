package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanPartitionTable() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanPartitionTable()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the partition table. "+
				"%s \"Wayfair CentOS 7\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		Read: dataSourceForemanPartitionTableRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanPartitionTableRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_partitiontable.go#Read")

	client := meta.(*api.Client)
	t := buildForemanPartitionTable(d)

	log.Debugf("ForemanPartitionTable: [%+v]", t)

	queryResponse, queryErr := client.QueryPartitionTable(t)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source partition table returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source partition table returned more than 1 result")
	}

	var queryPartitionTable api.ForemanPartitionTable
	var ok bool
	if queryPartitionTable, ok = queryResponse.Results[0].(api.ForemanPartitionTable); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanPartitionTable], got [%T]",
			queryResponse.Results[0],
		)
	}
	t = &queryPartitionTable

	log.Debugf("[DEBUG] ForemanPartitionTable: [%+v]", t)

	setResourceDataFromForemanPartitionTable(d, t)

	return nil
}

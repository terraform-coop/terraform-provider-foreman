package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

		ReadContext: dataSourceForemanPartitionTableRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanPartitionTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	t := buildForemanPartitionTable(d)

	utils.Debugf("ForemanPartitionTable: [%+v]", t)

	queryResponse, queryErr := client.QueryPartitionTable(ctx, t)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source partition table returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source partition table returned more than 1 result")
	}

	var queryPartitionTable api.ForemanPartitionTable
	var ok bool
	if queryPartitionTable, ok = queryResponse.Results[0].(api.ForemanPartitionTable); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanPartitionTable], got [%T]",
			queryResponse.Results[0],
		)
	}
	t = &queryPartitionTable

	utils.Debugf("[DEBUG] ForemanPartitionTable: [%+v]", t)

	setResourceDataFromForemanPartitionTable(d, t)

	return nil
}

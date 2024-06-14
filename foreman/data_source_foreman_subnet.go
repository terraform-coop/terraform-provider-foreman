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

func dataSourceForemanSubnet() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanSubnet()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["network"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: fmt.Sprintf(
			"Subnet network. "+
				"%s \"10.228.247.0\"",
			autodoc.MetaExample,
		),
	}

	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: fmt.Sprintf(
			"Name of a subnetwork. "+
				"%s \"public\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanSubnetRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	s := buildForemanSubnet(d)

	utils.Debugf("ForemanSubnet: [%+v]", s)

	queryResponse, queryErr := client.QuerySubnet(ctx, s)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source subnet returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source subnet returned more than 1 result")
	}

	var querySubnet api.ForemanSubnet
	var ok bool
	if querySubnet, ok = queryResponse.Results[0].(api.ForemanSubnet); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanSubnet], got [%T]",
			queryResponse.Results[0],
		)
	}
	s = &querySubnet

	utils.Debugf("ForemanSubnet: [%+v]", s)

	setResourceDataFromForemanSubnet(d, s)

	return nil
}

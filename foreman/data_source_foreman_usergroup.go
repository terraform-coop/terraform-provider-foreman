package foreman

import (
	"context"
	"fmt"

	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanUsergroup() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanUsergroup()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source

	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the usergroup."+
				"%s \"my_usergroup\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanUsergroupRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanUsergroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_usergroup.go#Read")

	client := meta.(*api.Client)
	u := buildForemanUsergroup(d)

	log.Debugf("ForemanUsergroup: [%+v]", u)

	queryResponse, queryErr := client.QueryUsergroup(ctx, u)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source usergroup returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source usergroup returned more than 1 result")
	}

	var queryUsergroup api.ForemanUsergroup
	var ok bool
	if queryUsergroup, ok = queryResponse.Results[0].(api.ForemanUsergroup); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanUsergroup], got [%T]",
			queryResponse.Results[0],
		)
	}
	u = &queryUsergroup

	log.Debugf("ForemanUsergroup: [%+v]", u)

	setResourceDataFromForemanUsergroup(d, u)

	return nil
}

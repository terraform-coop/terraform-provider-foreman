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

		ReadContext: dataSourceForemanOperatingSystemRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanOperatingSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	o := buildForemanOperatingSystem(d)

	utils.Debugf("ForemanOperatingSystem: [%+v]", o)

	queryResponse, queryErr := client.QueryOperatingSystem(ctx, o)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source operating system returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source operating system returned more than 1 result")
	}

	var queryOS api.ForemanOperatingSystem
	var ok bool
	if queryOS, ok = queryResponse.Results[0].(api.ForemanOperatingSystem); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanArchitecture], got [%T]",
			queryResponse.Results[0],
		)
	}
	o = &queryOS

	utils.Debugf("ForemanOperatingSystem: [%+v]", o)

	setResourceDataFromForemanOperatingSystem(d, o)

	return nil
}

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

		ReadContext: dataSourceForemanArchitectureRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanArchitectureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	arch := buildForemanArchitecture(d)

	utils.Debugf("ForemanArchitecture: [%+v]", arch)

	queryResponse, queryErr := client.QueryArchitecture(ctx, arch)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source architecture returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source architecture returned more than 1 result")
	}

	var queryArch api.ForemanArchitecture
	var ok bool
	if queryArch, ok = queryResponse.Results[0].(api.ForemanArchitecture); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanArchitecture], got [%T]",
			queryResponse.Results[0],
		)
	}
	arch = &queryArch

	utils.Debugf("ForemanArchitecture: [%+v]", arch)

	setResourceDataFromForemanArchitecture(d, arch)

	return nil
}

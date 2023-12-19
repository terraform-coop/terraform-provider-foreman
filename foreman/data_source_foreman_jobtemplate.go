package foreman

import (
	"context"
	"fmt"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"

	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
)

func dataSourceForemanJobTemplate() *schema.Resource {
	r := resourceForemanJobTemplate()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"job template name. %s \"change content sources\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{
		ReadContext: dataSourceForemanJobTemplateRead,
		Schema:      ds,
	}
}

func dataSourceForemanJobTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	jt := buildForemanJobTemplate(d)

	queryResponse, err := client.QueryJobTemplate(ctx, jt)
	if err != nil {
		return diag.FromErr(err)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source job_template returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source job_template returned more than 1 result")
	}

	var queryJt api.ForemanJobTemplate
	var ok bool
	if queryJt, ok = queryResponse.Results[0].(api.ForemanJobTemplate); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanJobTemplate], got [%T]",
			queryResponse.Results[0],
		)
	}

	log.Debugf("ForemanJobTemplate: [%+v]", queryJt)

	setResourceDataFromForemanJobTemplate(d, &queryJt)

	return nil
}

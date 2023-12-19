package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanModel() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanModel()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the hardware model. "+
				"%s \"PowerEdge C8220\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanModelRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	log.Debugf("ForemanModel: [%+v]", m)

	queryResponse, queryErr := client.QueryModel(ctx, m)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source model returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source model returned more than 1 result")
	}

	var queryModel api.ForemanModel
	var ok bool
	if queryModel, ok = queryResponse.Results[0].(api.ForemanModel); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanModel], got [%T]",
			queryResponse.Results[0],
		)
	}
	m = &queryModel

	log.Debugf("ForemanModel: [%+v]", m)

	setResourceDataFromForemanModel(d, m)

	return nil
}

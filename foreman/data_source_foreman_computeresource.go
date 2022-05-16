package foreman

import (
	"context"
	"fmt"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanComputeResource() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanComputeResource()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: fmt.Sprintf("The name of the compute resource. %s", autodoc.MetaExample),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanComputeResourceRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanComputeResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_computeresource.go#Read")

	client := meta.(*api.Client)
	computeresource := buildForemanComputeResource(d)

	log.Debugf("ForemanComputeResource: [%+v]", computeresource)

	queryResponse, queryErr := client.QueryComputeResource(ctx, computeresource)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source computeresource returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source computeresource returned more than 1 result")
	}

	var queryComputeResource api.ForemanComputeResource
	var ok bool
	if queryComputeResource, ok = queryResponse.Results[0].(api.ForemanComputeResource); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanComputeResource], got [%T]",
			queryResponse.Results[0],
		)
	}
	computeresource = &queryComputeResource

	log.Debugf("ForemanComputeResource: [%+v]", computeresource)

	setResourceDataFromForemanComputeResource(d, computeresource)

	return nil
}

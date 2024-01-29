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

func dataSourceForemanImage() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanImage()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: fmt.Sprintf("The name of the compute resource. %s", autodoc.MetaExample),
	}
	ds["compute_resource_id"] = &schema.Schema{
		Type:        schema.TypeInt,
		Required:    true,
		Description: "The id of the Compute Resource the image is associated with",
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanImageRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	image := buildForemanImage(d)

	utils.Debugf("ForemanImage: [%+v]", image)

	queryResponse, queryErr := client.QueryImage(ctx, image)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source image returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source image returned more than 1 result")
	}

	var queryImage api.ForemanImage
	var ok bool
	if queryImage, ok = queryResponse.Results[0].(api.ForemanImage); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanImage], got [%T]",
			queryResponse.Results[0],
		)
	}
	image = &queryImage

	utils.Debugf("ForemanImage: [%+v]", image)

	setResourceDataFromForemanImage(d, image)

	return nil
}

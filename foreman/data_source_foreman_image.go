package foreman

import (
	"fmt"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"

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
		Description: fmt.Sprintf("The id of the Compute Resource the image is associated with"),
	}

	return &schema.Resource{

		Read: dataSourceForemanImageRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanImageRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_image.go#Read")

	client := meta.(*api.Client)
	image := buildForemanImage(d)

	log.Debugf("ForemanImage: [%+v]", image)

	queryResponse, queryErr := client.QueryImage(image)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source image returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source image returned more than 1 result")
	}

	var queryImage api.ForemanImage
	var ok bool
	if queryImage, ok = queryResponse.Results[0].(api.ForemanImage); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanImage], got [%T]",
			queryResponse.Results[0],
		)
	}
	image = &queryImage

	log.Debugf("ForemanImage: [%+v]", image)

	setResourceDataFromForemanImage(d, image)

	return nil
}

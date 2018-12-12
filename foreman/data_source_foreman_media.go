package foreman

import (
	"fmt"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/helper"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanMedia() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanMedia()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"Name of the media. "+
				"%s \"Debian mirror\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		Read: dataSourceForemanMediaRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanMediaRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_media.go#Read")

	client := meta.(*api.Client)
	m := buildForemanMedia(d)

	log.Debugf("ForemanMedia: [%+v]", m)

	queryResponse, queryErr := client.QueryMedia(m)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source media returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source media returned more than 1 result")
	}

	var queryMedia api.ForemanMedia
	var ok bool
	if queryMedia, ok = queryResponse.Results[0].(api.ForemanMedia); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanMedia], got [%T]",
			queryResponse.Results[0],
		)
	}
	m = &queryMedia

	log.Debugf("ForemanMedia: [%+v]", m)

	setResourceDataFromForemanMedia(d, m)

	return nil
}

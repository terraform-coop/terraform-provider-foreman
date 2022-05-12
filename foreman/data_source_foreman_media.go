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

		ReadContext: dataSourceForemanMediaRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanMediaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_media.go#Read")

	client := meta.(*api.Client)
	m := buildForemanMedia(d)

	log.Debugf("ForemanMedia: [%+v]", m)

	queryResponse, queryErr := client.QueryMedia(ctx, m)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source media returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source media returned more than 1 result")
	}

	var queryMedia api.ForemanMedia
	var ok bool
	if queryMedia, ok = queryResponse.Results[0].(api.ForemanMedia); !ok {
		return diag.Errorf(
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

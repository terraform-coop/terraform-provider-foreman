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

func dataSourceForemanDefaultTemplate() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanDefaultTemplate()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the defaultTemplate - the full DNS defaultTemplate name. "+
				"%s \"dev.dc1.company.com\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanDefaultTemplateRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanDefaultTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_defaultTemplate.go#Read")

	client := meta.(*api.Client)
	defaultTemplate := buildForemanDefaultTemplate(d)

	log.Debugf("ForemanDefaultTemplate: [%+v]", defaultTemplate)

	queryResponse, queryErr := client.QueryDefaultTemplate(ctx, defaultTemplate)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source defaultTemplate returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source defaultTemplate returned more than 1 result")
	}

	var queryDefaultTemplate api.ForemanDefaultTemplate
	var ok bool
	if queryDefaultTemplate, ok = queryResponse.Results[0].(api.ForemanDefaultTemplate); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanDefaultTemplate], got [%T]",
			queryResponse.Results[0],
		)
	}
	defaultTemplate = &queryDefaultTemplate

	log.Debugf("ForemanDefaultTemplate: [%+v]", defaultTemplate)

	setResourceDataFromForemanDefaultTemplate(d, defaultTemplate)

	return nil
}

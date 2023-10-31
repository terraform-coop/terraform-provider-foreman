package foreman

import (
	"context"

	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
)

func dataSourceForemanJobTemplate() *schema.Resource {
	r := resourceForemanJobTemplate()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	return &schema.Resource{
		ReadContext: dataSourceForemanJobTemplateRead,
		Schema:      ds,
	}
}

func dataSourceForemanJobTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	TraceFunctionCall()

	client := meta.(*api.Client)
	jt := buildForemanJobTemplate(d)

	readResponse, readErr := client.ReadJobTemplate(ctx, jt.Id)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	jt.Name = readResponse.Name

	log.Debugf("ForemanJobTemplate: [%+v]", jt)

	setResourceDataFromForemanJobTemplate(d, jt)

	return nil
}

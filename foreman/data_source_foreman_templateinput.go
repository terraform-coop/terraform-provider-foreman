package foreman

import (
	"context"

	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
)

func dataSourceForemanTemplateInput() *schema.Resource {
	r := resourceForemanTemplateInput()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	return &schema.Resource{
		ReadContext: dataSourceForemanTemplateInputRead,
		Schema:      ds,
	}
}

func dataSourceForemanTemplateInputRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	built := buildForemanTemplateInput(d)

	readResponse, readErr := client.ReadTemplateInput(ctx, built)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	built.Name = readResponse.Name

	utils.Debugf("ForemanTemplateInput: [%+v]", built)

	setResourceDataFromForemanTemplateInput(d, built)

	return nil
}

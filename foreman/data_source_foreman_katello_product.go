package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
)

func dataSourceForemanKatelloProduct() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanKatelloProduct()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"Product name."+
				"%s \"Debian 10\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanKatelloProductRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanKatelloProductRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	product := buildForemanKatelloProduct(d)

	utils.Debugf("ForemanKatelloProduct: [%+v]", product)

	queryResponse, queryErr := client.QueryKatelloProduct(ctx, product)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("data source product returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("data source product returned more than 1 result")
	}

	var queryKatelloProduct api.ForemanKatelloProduct
	var ok bool
	if queryKatelloProduct, ok = queryResponse.Results[0].(api.ForemanKatelloProduct); !ok {
		return diag.Errorf(
			"data source results contain unexpected type. Expected "+
				"[api.ForemanKatelloProduct], got [%T]",
			queryResponse.Results[0],
		)
	}
	product = &queryKatelloProduct

	utils.Debugf("ForemanKatelloProduct: [%+v]", product)

	setResourceDataFromForemanKatelloProduct(d, product)

	return nil
}

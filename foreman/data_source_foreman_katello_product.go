package foreman

import (
	"fmt"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

		Read: dataSourceForemanKatelloProductRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanKatelloProductRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_katello_product.go#Read")

	client := meta.(*api.Client)
	product := buildForemanKatelloProduct(d)

	log.Debugf("ForemanKatelloProduct: [%+v]", product)

	queryResponse, queryErr := client.QueryKatelloProduct(product)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("data source product returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("data source product returned more than 1 result")
	}

	var queryKatelloProduct api.ForemanKatelloProduct
	var ok bool
	if queryKatelloProduct, ok = queryResponse.Results[0].(api.ForemanKatelloProduct); !ok {
		return fmt.Errorf(
			"data source results contain unexpected type. Expected "+
				"[api.ForemanKatelloProduct], got [%T]",
			queryResponse.Results[0],
		)
	}
	product = &queryKatelloProduct

	log.Debugf("ForemanKatelloProduct: [%+v]", product)

	setResourceDataFromForemanKatelloProduct(d, product)

	return nil
}

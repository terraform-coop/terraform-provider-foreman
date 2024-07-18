package foreman

import (
	"context"
	"fmt"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
)

func dataSourceForemanKatelloContentView() *schema.Resource {
	r := resourceForemanKatelloContentView()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: fmt.Sprintf("Name of the content view. %s \"my content view\"", autodoc.MetaExample),
	}

	return &schema.Resource{
		ReadContext: dataSourceForemanKatelloContentViewRead,
		Schema:      ds,
	}
}

func dataSourceForemanKatelloContentViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	cv := buildForemanKatelloContentView(d)

	utils.Debugf("cv: %+v", cv)

	queryResponse, err := client.QueryContentView(ctx, cv)
	if err != nil {
		return diag.FromErr(err)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("data source content_view returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("data source content_view returned more than 1 result")
	}

	if queryCv, ok := queryResponse.Results[0].(api.ContentView); !ok {
		return diag.Errorf(
			"data source results contain unexpected type. Expected "+
				"[api.ContentView], got [%T]",
			queryResponse.Results[0],
		)
	} else {
		cv = &queryCv
	}

	filtersResult, err := client.QueryContentViewFilters(ctx, cv.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, item := range filtersResult.Results {
		asserted := item.(api.ContentViewFilter)
		cv.Filters = append(cv.Filters, asserted)
		utils.Debugf("%+v", asserted)
	}

	utils.Debugf("cv: %+v", cv)

	setResourceDataFromForemanKatelloContentView(d, cv)

	return nil
}

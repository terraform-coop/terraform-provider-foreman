package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanDomain() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanDomain()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"The name of the domain - the full DNS domain name. "+
				"%s \"dev.dc1.company.com\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanDomainRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	domain := buildForemanDomain(d)

	log.Debugf("ForemanDomain: [%+v]", domain)

	queryResponse, queryErr := client.QueryDomain(ctx, domain)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source domain returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source domain returned more than 1 result")
	}

	var queryDomain api.ForemanDomain
	var ok bool
	if queryDomain, ok = queryResponse.Results[0].(api.ForemanDomain); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanDomain], got [%T]",
			queryResponse.Results[0],
		)
	}
	domain = &queryDomain

	log.Debugf("ForemanDomain: [%+v]", domain)

	setResourceDataFromForemanDomain(d, domain)

	return nil
}

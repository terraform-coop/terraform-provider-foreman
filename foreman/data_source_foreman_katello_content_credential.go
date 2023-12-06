package foreman

import (
	"context"
	"fmt"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanKatelloContentCredential() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanKatelloContentCredential()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"Identifier of the content credential."+
				"%s \"RPM-GPG-KEY-centos7\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanKatelloContentCredentialRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanKatelloContentCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_katello_content_credential.go#Read")

	client := meta.(*api.Client)
	contentCredential := buildForemanKatelloContentCredential(d)

	log.Debugf("ForemanKatelloContentCredential: [%+v]", contentCredential)

	queryResponse, queryErr := client.QueryKatelloContentCredential(ctx, contentCredential)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source smart proxy returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source smart proxy returned more than 1 result")
	}

	var queryKatelloContentCredential api.ForemanKatelloContentCredential
	var ok bool
	if queryKatelloContentCredential, ok = queryResponse.Results[0].(api.ForemanKatelloContentCredential); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanKatelloContentCredential], got [%T]",
			queryResponse.Results[0],
		)
	}
	contentCredential = &queryKatelloContentCredential

	log.Debugf("ForemanKatelloContentCredential: [%+v]", contentCredential)

	setResourceDataFromForemanKatelloContentCredential(d, contentCredential)

	return nil
}

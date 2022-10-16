package foreman

import (
	"context"
	"fmt"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
)

func dataSourceForemanKatelloRepository() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanKatelloRepository()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		Description: fmt.Sprintf(
			"Repository name."+
				"%s \"centos7-base\"",
			autodoc.MetaExample,
		),
	}

	return &schema.Resource{

		ReadContext: dataSourceForemanKatelloRepositoryRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanKatelloRepositoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_katello_repository.go#Read")

	client := meta.(*api.Client)
	repository := buildForemanKatelloRepository(d)

	log.Debugf("ForemanKatelloRepository: [%+v]", repository)

	queryResponse, queryErr := client.QueryKatelloRepository(ctx, repository)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("data source repository returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("data source repository returned more than 1 result")
	}

	var queryKatelloRepository api.ForemanKatelloRepository
	var ok bool
	if queryKatelloRepository, ok = queryResponse.Results[0].(api.ForemanKatelloRepository); !ok {
		return diag.Errorf(
			"data source results contain unexpected type. Expected "+
				"[api.ForemanKatelloRepository], got [%T]",
			queryResponse.Results[0],
		)
	}
	repository = &queryKatelloRepository

	log.Debugf("ForemanKatelloRepository: [%+v]", repository)

	setResourceDataFromForemanKatelloRepository(d, repository)

	return nil
}

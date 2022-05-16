package foreman

import (
	"fmt"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/helper"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanUser() *schema.Resource {
	// copy attributes from resource definition
	r := resourceForemanUser()
	ds := helper.DataSourceSchemaFromResourceSchema(r.Schema)

	// define searchable attributes for the data source
	ds["description"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: fmt.Sprintf(
			"User description. "+
				"%s \"api user\"",
			autodoc.MetaExample,
		),
	}

	ds["firstname"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: fmt.Sprintf(
			"Firstname of the user."+
				"%s \"Louis\"",
			autodoc.MetaExample,
		),
	}

	ds["lastname"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: fmt.Sprintf(
			"Lastname of the user."+
				"%s \"Jansens\"",
			autodoc.MetaExample,
		),
	}

	ds["login"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: fmt.Sprintf(
			"loginname of the user."+
				"%s \"username\"",
			autodoc.MetaExample,
		),
	}

	ds["mail"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: fmt.Sprintf(
			"email of the user."+
				"%s \"test@example.com\"",
			autodoc.MetaExample,
		),
	}
	return &schema.Resource{

		Read: dataSourceForemanUserRead,

		// NOTE(ALL): See comments in the corresponding resource file
		Schema: ds,
	}
}

func dataSourceForemanUserRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_user.go#Read")

	client := meta.(*api.Client)
	s := buildForemanUser(d)

	log.Debugf("ForemanUser: [%+v]", s)

	queryResponse, queryErr := client.QueryUser(s)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source user returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source user returned more than 1 result")
	}

	var queryUser api.ForemanUser
	var ok bool
	if queryUser, ok = queryResponse.Results[0].(api.ForemanUser); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanUser], got [%T]",
			queryResponse.Results[0],
		)
	}
	s = &queryUser

	log.Debugf("ForemanUser: [%+v]", s)

	setResourceDataFromForemanUser(d, s)

	return nil
}

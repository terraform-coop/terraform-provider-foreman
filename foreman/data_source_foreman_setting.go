package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
)

func dataSourceForemanSetting() *schema.Resource {
	// Build schema from scratch, because Setting is a data source only.
	// There is no resource schema we could re-use.

	dataSourceSchema := map[string]*schema.Schema{

		autodoc.MetaAttribute: {
			Type:     schema.TypeBool,
			Computed: true,
			Description: fmt.Sprintf(
				"%s Setting can be used to read settings from Foreman.",
				autodoc.MetaSummary,
			),
		},

		"name": {
			Type:     schema.TypeString,
			Required: true,
			Description: fmt.Sprintf(
				"Name of the setting"+
					"%s \"foreman_url\"",
				autodoc.MetaExample,
			),
		},

		// Value can be either string, int or bool.
		// The API client converts int and bool to string..
		"value": {
			Type:     schema.TypeString,
			Computed: true,
			Description: fmt.Sprintf(
				"Value of the setting"+
					"%s \"https://foreman.company.com\"",
				autodoc.MetaExample,
			),
		},

		"default": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Default value of the setting",
		},

		"readonly": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Indicates whether the setting is read-only or not.",
		},

		"description": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Description of the setting",
		},

		"category_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the category the setting is in.",
		},

		"settings_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Data type of this setting (boolean, string, ..)",
		},
	}

	return &schema.Resource{
		ReadContext: dataSourceForemanSettingRead,
		Schema:      dataSourceSchema,
	}
}

func dataSourceForemanSettingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_setting.go#Read")

	var diags diag.Diagnostics

	client := meta.(*api.Client)
	setting := &api.ForemanSetting{}

	// Build basic Foreman object inside struct
	obj := buildForemanObject(d)
	setting.ForemanObject = *obj

	log.Debugf("ForemanSetting: [%+v]", setting)

	queryResponse, queryErr := client.QuerySetting(ctx, setting)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source setting returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source setting returned more than 1 result")
	}

	var querySetting api.ForemanSetting
	var ok bool
	if querySetting, ok = queryResponse.Results[0].(api.ForemanSetting); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanSetting], got [%T]",
			queryResponse.Results[0],
		)
	}
	setting = &querySetting

	// Convert boolean or integer values to strings to match the schema
	switch setting.Value.(type) {
	case bool:
		setting.Value = strconv.FormatBool(setting.Value.(bool))
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("The value for setting %s was a boolean and was converted to string", setting.Name),
		})
	case int:
		setting.Value = strconv.FormatInt(setting.Value.(int64), 10)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("The value for setting %s was an integer and was converted to string", setting.Name),
		})
	case string:
	default:
		// noop
	}

	log.Debugf("ForemanSetting: [%+v]", setting)

	d.SetId(setting.Id)
	d.Set("name", setting.Name)
	d.Set("value", setting.Value)
	d.Set("description", setting.Description)
	d.Set("default", setting.Default)
	d.Set("category_name", setting.CategoryName)
	d.Set("readonly", setting.ReadOnly)
	d.Set("settings_type", setting.SettingsType)

	return diags
}

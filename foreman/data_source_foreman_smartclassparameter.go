package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanSmartClassParameter() *schema.Resource {
	return &schema.Resource{

		ReadContext: dataSourceForemanSmartClassParameterRead,

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of a smart class parameter.",
					autodoc.MetaSummary,
				),
			},

			"parameter": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Smart class parameter name."+
						"%s \"example_param\"",
					autodoc.MetaExample,
				),
			},

			"puppetclass_id": {
				Type:     schema.TypeInt,
				Required: true,
				Description: fmt.Sprintf(
					"ID of the puppet class containing this parameter."+
						"%s 1",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanSmartClassParameter constructs a ForemanSmartClassParameter reference from a
// resource data reference.  The struct's  members are populated from the data
// populated in the resource data.  Missing members will be left to the zero
// value for that member's type.
func buildForemanSmartClassParameter(d *schema.ResourceData) *api.ForemanSmartClassParameter {
	t := api.ForemanSmartClassParameter{}
	obj := buildForemanObject(d)
	t.ForemanObject = *obj

	t.Parameter = d.Get("parameter").(string)
	t.PuppetClassId = d.Get("puppetclass_id").(int)

	return &t
}

// setResourceDataFromForemanSmartClassParameter sets a ResourceData's attributes from
// the attributes of the supplied ForemanSmartClassParameter reference
func setResourceDataFromForemanSmartClassParameter(d *schema.ResourceData, fk *api.ForemanSmartClassParameter) {
	d.SetId(strconv.Itoa(fk.Id))
	d.Set("parameter", fk.Parameter)
	d.Set("puppetclass_id", fk.PuppetClassId)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func dataSourceForemanSmartClassParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_smartclassparameter.go#Read")

	client := meta.(*api.Client)
	t := buildForemanSmartClassParameter(d)

	log.Debugf("ForemanSmartClassParameter: [%+v]", t)

	queryResponse, queryErr := client.QuerySmartClassParameter(ctx, t)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source smart class parameter returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source smart class parameter returned more than 1 result")
	}

	var querySmartClassParameter api.ForemanSmartClassParameter
	var ok bool
	if querySmartClassParameter, ok = queryResponse.Results[0].(api.ForemanSmartClassParameter); !ok {
		return diag.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanSmartClassParameter], got [%T]",
			queryResponse.Results[0],
		)
	}
	t = &querySmartClassParameter

	log.Debugf("ForemanSmartClassParameter: [%+v]", t)

	setResourceDataFromForemanSmartClassParameter(d, t)

	return nil
}

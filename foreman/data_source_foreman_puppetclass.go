package foreman

import (
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceForemanPuppetClass() *schema.Resource {
	return &schema.Resource{

		Read: dataSourceForemanPuppetClassRead,

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of a Puppet class.",
					autodoc.MetaSummary,
				),
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Puppet class name."+
						"%s \"example_class\"",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanPuppetClass constructs a ForemanPuppetClass reference from a
// resource data reference.  The struct's  members are populated from the data
// populated in the resource data.  Missing members will be left to the zero
// value for that member's type.
func buildForemanPuppetClass(d *schema.ResourceData) *api.ForemanPuppetClass {
	t := api.ForemanPuppetClass{}
	obj := buildForemanObject(d)
	t.ForemanObject = *obj
	return &t
}

// setResourceDataFromForemanPuppetClass sets a ResourceData's attributes from
// the attributes of the supplied ForemanPuppetClass reference
func setResourceDataFromForemanPuppetClass(d *schema.ResourceData, fk *api.ForemanPuppetClass) {
	d.SetId(strconv.Itoa(fk.Id))
	d.Set("name", fk.Name)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func dataSourceForemanPuppetClassRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_puppetclass.go#Read")

	client := meta.(*api.Client)
	t := buildForemanPuppetClass(d)

	log.Debugf("ForemanPuppetClass: [%+v]", t)

	queryResponse, queryErr := client.QueryPuppetClass(t)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source puppet class returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source puppet class returned more than 1 result")
	}

	var queryPuuppetClass api.ForemanPuppetClass
	var ok bool
	if queryPuuppetClass, ok = queryResponse.Results[0].(api.ForemanPuppetClass); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanPuppetClass], got [%T]",
			queryResponse.Results[0],
		)
	}
	t = &queryPuuppetClass

	log.Debugf("ForemanPuppetClass: [%+v]", t)

	setResourceDataFromForemanPuppetClass(d, t)

	return nil
}

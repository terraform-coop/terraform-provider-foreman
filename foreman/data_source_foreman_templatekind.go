package foreman

import (
	"fmt"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanTemplateKind() *schema.Resource {
	return &schema.Resource{

		Read: dataSourceForemanTemplateKindRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Type of template. "+
						"%s \"PXELinux\"",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanTemplateKind constructs a ForemanTemplateKind reference from a
// resource data reference.  The struct's  members are populated from the data
// populated in the resource data.  Missing members will be left to the zero
// value for that member's type.
func buildForemanTemplateKind(d *schema.ResourceData) *api.ForemanTemplateKind {
	t := api.ForemanTemplateKind{}
	obj := buildForemanObject(d)
	t.ForemanObject = *obj
	return &t
}

// setResourceDataFromForemanTemplateKind sets a ResourceData's attributes from
// the attributes of the supplied ForemanTemplateKind reference
func setResourceDataFromForemanTemplateKind(d *schema.ResourceData, fk *api.ForemanTemplateKind) {
	d.SetId(strconv.Itoa(fk.Id))
	d.Set("name", fk.Name)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func dataSourceForemanTemplateKindRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_architecture.go#Read")

	client := meta.(*api.Client)
	t := buildForemanTemplateKind(d)

	log.Debugf("ForemanTemplateKind: [%+v]", t)

	queryResponse, queryErr := client.QueryTemplateKind(t)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source template kind returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source template kind returned more than 1 result")
	}

	var queryTemplateKind api.ForemanTemplateKind
	var ok bool
	if queryTemplateKind, ok = queryResponse.Results[0].(api.ForemanTemplateKind); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanTemplateKind], got [%T]",
			queryResponse.Results[0],
		)
	}
	t = &queryTemplateKind

	log.Debugf("ForemanTemplateKind: [%+v]", t)

	setResourceDataFromForemanTemplateKind(d, t)

	return nil
}

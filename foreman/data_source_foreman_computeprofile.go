package foreman

import (
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceForemanComputeProfile() *schema.Resource {
	return &schema.Resource{

		Read: dataSourceForemanComputeProfileRead,

		Schema: map[string]*schema.Schema{

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Compute Profile",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanComputeProfile constructs a ForemanComputeProfile reference from a
// resource data reference.  The struct's  members are populated from the data
// populated in the resource data.  Missing members will be left to the zero
// value for that member's type.
func buildForemanComputeProfile(d *schema.ResourceData) *api.ForemanComputeProfile {
	t := api.ForemanComputeProfile{}
	obj := buildForemanObject(d)
	t.ForemanObject = *obj
	return &t
}

// setResourceDataFromForemanComputeProfile sets a ResourceData's attributes from
// the attributes of the supplied ForemanComputeProfile reference
func setResourceDataFromForemanComputeProfile(d *schema.ResourceData, fk *api.ForemanComputeProfile) {
	d.SetId(strconv.Itoa(fk.Id))
	d.Set("name", fk.Name)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func dataSourceForemanComputeProfileRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("data_source_foreman_architecture.go#Read")

	client := meta.(*api.Client)
	t := buildForemanComputeProfile(d)

	log.Debugf("ForemanComputeProfile: [%+v]", t)

	queryResponse, queryErr := client.QueryComputeProfile(t)
	if queryErr != nil {
		return queryErr
	}

	if queryResponse.Subtotal == 0 {
		return fmt.Errorf("Data source template kind returned no results")
	} else if queryResponse.Subtotal > 1 {
		return fmt.Errorf("Data source template kind returned more than 1 result")
	}

	var queryComputeProfile api.ForemanComputeProfile
	var ok bool
	if queryComputeProfile, ok = queryResponse.Results[0].(api.ForemanComputeProfile); !ok {
		return fmt.Errorf(
			"Data source results contain unexpected type. Expected "+
				"[api.ForemanComputeProfile], got [%T]",
			queryResponse.Results[0],
		)
	}
	t = &queryComputeProfile

	log.Debugf("ForemanComputeProfile: [%+v]", t)

	setResourceDataFromForemanComputeProfile(d, t)

	return nil
}

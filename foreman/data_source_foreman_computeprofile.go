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

func dataSourceForemanComputeProfile() *schema.Resource {
	return &schema.Resource{

		ReadContext: dataSourceForemanComputeProfileRead,

		Schema: map[string]*schema.Schema{

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Compute profile name."+
						"%s \"2-Medium\"",
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

func dataSourceForemanComputeProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_architecture.go#Read")

	client := meta.(*api.Client)
	t := buildForemanComputeProfile(d)

	log.Debugf("ForemanComputeProfile: [%+v]", t)

	queryResponse, queryErr := client.QueryComputeProfile(ctx, t)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source template kind returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source template kind returned more than 1 result")
	}

	var queryComputeProfile api.ForemanComputeProfile
	var ok bool
	if queryComputeProfile, ok = queryResponse.Results[0].(api.ForemanComputeProfile); !ok {
		return diag.Errorf(
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

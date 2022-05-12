package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceForemanTemplateKind() *schema.Resource {
	return &schema.Resource{

		ReadContext: dataSourceForemanTemplateKindRead,

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

func dataSourceForemanTemplateKindRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("data_source_foreman_architecture.go#Read")

	client := meta.(*api.Client)
	t := buildForemanTemplateKind(d)

	log.Debugf("ForemanTemplateKind: [%+v]", t)

	queryResponse, queryErr := client.QueryTemplateKind(ctx, t)
	if queryErr != nil {
		return diag.FromErr(queryErr)
	}

	if queryResponse.Subtotal == 0 {
		return diag.Errorf("Data source template kind returned no results")
	} else if queryResponse.Subtotal > 1 {
		return diag.Errorf("Data source template kind returned more than 1 result")
	}

	var queryTemplateKind api.ForemanTemplateKind
	var ok bool
	if queryTemplateKind, ok = queryResponse.Results[0].(api.ForemanTemplateKind); !ok {
		return diag.Errorf(
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

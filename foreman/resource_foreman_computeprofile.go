package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
)

func resourceForemanComputeProfile() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanComputeprofileCreate,
		ReadContext:   resourceForemanComputeprofileRead,
		UpdateContext: resourceForemanComputeprofileUpdate,
		DeleteContext: resourceForemanComputeprofileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of a compute profile.",
					autodoc.MetaSummary,
				),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the compute profile",
			},
			"compute_attributes": {
				Type:             schema.TypeString,
				ValidateFunc:     validation.StringIsJSON,
				Optional:         true,
				Computed:         true,
				Description:      "Hypervisor specific VM options. Must be a JSON string, as every compute provider has different attributes schema",
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
			// "compute_attributes": {
			// 	Type: schema.TypeList,
			// 	Required: true,
			// 	Description: "List of compute attributes",
			// 	Elem: &schema.Schema{
			// 	},
			// },
		},
	}
}

// buildForemanComputeProfile constructs a ForemanComputeProfile reference from a
// resource data reference.  The struct's  members are populated from the data
// populated in the resource data.  Missing members will be left to the zero
// value for that member's type.
func buildForemanComputeProfile(d *schema.ResourceData) *api.ForemanComputeProfile {
	log.Tracef("foreman/resource_foreman_computeprofile.go#buildForemanComputeProfile")

	t := api.ForemanComputeProfile{}
	obj := buildForemanObject(d)
	t.ForemanObject = *obj
	t.ComputeAttributes = d.Get("compute_attributes").(string)
	return &t
}

// setResourceDataFromForemanComputeProfile sets a ResourceData's attributes from
// the attributes of the supplied ForemanComputeProfile reference
func setResourceDataFromForemanComputeProfile(d *schema.ResourceData, fk *api.ForemanComputeProfile) {
	log.Tracef("foreman/resource_foreman_computeprofile.go#setResourceDataFromForemanComputeProfile")

	d.SetId(strconv.Itoa(fk.Id))
	d.Set("name", fk.Name)
}

func resourceForemanComputeprofileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("foreman/resource_foreman_computeprofile.go#resourceForemanComputeprofileCreate")

	client := meta.(*api.Client)
	p := buildForemanComputeProfile(d)

	createdComputeprofile, createErr := client.CreateComputeprofile(ctx, p)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanComputeprofile [%+v]", createdComputeprofile)

	setResourceDataFromForemanComputeProfile(d, createdComputeprofile)

	return nil
}

func resourceForemanComputeprofileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("foreman/resource_foreman_computeprofile.go#resourceForemanComputeprofileRead")
	return nil
}

func resourceForemanComputeprofileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("foreman/resource_foreman_computeprofile.go#resourceForemanComputeprofileUpdate")
	return nil
}

func resourceForemanComputeprofileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("foreman/resource_foreman_computeprofile.go#resourceForemanComputeprofileDelete")
	return nil
}

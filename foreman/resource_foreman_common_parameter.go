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

func resourceForemanCommonParameter() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanCommonParameterCreate,
		ReadContext:   resourceForemanCommonParameterRead,
		UpdateContext: resourceForemanCommonParameterUpdate,
		DeleteContext: resourceForemanCommonParameterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of common_parameter. Global parameters are available for all resources",
					autodoc.MetaSummary,
				),
			},

			// -- Actual Content --
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanCommonParameter constructs a ForemanCommonParameter reference from a resource data
// reference.  The struct's  members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanCommonParameter(d *schema.ResourceData) *api.ForemanCommonParameter {
	log.Tracef("resource_foreman_common_parameter.go#buildForemanCommonParameter")

	commonParameter := api.ForemanCommonParameter{}

	obj := buildForemanObject(d)
	commonParameter.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		commonParameter.Name = attr.(string)
	}
	if attr, ok = d.GetOk("value"); ok {
		commonParameter.Value = attr.(string)
	}
	return &commonParameter
}

// setResourceDataFromForemanCommonParameter sets a ResourceData's attributes from the
// attributes of the supplied ForemanCommonParameter reference
func setResourceDataFromForemanCommonParameter(d *schema.ResourceData, fd *api.ForemanCommonParameter) {
	log.Tracef("resource_foreman_common_parameter.go#setResourceDataFromForemanCommonParameter")

	d.SetId(strconv.Itoa(fd.Id))
	d.Set("name", fd.Name)
	d.Set("value", fd.Value)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanCommonParameterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_common_parameter.go#Create")

	client := meta.(*api.Client)
	p := buildForemanCommonParameter(d)

	log.Debugf("ForemanCommonParameter: [%+v]", d)

	createdParam, createErr := client.CreateCommonParameter(ctx, p)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanCommonParameter: [%+v]", createdParam)

	setResourceDataFromForemanCommonParameter(d, createdParam)

	return nil
}

func resourceForemanCommonParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_common_parameter.go#Read")

	client := meta.(*api.Client)
	commonParameter := buildForemanCommonParameter(d)

	log.Debugf("ForemanCommonParameter: [%+v]", commonParameter)

	readCommonParameter, readErr := client.ReadCommonParameter(ctx, commonParameter, commonParameter.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanCommonParameter: [%+v]", readCommonParameter)

	setResourceDataFromForemanCommonParameter(d, readCommonParameter)

	return nil
}

func resourceForemanCommonParameterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_common_parameter.go#Update")

	client := meta.(*api.Client)
	p := buildForemanCommonParameter(d)

	log.Debugf("ForemanCommonParameter: [%+v]", p)

	updatedParam, updateErr := client.UpdateCommonParameter(ctx, p, p.Id)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanCommonParameter: [%+v]", updatedParam)

	setResourceDataFromForemanCommonParameter(d, updatedParam)

	return nil
}

func resourceForemanCommonParameterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_common_parameter.go#Delete")

	client := meta.(*api.Client)
	p := buildForemanCommonParameter(d)

	log.Debugf("ForemanCommonParameter: [%+v]", p)

	return diag.FromErr(api.CheckDeleted(d, client.DeleteCommonParameter(ctx, p, p.Id)))
}

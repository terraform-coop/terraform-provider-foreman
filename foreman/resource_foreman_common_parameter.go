package foreman

import (
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceForemanCommonParameter() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanCommonParameterCreate,
		Read:   resourceForemanCommonParameterRead,
		Update: resourceForemanCommonParameterUpdate,
		Delete: resourceForemanCommonParameterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of common_parameter. Global parameters are available for all resources",
					autodoc.MetaSummary,
				),
			},

			// -- Actual Content --
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"value": &schema.Schema{
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

	common_parameter := api.ForemanCommonParameter{}

	obj := buildForemanObject(d)
	common_parameter.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		common_parameter.Name = attr.(string)
	}
	if attr, ok = d.GetOk("value"); ok {
		common_parameter.Value = attr.(string)
	}
	return &common_parameter
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

func resourceForemanCommonParameterCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_common_parameter.go#Create")

	client := meta.(*api.Client)
	p := buildForemanCommonParameter(d)

	log.Debugf("ForemanCommonParameter: [%+v]", d)

	createdParam, createErr := client.CreateCommonParameter(p)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanCommonParameter: [%+v]", createdParam)

	setResourceDataFromForemanCommonParameter(d, createdParam)

	return nil
}

func resourceForemanCommonParameterRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_common_parameter.go#Read")

	client := meta.(*api.Client)
	common_parameter := buildForemanCommonParameter(d)

	log.Debugf("ForemanCommonParameter: [%+v]", common_parameter)

	readCommonParameter, readErr := client.ReadCommonParameter(common_parameter, common_parameter.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanCommonParameter: [%+v]", readCommonParameter)

	setResourceDataFromForemanCommonParameter(d, readCommonParameter)

	return nil
}

func resourceForemanCommonParameterUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_common_parameter.go#Update")

	client := meta.(*api.Client)
	p := buildForemanCommonParameter(d)

	log.Debugf("ForemanCommonParameter: [%+v]", p)

	updatedParam, updateErr := client.UpdateCommonParameter(p, p.Id)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("Updated ForemanCommonParameter: [%+v]", updatedParam)

	setResourceDataFromForemanCommonParameter(d, updatedParam)

	return nil
}

func resourceForemanCommonParameterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_common_parameter.go#Delete")

	client := meta.(*api.Client)
	p := buildForemanCommonParameter(d)

	log.Debugf("ForemanCommonParameter: [%+v]", p)

	return client.DeleteCommonParameter(p, p.Id)
}

package foreman

import (
	"fmt"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceForemanEnvironment() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanEnvironmentCreate,
		Read:   resourceForemanEnvironmentRead,
		Update: resourceForemanEnvironmentUpdate,
		Delete: resourceForemanEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s A puppet environment, branch.",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Name of the environment. Usually maps to the name of "+
						"a puppet branch. "+
						"%s \"production\"",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanEnvironment constructs a ForemanEnvironment reference from a
// resource data reference.  The struct's  members are populated from the data
// populated in the resource data.  Missing members will be left to the zero
// value for that member's type.
func buildForemanEnvironment(d *schema.ResourceData) *api.ForemanEnvironment {
	log.Tracef("resource_foreman_environment.go#buildForemanEnvironment")

	environment := api.ForemanEnvironment{}

	obj := buildForemanObject(d)
	environment.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		environment.Name = attr.(string)
	}

	return &environment
}

// setResourceDataFromForemanEnvironment sets a ResourceData's attributes from
// the attributes of the supplied ForemanEnvironment reference
func setResourceDataFromForemanEnvironment(d *schema.ResourceData, fe *api.ForemanEnvironment) {
	log.Tracef("resource_foreman_environment.go#setResourceDataFromForemanEnvironment")

	d.SetId(strconv.Itoa(fe.Id))
	d.Set("name", fe.Name)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_environment.go#Create")

	client := meta.(*api.Client)
	e := buildForemanEnvironment(d)

	log.Debugf("ForemanEnvironment: [%+v]", e)

	createdEnv, createErr := client.CreateEnvironment(e)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanEnvironment: [%+v]", createdEnv)

	setResourceDataFromForemanEnvironment(d, createdEnv)

	return nil
}

func resourceForemanEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_environment.go#Read")

	client := meta.(*api.Client)
	e := buildForemanEnvironment(d)

	log.Debugf("ForemanEnvironment: [%+v]", e)

	readEnvironment, readErr := client.ReadEnvironment(e.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanEnvironment: [%+v]", readEnvironment)

	setResourceDataFromForemanEnvironment(d, readEnvironment)

	return nil
}

func resourceForemanEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_environment.go#Update")

	client := meta.(*api.Client)
	e := buildForemanEnvironment(d)

	log.Debugf("ForemanEnvironment: [%+v]", e)

	updatedEnv, updateErr := client.UpdateEnvironment(e)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("Updated ForemanEnvironment: [%+v]", updatedEnv)

	setResourceDataFromForemanEnvironment(d, updatedEnv)

	return nil
}

func resourceForemanEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_environment.go#Delete")

	client := meta.(*api.Client)
	e := buildForemanEnvironment(d)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return client.DeleteEnvironment(e.Id)
}

package foreman

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/conv"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

const (
	// Regex validation on the name for architectures.  Architectures
	// can only contain alphanumeric characters, underscore, hyphen,
	// and period.  Any other characters are not allowed.
	architectureNameRegex = `^[A-Za-z0-9-_.]+$`
)

func resourceForemanArchitecture() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanArchitectureCreate,
		Read:   resourceForemanArchitectureRead,
		Update: resourceForemanArchitectureUpdate,
		Delete: resourceForemanArchitectureDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of an instruction set architecture (ISA).",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(architectureNameRegex),
					"Name contains invalid characters. Name can only contain "+
						"alphanumeric characters (A-Z, a-z, 0-9), an undescore (_), "+
						"a hyphen (-), and period (.).",
				),
				Description: fmt.Sprintf(
					"The name of the architecture. Valid characters: %s. "+
						"%s \"i386\"",
					architectureNameRegex,
					autodoc.MetaExample,
				),
			},

			// -- Foreign Key Relationships --

			"operatingsystem_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the operating systems associated with this " +
					"architecture",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanArchitecture constructs a ForemanArchitecture reference from a
// resource data reference.  The struct's  members are populated from the data
// populated in the resource data.  Missing members will be left to the zero
// value for that member's type.
func buildForemanArchitecture(d *schema.ResourceData) *api.ForemanArchitecture {
	log.Tracef("resource_foreman_architecture.go#buildForemanArchitecture")

	arch := api.ForemanArchitecture{}

	obj := buildForemanObject(d)
	arch.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("operatingsystem_ids"); ok {
		attrSet := attr.(*schema.Set)
		arch.OperatingSystemIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	return &arch
}

// setResourceDataFromForemanArchitecture sets a ResourceData's attributes from
// the attributes of the supplied ForemanArchitecture reference
func setResourceDataFromForemanArchitecture(d *schema.ResourceData, fa *api.ForemanArchitecture) {
	log.Tracef("resource_foreman_architecture.go#setResourceDataFromForemanArchitecture")

	d.SetId(strconv.Itoa(fa.Id))
	d.Set("name", fa.Name)
	d.Set("operatingsystem_ids", fa.OperatingSystemIds)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanArchitectureCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_architecture.go#Create")

	client := meta.(*api.Client)
	a := buildForemanArchitecture(d)

	log.Debugf("ForemanArchitecture: [%+v]", a)

	createdArch, createErr := client.CreateArchitecture(a)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanArchitecture: [%+v]", createdArch)

	setResourceDataFromForemanArchitecture(d, createdArch)

	return nil
}

func resourceForemanArchitectureRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_architecture.go#Read")

	client := meta.(*api.Client)
	a := buildForemanArchitecture(d)

	log.Debugf("ForemanArchitecture: [%+v]", a)

	readArch, readErr := client.ReadArchitecture(a.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanArchitecture: [%+v]", readArch)

	setResourceDataFromForemanArchitecture(d, readArch)

	return nil
}

func resourceForemanArchitectureUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_architecture.go#Update")

	client := meta.(*api.Client)
	a := buildForemanArchitecture(d)

	log.Debugf("ForemanArchitecture: [%+v]", a)

	updatedArch, updateErr := client.UpdateArchitecture(a)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("Updated ForemanArchitecture: [%+v]", updatedArch)

	setResourceDataFromForemanArchitecture(d, updatedArch)

	return nil
}

func resourceForemanArchitectureDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_architecture.go#Delete")

	client := meta.(*api.Client)
	a := buildForemanArchitecture(d)

	log.Debugf("ForemanArchitecture: [%+v]", a)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return client.DeleteArchitecture(a.Id)
}

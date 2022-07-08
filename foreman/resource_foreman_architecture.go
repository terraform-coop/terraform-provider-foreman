package foreman

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/conv"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	// Regex validation on the name for architectures.  Architectures
	// can only contain alphanumeric characters, underscore, hyphen,
	// and period.  Any other characters are not allowed.
	architectureNameRegex = `^[A-Za-z0-9-_.]+$`
)

func resourceForemanArchitecture() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanArchitectureCreate,
		ReadContext:   resourceForemanArchitectureRead,
		UpdateContext: resourceForemanArchitectureUpdate,
		DeleteContext: resourceForemanArchitectureDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceForemanArchitectureCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_architecture.go#Create")

	client := meta.(*api.Client)
	a := buildForemanArchitecture(d)

	log.Debugf("ForemanArchitecture: [%+v]", a)

	createdArch, createErr := client.CreateArchitecture(ctx, a)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanArchitecture: [%+v]", createdArch)

	setResourceDataFromForemanArchitecture(d, createdArch)

	return nil
}

func resourceForemanArchitectureRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_architecture.go#Read")

	client := meta.(*api.Client)
	a := buildForemanArchitecture(d)

	log.Debugf("ForemanArchitecture: [%+v]", a)

	readArch, readErr := client.ReadArchitecture(ctx, a.Id)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	log.Debugf("Read ForemanArchitecture: [%+v]", readArch)

	setResourceDataFromForemanArchitecture(d, readArch)

	return nil
}

func resourceForemanArchitectureUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_architecture.go#Update")

	client := meta.(*api.Client)
	a := buildForemanArchitecture(d)

	log.Debugf("ForemanArchitecture: [%+v]", a)

	updatedArch, updateErr := client.UpdateArchitecture(ctx, a)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanArchitecture: [%+v]", updatedArch)

	setResourceDataFromForemanArchitecture(d, updatedArch)

	return nil
}

func resourceForemanArchitectureDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_architecture.go#Delete")

	client := meta.(*api.Client)
	a := buildForemanArchitecture(d)

	log.Debugf("ForemanArchitecture: [%+v]", a)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return diag.FromErr(client.DeleteArchitecture(ctx, a.Id))
}

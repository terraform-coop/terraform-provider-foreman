package foreman

import (
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceForemanImage() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanImageCreate,
		Read:   resourceForemanImageRead,
		Update: resourceForemanImageUpdate,
		Delete: resourceForemanImageDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of image.",
					autodoc.MetaSummary,
				),
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
			"uuid": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
			"compute_resource_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "",
			},
			"operating_system_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "",
			},
			"architecture_id": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanImage constructs a ForemanImage reference from a resource data
// reference.  The struct's  members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanImage(d *schema.ResourceData) *api.ForemanImage {
	log.Tracef("resource_foreman_image.go#buildForemanImage")

	image := api.ForemanImage{}

	obj := buildForemanObject(d)
	image.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		image.Name = attr.(string)
	}
	if attr, ok = d.GetOk("uuid"); ok {
		image.UUID = attr.(string)
	}
	if attr, ok = d.GetOk("username"); ok {
		image.Username = attr.(string)
	}
	if attr, ok = d.GetOk("operating_system_id"); ok {
		image.OperatingSystemID = attr.(int)
	}
	if attr, ok = d.GetOk("architecture_id"); ok {
		image.ArchitectureID = attr.(int)
	}
	if attr, ok = d.GetOk("compute_resource_id"); ok {
		image.ComputeResourceID = attr.(int)
	}

	return &image
}

// setResourceDataFromForemanImage sets a ResourceData's attributes from the
// attributes of the supplied ForemanImage reference
func setResourceDataFromForemanImage(d *schema.ResourceData, fd *api.ForemanImage) {
	log.Tracef("resource_foreman_image.go#setResourceDataFromForemanImage")

	d.SetId(strconv.Itoa(fd.Id))
	d.Set("name", fd.Name)
	d.Set("username", fd.Username)
	d.Set("uuid", fd.UUID)
	d.Set("operating_system_id", fd.OperatingSystemID)
	d.Set("architecture_id", fd.ArchitectureID)
	d.Set("compute_resource_id", fd.ComputeResourceID)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanImageCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_image.go#Create")
	return nil
}

func resourceForemanImageRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_image.go#Read")

	client := meta.(*api.Client)
	image := buildForemanImage(d)

	log.Debugf("ForemanImage: [%+v]", image)

	readImage, readErr := client.ReadImage(image)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanImage: [%+v]", readImage)

	setResourceDataFromForemanImage(d, readImage)

	return nil
}

func resourceForemanImageUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_image.go#Update")
	return nil
}

func resourceForemanImageDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_image.go#Delete")

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return nil
}

package foreman

import (
	"fmt"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceForemanModel() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanModelCreate,
		Read:   resourceForemanModelRead,
		Update: resourceForemanModelUpdate,
		Delete: resourceForemanModelDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Vendor-specific hardware model.",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The name of the hardware model. "+
						"%s \"PowerEdge FC630\"",
					autodoc.MetaExample,
				),
			},

			"info": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional information about this hardware model.",
			},

			"vendor_class": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name or class of the hardware vendor.",
			},

			"hardware_model": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the specific hardware model.",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanModel constructs a ForemanModel struct from a resource data
// reference.  The struct's members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanModel(d *schema.ResourceData) *api.ForemanModel {
	log.Tracef("resource_foreman_model.go#buildForemanModel")

	model := api.ForemanModel{}

	obj := buildForemanObject(d)
	model.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("info"); ok {
		model.Info = attr.(string)
	}

	if attr, ok = d.GetOk("vendor_class"); ok {
		model.VendorClass = attr.(string)
	}

	if attr, ok = d.GetOk("hardware_model"); ok {
		model.HardwareModel = attr.(string)
	}

	return &model
}

// setResourceDataFromForemanModel sets a ResourceData's attributes from the
// attributes of the supplied ForemanModel struct
func setResourceDataFromForemanModel(d *schema.ResourceData, fm *api.ForemanModel) {
	log.Tracef("resource_foreman_model.go#setResourceDataFromForemanModel")

	d.SetId(strconv.Itoa(fm.Id))
	d.Set("name", fm.Name)
	d.Set("info", fm.Info)
	d.Set("vendor_class", fm.VendorClass)
	d.Set("hardware_model", fm.HardwareModel)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanModelCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_model.go#Create")

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	log.Debugf("ForemanModel: [%+v]", m)

	createdModel, createErr := client.CreateModel(m)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanModel: [%+v]", createdModel)

	setResourceDataFromForemanModel(d, createdModel)

	return nil
}

func resourceForemanModelRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_model.go#Read")

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	log.Debugf("ForemanModel: [%+v]", m)

	readModel, readErr := client.ReadModel(m.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanModel: [%+v]", readModel)

	setResourceDataFromForemanModel(d, readModel)

	return nil
}

func resourceForemanModelUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_model.go#Update")

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	log.Debugf("ForemanModel: [%+v]", m)

	updatedModel, updateErr := client.UpdateModel(m)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("Updated ForemanModel: [%+v]", updatedModel)

	setResourceDataFromForemanModel(d, updatedModel)

	return nil
}

func resourceForemanModelDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_model.go#Delete")

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	log.Debugf("ForemanModel: [%+v]", m)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return client.DeleteModel(m.Id)
}

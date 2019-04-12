package foreman

import (
	"fmt"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceForemanDefaultTemplate() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanDefaultTemplateCreate,
		Read:   resourceForemanDefaultTemplateRead,
		Update: resourceForemanDefaultTemplateUpdate,
		Delete: resourceForemanDefaultTemplateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of default Template. Default Templates serve as an "+
						"identification string that defines autonomy, authority, or control "+
						"for a portion of a network.",
					autodoc.MetaSummary,
				),
			},

			"operatingsystem_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "ID of the operating system to assign this Default Template to",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"provisioningtemplate_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Id of the Provisioning Template",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"templatekind_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Template Kind Id to define the Default Template",
				ValidateFunc: validation.IntAtLeast(1),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanDefaultTemplate constructs a ForemanDefaultTemplate reference from a resource data
// reference.  The struct's  members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanDefaultTemplate(d *schema.ResourceData) *api.ForemanDefaultTemplate {
	log.Tracef("resource_foreman_defaultTemplate.go#buildForemanDefaultTemplate")

	defaultTemplate := api.ForemanDefaultTemplate{}

	obj := buildForemanObject(d)
	defaultTemplate.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("provisioningtemplate_id"); ok {
		defaultTemplate.ProvisioningTemplateId = attr.(int)
	}
	if attr, ok = d.GetOk("templatekind_id"); ok {
		defaultTemplate.TemplateKindId = attr.(int)
	}
	if attr, ok = d.GetOk("operatingsystem_id"); ok {
		defaultTemplate.OperatingSystemId = attr.(int)
	}
	return &defaultTemplate
}

// setResourceDataFromForemanDefaultTemplate sets a ResourceData's attributes from the
// attributes of the supplied ForemanDefaultTemplate reference
func setResourceDataFromForemanDefaultTemplate(d *schema.ResourceData, fd *api.ForemanDefaultTemplate) {
	log.Tracef("resource_foreman_defaultTemplate.go#setResourceDataFromForemanDefaultTemplate")

	d.SetId(strconv.Itoa(fd.Id))
	d.Set("provisioningtemplate_id", fd.ProvisioningTemplateId)
	d.Set("templatekind_id", fd.TemplateKindId)
	d.Set("operatingsystem_id", fd.OperatingSystemId)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanDefaultTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_defaultTemplate.go#Create")

	client := meta.(*api.Client)
	p := buildForemanDefaultTemplate(d)

	log.Debugf("ForemanDefaultTemplate: [%+v]", d)

	createdParam, createErr := client.CreateDefaultTemplate(p)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanDefaultTemplate: [%+v]", createdParam)

	setResourceDataFromForemanDefaultTemplate(d, createdParam)

	return nil
}

func resourceForemanDefaultTemplateRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_defaultTemplate.go#Read")

	client := meta.(*api.Client)
	defaultTemplate := buildForemanDefaultTemplate(d)

	log.Debugf("ForemanDefaultTemplate: [%+v]", defaultTemplate)

	readDefaultTemplate, readErr := client.ReadDefaultTemplate(defaultTemplate, defaultTemplate.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanDefaultTemplate: [%+v]", readDefaultTemplate)

	setResourceDataFromForemanDefaultTemplate(d, readDefaultTemplate)

	return nil
}

func resourceForemanDefaultTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_defaultTemplate.go#Update")

	client := meta.(*api.Client)
	p := buildForemanDefaultTemplate(d)

	log.Debugf("ForemanDefaultTemplate: [%+v]", p)

	updatedParam, updateErr := client.UpdateDefaultTemplate(p, p.Id)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("Updated ForemanDefaultTemplate: [%+v]", updatedParam)

	setResourceDataFromForemanDefaultTemplate(d, updatedParam)

	return nil
}

func resourceForemanDefaultTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_defaultTemplate.go#Delete")

	client := meta.(*api.Client)
	p := buildForemanDefaultTemplate(d)

	log.Debugf("ForemanDefaultTemplate: [%+v]", p)

	return client.DeleteDefaultTemplate(p, p.Id)
}

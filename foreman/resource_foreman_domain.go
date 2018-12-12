package foreman

import (
	"fmt"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceForemanDomain() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanDomainCreate,
		Read:   resourceForemanDomainRead,
		Update: resourceForemanDomainUpdate,
		Delete: resourceForemanDomainDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of domain. Domains serve as an "+
						"identification string that defines autonomy, authority, or control "+
						"for a portion of a network.",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The name of the domain - the full DNS domain name. "+
						"%s \"dev.dc1.company.com\"",
					autodoc.MetaExample,
				),
			},

			"fullname": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the domain",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanDomain constructs a ForemanDomain reference from a resource data
// reference.  The struct's  members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanDomain(d *schema.ResourceData) *api.ForemanDomain {
	log.Tracef("resource_foreman_domain.go#buildForemanDomain")

	domain := api.ForemanDomain{}

	obj := buildForemanObject(d)
	domain.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("fullname"); ok {
		domain.Fullname = attr.(string)
	}

	return &domain
}

// setResourceDataFromForemanDomain sets a ResourceData's attributes from the
// attributes of the supplied ForemanDomain reference
func setResourceDataFromForemanDomain(d *schema.ResourceData, fd *api.ForemanDomain) {
	log.Tracef("resource_foreman_domain.go#setResourceDataFromForemanDomain")

	d.SetId(strconv.Itoa(fd.Id))
	d.Set("name", fd.Name)
	d.Set("fullname", fd.Fullname)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanDomainCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_domain.go#Create")
	return nil
}

func resourceForemanDomainRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_domain.go#Read")

	client := meta.(*api.Client)
	domain := buildForemanDomain(d)

	log.Debugf("ForemanDomain: [%+v]", domain)

	readDomain, readErr := client.ReadDomain(domain.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanDomain: [%+v]", readDomain)

	setResourceDataFromForemanDomain(d, readDomain)

	return nil
}

func resourceForemanDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_domain.go#Update")
	return nil
}

func resourceForemanDomainDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_domain.go#Delete")

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return nil
}

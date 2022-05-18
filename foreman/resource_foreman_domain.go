package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceForemanDomain() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanDomainCreate,
		ReadContext:   resourceForemanDomainRead,
		UpdateContext: resourceForemanDomainUpdate,
		DeleteContext: resourceForemanDomainDelete,

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

			"parameters": &schema.Schema{
				Type:     schema.TypeMap,
				ForceNew: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of parameters that will be saved as domain parameters " +
					"in the domain config.",
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

	if attr, ok = d.GetOk("parameters"); ok {
		domain.DomainParameters = api.ToKV(attr.(map[string]string))
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
	d.Set("parameters", api.FromKV(fd.DomainParameters))
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_domain.go#Create")

	client := meta.(*api.Client)
	p := buildForemanDomain(d)

	log.Debugf("ForemanDomain: [%+v]", d)

	createdDomain, createErr := client.CreateDomain(ctx, p)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanDomain: [%+v]", createdDomain)

	setResourceDataFromForemanDomain(d, createdDomain)

	return nil
}

func resourceForemanDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_domain.go#Read")

	client := meta.(*api.Client)
	domain := buildForemanDomain(d)

	log.Debugf("ForemanDomain: [%+v]", domain)

	readDomain, readErr := client.ReadDomain(ctx, domain.Id)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	log.Debugf("Read ForemanDomain: [%+v]", readDomain)

	setResourceDataFromForemanDomain(d, readDomain)

	return nil
}

func resourceForemanDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_domain.go#Update")
	client := meta.(*api.Client)
	do := buildForemanDomain(d)

	log.Debugf("ForemanDomain: [%+v]", do)

	updatedDomain, updateErr := client.UpdateDomain(ctx, do, do.Id)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanDomain: [%+v]", updatedDomain)

	setResourceDataFromForemanDomain(d, updatedDomain)

	return nil
}

func resourceForemanDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_domain.go#Delete")

	client := meta.(*api.Client)
	do := buildForemanDomain(d)

	log.Debugf("ForemanDomain: [%+v]", do)

	return diag.FromErr(client.DeleteDomain(ctx, do.Id))
}

package foreman

import (
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceForemanHTTPProxy() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanHTTPProxyCreate,
		Read:   resourceForemanHTTPProxyRead,
		Update: resourceForemanHTTPProxyUpdate,
		Delete: resourceForemanHTTPProxyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Defining HTTP Proxies that exist on your network allows "+
						"you to perform various actions through those proxies.",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The name of the http proxy. "+
						"%s \"proxy.company.com\"",
					autodoc.MetaExample,
				),
			},

			"url": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Description: fmt.Sprintf(
					"Uniform resource locator of the proxy. "+
						"%s \"https://proxy.company.com:8443\"",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanHTTPProxy constructs a ForemanHTTPProxy struct from a resource
// data reference.  The struct's members are populated from the data populated
// in the resource data.  Missing members will be left to the zero value for
// that member's type.
func buildForemanHTTPProxy(d *schema.ResourceData) *api.ForemanHTTPProxy {
	log.Tracef("resource_foreman_httpproxy.go#buildForemanHTTPProxy")

	proxy := api.ForemanHTTPProxy{}

	obj := buildForemanObject(d)
	proxy.ForemanObject = *obj

	proxy.URL = d.Get("url").(string)

	return &proxy
}

// setResourceDataFromForemanHTTPProxy sets a ResourceData's attributes from
// the attributes of the supplied ForemanHTTPProxy struct
func setResourceDataFromForemanHTTPProxy(d *schema.ResourceData, fp *api.ForemanHTTPProxy) {
	log.Tracef("resource_foreman_httpproxy.go#setResourceDataFromForemanHTTPProxy")

	d.SetId(strconv.Itoa(fp.Id))
	d.Set("name", fp.Name)
	d.Set("url", fp.URL)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanHTTPProxyCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_httpproxy.go#Create")

	client := meta.(*api.Client)
	p := buildForemanHTTPProxy(d)

	log.Debugf("ForemanHTTPProxy: [%+v]", p)

	createdHTTPProxy, createErr := client.CreateHTTPProxy(p)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanHTTPProxy: [%+v]", createdHTTPProxy)

	setResourceDataFromForemanHTTPProxy(d, createdHTTPProxy)

	return nil
}

func resourceForemanHTTPProxyRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_httpproxy.go#Read")

	client := meta.(*api.Client)
	p := buildForemanHTTPProxy(d)

	log.Debugf("ForemanHTTPProxy: [%+v]", p)

	readHTTPProxy, readErr := client.ReadHTTPProxy(p.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanHTTPProxy: [%+v]", readHTTPProxy)

	setResourceDataFromForemanHTTPProxy(d, readHTTPProxy)

	return nil
}

func resourceForemanHTTPProxyUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_httpproxy.go#Update")

	client := meta.(*api.Client)
	p := buildForemanHTTPProxy(d)

	log.Debugf("ForemanHTTPProxy: [%+v]", p)

	updatedHTTPProxy, updateErr := client.UpdateHTTPProxy(p)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("ForemanHTTPProxy: [%+v]", updatedHTTPProxy)

	setResourceDataFromForemanHTTPProxy(d, updatedHTTPProxy)

	return nil
}

func resourceForemanHTTPProxyDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_httpproxy.go#Delete")

	client := meta.(*api.Client)
	p := buildForemanHTTPProxy(d)

	log.Debugf("ForemanHTTPProxy: [%+v]", p)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return client.DeleteHTTPProxy(p.Id)
}

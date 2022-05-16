package foreman

import (
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanSmartProxy() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanSmartProxyCreate,
		Read:   resourceForemanSmartProxyRead,
		Update: resourceForemanSmartProxyUpdate,
		Delete: resourceForemanSmartProxyDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Smart proxies provide an API for a higher-level orchestration "+
						"tool. Foreman supports the following smart proxies: DHCP "+
						"(ISC DHCP & MS DHCP servers), DNS (bind & MS DNS servers), "+
						"Puppet >= 0.24.x, Puppet CA, Realm (FreeIPA), Templates, TFTP.",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The name of the smart proxy. "+
						"%s \"dns.dc1.company.com\"",
					autodoc.MetaExample,
				),
			},

			"url": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Description: fmt.Sprintf(
					"Uniform resource locator of the proxy. "+
						"%s \"https://dns.dc1.company.com:8443\"",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanSmartProxy constructs a ForemanSmartProxy struct from a resource
// data reference.  The struct's members are populated from the data populated
// in the resource data.  Missing members will be left to the zero value for
// that member's type.
func buildForemanSmartProxy(d *schema.ResourceData) *api.ForemanSmartProxy {
	log.Tracef("resource_foreman_smartproxy.go#buildForemanSmartProxy")

	proxy := api.ForemanSmartProxy{}

	obj := buildForemanObject(d)
	proxy.ForemanObject = *obj

	proxy.URL = d.Get("url").(string)

	return &proxy
}

// setResourceDataFromForemanSmartProxy sets a ResourceData's attributes from
// the attributes of the supplied ForemanSmartProxy struct
func setResourceDataFromForemanSmartProxy(d *schema.ResourceData, fp *api.ForemanSmartProxy) {
	log.Tracef("resource_foreman_smartproxy.go#setResourceDataFromForemanSmartProxy")

	d.SetId(strconv.Itoa(fp.Id))
	d.Set("name", fp.Name)
	d.Set("url", fp.URL)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanSmartProxyCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_smartproxy.go#Create")

	client := meta.(*api.Client)
	s := buildForemanSmartProxy(d)

	log.Debugf("ForemanSmartProxy: [%+v]", s)

	createdSmartProxy, createErr := client.CreateSmartProxy(s)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanSmartProxy: [%+v]", createdSmartProxy)

	setResourceDataFromForemanSmartProxy(d, createdSmartProxy)

	return nil
}

func resourceForemanSmartProxyRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_smartproxy.go#Read")

	client := meta.(*api.Client)
	s := buildForemanSmartProxy(d)

	log.Debugf("ForemanSmartProxy: [%+v]", s)

	readSmartProxy, readErr := client.ReadSmartProxy(s.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanSmartProxy: [%+v]", readSmartProxy)

	setResourceDataFromForemanSmartProxy(d, readSmartProxy)

	return nil
}

func resourceForemanSmartProxyUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_smartproxy.go#Update")

	client := meta.(*api.Client)
	s := buildForemanSmartProxy(d)

	log.Debugf("ForemanSmartProxy: [%+v]", s)

	updatedSmartProxy, updateErr := client.UpdateSmartProxy(s)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("ForemanSmartProxy: [%+v]", updatedSmartProxy)

	setResourceDataFromForemanSmartProxy(d, updatedSmartProxy)

	return nil
}

func resourceForemanSmartProxyDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_smartproxy.go#Delete")

	client := meta.(*api.Client)
	s := buildForemanSmartProxy(d)

	log.Debugf("ForemanSmartProxy: [%+v]", s)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return client.DeleteSmartProxy(s.Id)
}

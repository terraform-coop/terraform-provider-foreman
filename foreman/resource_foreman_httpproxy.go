package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanHTTPProxy() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanHTTPProxyCreate,
		ReadContext:   resourceForemanHTTPProxyRead,
		UpdateContext: resourceForemanHTTPProxyUpdate,
		DeleteContext: resourceForemanHTTPProxyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Defining HTTP Proxies that exist on your network allows "+
						"you to perform various actions through those proxies.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The name of the http proxy. "+
						"%s \"proxy.company.com\"",
					autodoc.MetaExample,
				),
			},

			"url": {
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
	utils.TraceFunctionCall()

	proxy := api.ForemanHTTPProxy{}

	obj := buildForemanObject(d)
	proxy.ForemanObject = *obj

	proxy.URL = d.Get("url").(string)

	return &proxy
}

// setResourceDataFromForemanHTTPProxy sets a ResourceData's attributes from
// the attributes of the supplied ForemanHTTPProxy struct
func setResourceDataFromForemanHTTPProxy(d *schema.ResourceData, fp *api.ForemanHTTPProxy) {
	utils.TraceFunctionCall()

	d.SetId(strconv.Itoa(fp.Id))
	d.Set("name", fp.Name)
	d.Set("url", fp.URL)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanHTTPProxyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	p := buildForemanHTTPProxy(d)

	log.Debugf("ForemanHTTPProxy: [%+v]", p)

	createdHTTPProxy, createErr := client.CreateHTTPProxy(ctx, p)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanHTTPProxy: [%+v]", createdHTTPProxy)

	setResourceDataFromForemanHTTPProxy(d, createdHTTPProxy)

	return nil
}

func resourceForemanHTTPProxyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	p := buildForemanHTTPProxy(d)

	log.Debugf("ForemanHTTPProxy: [%+v]", p)

	readHTTPProxy, readErr := client.ReadHTTPProxy(ctx, p.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanHTTPProxy: [%+v]", readHTTPProxy)

	setResourceDataFromForemanHTTPProxy(d, readHTTPProxy)

	return nil
}

func resourceForemanHTTPProxyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	p := buildForemanHTTPProxy(d)

	log.Debugf("ForemanHTTPProxy: [%+v]", p)

	updatedHTTPProxy, updateErr := client.UpdateHTTPProxy(ctx, p)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("ForemanHTTPProxy: [%+v]", updatedHTTPProxy)

	setResourceDataFromForemanHTTPProxy(d, updatedHTTPProxy)

	return nil
}

func resourceForemanHTTPProxyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	p := buildForemanHTTPProxy(d)

	log.Debugf("ForemanHTTPProxy: [%+v]", p)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return diag.FromErr(api.CheckDeleted(d, client.DeleteHTTPProxy(ctx, p.Id)))
}

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
)

func resourceForemanKatelloProduct() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanKatelloProductCreate,
		ReadContext:   resourceForemanKatelloProductRead,
		UpdateContext: resourceForemanKatelloProductUpdate,
		DeleteContext: resourceForemanKatelloProductDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Poducts are mostly operating systems to which repositories are assigned.",
					autodoc.MetaSummary,
				),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Product name."+
						"%s \"Debian 10\"",
					autodoc.MetaExample,
				),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Product description."+
						"%s \"A product description\"",
					autodoc.MetaExample,
				),
			},
			"gpg_key_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Identifier of the GPG key."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"ssl_ca_cert_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Idenifier of the SSL CA Cert."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"ssl_client_cert_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Identifier of the SSL Client Cert."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"ssl_client_key_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Identifier of the SSL Client Key."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"sync_plan_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Plan numeric identifier."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"label": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true, // Created from name if not passed in
				ForceNew: true,
				Description: fmt.Sprintf(
					"Label for the product. Cannot be changed after creation. By default set to the name, "+
						"with underscores as spaces replacement. %s",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanKatelloProduct constructs a ForemanKatelloProduct struct from a resource
// data reference. The struct's members are populated from the data populated
// in the resource data. Missing members will be left to the zero value for
// that member's type.
func buildForemanKatelloProduct(d *schema.ResourceData) *api.ForemanKatelloProduct {
	utils.TraceFunctionCall()

	Product := api.ForemanKatelloProduct{}

	obj := buildForemanObject(d)
	Product.ForemanObject = *obj

	Product.Description = d.Get("description").(string)
	Product.GpgKeyId = d.Get("gpg_key_id").(int)
	Product.SslCaCertId = d.Get("ssl_ca_cert_id").(int)
	Product.SslClientCertId = d.Get("ssl_client_cert_id").(int)
	Product.SslClientKeyId = d.Get("ssl_client_key_id").(int)
	Product.SyncPlanId = d.Get("sync_plan_id").(int)
	Product.Label = d.Get("label").(string)

	return &Product
}

// setResourceDataFromForemanKatelloProduct sets a ResourceData's attributes from
// the attributes of the supplied ForemanKatelloProduct struct
func setResourceDataFromForemanKatelloProduct(d *schema.ResourceData, Product *api.ForemanKatelloProduct) {
	utils.TraceFunctionCall()

	d.SetId(strconv.Itoa(Product.Id))
	d.Set("name", Product.Name)
	d.Set("description", Product.Description)
	d.Set("gpg_key_id", Product.GpgKeyId)
	d.Set("ssl_ca_cert_id", Product.SslCaCertId)
	d.Set("ssl_client_cert_id", Product.SslClientCertId)
	d.Set("ssl_client_key_id", Product.SslClientKeyId)
	d.Set("sync_plan_id", Product.SyncPlanId)
	d.Set("label", Product.Label)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanKatelloProductCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	product := buildForemanKatelloProduct(d)

	log.Debugf("ForemanKatelloProduct: [%+v]", product)

	createdKatelloProduct, createErr := client.CreateKatelloProduct(ctx, product)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanKatelloProduct: [%+v]", createdKatelloProduct)

	setResourceDataFromForemanKatelloProduct(d, createdKatelloProduct)

	return nil
}

func resourceForemanKatelloProductRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	product := buildForemanKatelloProduct(d)

	log.Debugf("ForemanKatelloProduct: [%+v]", product)

	readKatelloProduct, readErr := client.ReadKatelloProduct(ctx, product.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanKatelloProduct: [%+v]", readKatelloProduct)

	setResourceDataFromForemanKatelloProduct(d, readKatelloProduct)

	return nil
}

func resourceForemanKatelloProductUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	product := buildForemanKatelloProduct(d)

	log.Debugf("ForemanKatelloProduct: [%+v]", product)

	updatedKatelloProduct, updateErr := client.UpdateKatelloProduct(ctx, product)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("ForemanKatelloProduct: [%+v]", updatedKatelloProduct)

	setResourceDataFromForemanKatelloProduct(d, updatedKatelloProduct)

	return nil
}

func resourceForemanKatelloProductDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	product := buildForemanKatelloProduct(d)

	log.Debugf("ForemanKatelloProduct: [%+v]", product)

	return diag.FromErr(api.CheckDeleted(d, client.DeleteKatelloProduct(ctx, product.Id)))
}

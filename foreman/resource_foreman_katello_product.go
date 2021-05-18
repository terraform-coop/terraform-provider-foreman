package foreman

import (
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceForemanKatelloProduct() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanKatelloProductCreate,
		Read:   resourceForemanKatelloProductRead,
		Update: resourceForemanKatelloProductUpdate,
		Delete: resourceForemanKatelloProductDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Poducts are mostly operating systems to which repositories are assigned.",
					autodoc.MetaSummary,
				),
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Product name."+
						"%s \"Debian 10\"",
					autodoc.MetaExample,
				),
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"Product description."+
						"%s \"A product description\"",
					autodoc.MetaExample,
				),
			},
			"gpg_key_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Identifier of the GPG key."+
						"%s",
					autodoc.MetaExample,
				),
			},
			/*
						"ssl_ca_cert_id": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Description: fmt.Sprintf(
								"Idenifier of the SSL CA Cert."+
									"%s",
								autodoc.MetaExample,
							),
						},
			            "ssl_client_cert_id": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Description: fmt.Sprintf(
								"Identifier of the SSL Client Cert."+
									"%s",
								autodoc.MetaExample,
							),
						},
			            "ssl_client_key_id": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Description: fmt.Sprintf(
								"Identifier of the SSL Client Key."+
									"%s",
								autodoc.MetaExample,
							),
						}, */
			"sync_plan_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Description: fmt.Sprintf(
					"Plan numeric identifier."+
						"%s",
					autodoc.MetaExample,
				),
			},
			"label": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Description: fmt.Sprintf(
					"%s",
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
	log.Tracef("resource_foreman_katello_product.go#buildForemanKatelloProduct")

	Product := api.ForemanKatelloProduct{}

	obj := buildForemanObject(d)
	Product.ForemanObject = *obj

	Product.Description = d.Get("description").(string)
	Product.GpgKeyId = d.Get("gpg_key_id").(int)
	/* 	Product.SslCaCertId = d.Get("ssl_ca_cert_id").(int)
	   	Product.SslClientCertId = d.Get("ssl_client_cert_id").(int)
	       Product.SslClientKeyId = d.Get("ssl_client_key_id").(int) */
	Product.SyncPlanId = d.Get("sync_plan_id").(int)
	Product.Label = d.Get("label").(string)

	return &Product
}

// setResourceDataFromForemanKatelloProduct sets a ResourceData's attributes from
// the attributes of the supplied ForemanKatelloProduct struct
func setResourceDataFromForemanKatelloProduct(d *schema.ResourceData, Product *api.ForemanKatelloProduct) {
	log.Tracef("resource_foreman_katello_product.go#setResourceDataFromForemanKatelloProduct")

	d.SetId(strconv.Itoa(Product.Id))
	d.Set("name", Product.Name)
	d.Set("description", Product.Description)
	d.Set("gpg_key_id", Product.GpgKeyId)
	/* 	d.Set("ssl_ca_cert_id", Product.SslCaCertId)
	   	d.Set("ssl_client_cert_id", Product.SslClientCertId)
	   	d.Set("ssl_client_key_id", Product.SslClientKeyId) */
	d.Set("sync_plan_id", Product.SyncPlanId)
	d.Set("label", Product.Label)

}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanKatelloProductCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_katello_product.go#Create")

	client := meta.(*api.Client)
	Product := buildForemanKatelloProduct(d)

	log.Debugf("ForemanKatelloProduct: [%+v]", Product)

	createdKatelloProduct, createErr := client.CreateKatelloProduct(Product)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanKatelloProduct: [%+v]", createdKatelloProduct)

	setResourceDataFromForemanKatelloProduct(d, createdKatelloProduct)

	return nil
}

func resourceForemanKatelloProductRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_katello_product.go#Read")

	client := meta.(*api.Client)
	Product := buildForemanKatelloProduct(d)

	log.Debugf("ForemanKatelloProduct: [%+v]", Product)

	readKatelloProduct, readErr := client.ReadKatelloProduct(Product.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanKatelloProduct: [%+v]", readKatelloProduct)

	setResourceDataFromForemanKatelloProduct(d, readKatelloProduct)

	return nil
}

func resourceForemanKatelloProductUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_katello_product.go#Update")

	client := meta.(*api.Client)
	Product := buildForemanKatelloProduct(d)

	log.Debugf("ForemanKatelloProduct: [%+v]", Product)

	updatedKatelloProduct, updateErr := client.UpdateKatelloProduct(Product)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("ForemanKatelloProduct: [%+v]", updatedKatelloProduct)

	setResourceDataFromForemanKatelloProduct(d, updatedKatelloProduct)

	return nil
}

func resourceForemanKatelloProductDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_katello_product.go#Delete")

	client := meta.(*api.Client)
	Product := buildForemanKatelloProduct(d)

	log.Debugf("ForemanKatelloProduct: [%+v]", Product)

	return client.DeleteKatelloProduct(Product.Id)
}

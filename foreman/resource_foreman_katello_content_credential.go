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

func resourceForemanKatelloContentCredential() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanKatelloContentCredentialCreate,
		ReadContext:   resourceForemanKatelloContentCredentialRead,
		UpdateContext: resourceForemanKatelloContentCredentialUpdate,
		DeleteContext: resourceForemanKatelloContentCredentialDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Content Credentials are used to store credentials like GPG Keys and Certificates "+
						"for the authentication to Products / Repositories.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Identifier of the content credential."+
						"%s \"RPM-GPG-KEY-centos7\"",
					autodoc.MetaExample,
				),
			},

			"content": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Public key block in DER encoding or certificate content. "+
						"%s \"-----BEGIN PGP PUBLIC KEY BLOCK-----\n"+
						"...\n"+
						"-----END PGP PUBLIC KEY BLOCK-----\"",
					autodoc.MetaExample,
				),
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanKatelloContentCredential constructs a ForemanKatelloContentCredential struct from a resource
// data reference.  The struct's members are populated from the data populated
// in the resource data.  Missing members will be left to the zero value for
// that member's type.
func buildForemanKatelloContentCredential(d *schema.ResourceData) *api.ForemanKatelloContentCredential {
	utils.TraceFunctionCall()

	contentCredential := api.ForemanKatelloContentCredential{}

	obj := buildForemanObject(d)
	contentCredential.ForemanObject = *obj

	contentCredential.Content = d.Get("content").(string)

	return &contentCredential
}

// setResourceDataFromForemanKatelloContentCredential sets a ResourceData's attributes from
// the attributes of the supplied ForemanKatelloContentCredential struct
func setResourceDataFromForemanKatelloContentCredential(d *schema.ResourceData, contentCredential *api.ForemanKatelloContentCredential) {
	utils.TraceFunctionCall()

	d.SetId(strconv.Itoa(contentCredential.Id))
	d.Set("name", contentCredential.Name)
	d.Set("content", contentCredential.Content)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanKatelloContentCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	contentCredential := buildForemanKatelloContentCredential(d)

	log.Debugf("ForemanKatelloContentCredential: [%+v]", contentCredential)

	createdKatelloContentCredential, createErr := client.CreateKatelloContentCredential(ctx, contentCredential)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanKatelloContentCredential: [%+v]", createdKatelloContentCredential)

	setResourceDataFromForemanKatelloContentCredential(d, createdKatelloContentCredential)

	return nil
}

func resourceForemanKatelloContentCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	contentCredential := buildForemanKatelloContentCredential(d)

	log.Debugf("ForemanKatelloContentCredential: [%+v]", contentCredential)

	readKatelloContentCredential, readErr := client.ReadKatelloContentCredential(ctx, contentCredential.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanKatelloContentCredential: [%+v]", readKatelloContentCredential)

	setResourceDataFromForemanKatelloContentCredential(d, readKatelloContentCredential)

	return nil
}

func resourceForemanKatelloContentCredentialUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	contentCredential := buildForemanKatelloContentCredential(d)

	log.Debugf("ForemanKatelloContentCredential: [%+v]", contentCredential)

	updatedKatelloContentCredential, updateErr := client.UpdateKatelloContentCredential(ctx, contentCredential)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("ForemanKatelloContentCredential: [%+v]", updatedKatelloContentCredential)

	setResourceDataFromForemanKatelloContentCredential(d, updatedKatelloContentCredential)

	return nil
}

func resourceForemanKatelloContentCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	contentCredential := buildForemanKatelloContentCredential(d)

	log.Debugf("ForemanKatelloContentCredential: [%+v]", contentCredential)

	return diag.FromErr(api.CheckDeleted(d, client.DeleteKatelloContentCredential(ctx, contentCredential.Id)))
}

package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanWebhook() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanWebhookCreate,
		ReadContext:   resourceForemanWebhookRead,
		UpdateContext: resourceForemanWebhookUpdate,
		DeleteContext: resourceForemanWebhookDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "Webhooks provide integration to 3rd parties via web services with " +
					"configurable payload." +
					autodoc.MetaSummary,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(8, 256),
				Description: fmt.Sprintf(
					"Webhook name "+
						"%s \"compute\"",
					autodoc.MetaExample,
				),
			},

			"target_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Webhook Target URL.",
			},

			"http_method": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(3, 6),
				Description:  "Must be one of POST, GET, PUT, DELETE, PATCH.",
			},

			"http_content_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "HTTP content type.",
			},

			"http_headers": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "HTTP headers to send.",
			},

			"event": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Event that triggers the webhook.",
			},

			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "If the webhook is enabled.",
			},

			"verify_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Verify target's SSL certificate.",
			},

			"ssl_ca_certs": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "X509 Certification Authorities concatenated in PEM format.",
			},

			"proxy_authorization": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Authorize with Foreman client certificate and validate smart-proxy CA from Settings.",
			},

			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "User for authentication.",
			},

			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Password for authentication.",
			},

			"webhook_template_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Password for authentication.",
			},
		},
	}
}

// buildForemanWebhook constructs a ForemanWebhook struct from a resource
// data reference. The struct's members are populated from the data populated
// in the resource data. Missing members will be left to the zero value for
// that member's type.
func buildForemanWebhook(d *schema.ResourceData) *api.ForemanWebhook {
	log.Tracef("resource_foreman_webhook.go#buildForemanWebhook")

	webhook := api.ForemanWebhook{}

	obj := buildForemanObject(d)
	webhook.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		webhook.Name = attr.(string)
	}

	if attr, ok = d.GetOk("target_url"); ok {
		webhook.TargetURL = attr.(string)
	}

	if attr, ok = d.GetOk("http_method"); ok {
		webhook.HTTPMethod = attr.(string)
	}

	if attr, ok = d.GetOk("http_content_type"); ok {
		webhook.HTTPContentType = attr.(string)
	}

	if attr, ok = d.GetOk("http_headers"); ok {
		webhook.HTTPHeaders = attr.(string)
	}

	if attr, ok = d.GetOk("event"); ok {
		webhook.Event = attr.(string)
	}

	if attr, ok = d.GetOk("enabled"); ok {
		webhook.Enabled = attr.(bool)
	}

	if attr, ok = d.GetOk("verify_ssl"); ok {
		webhook.VerifySSL = attr.(bool)
	}

	if attr, ok = d.GetOk("ssl_ca_certs"); ok {
		webhook.SSLCACerts = attr.(string)
	}

	if attr, ok = d.GetOk("proxy_authorization"); ok {
		webhook.ProxyAuthorization = attr.(bool)
	}

	if attr, ok = d.GetOk("user"); ok {
		webhook.User = attr.(string)
	}

	if attr, ok = d.GetOk("password"); ok {
		webhook.Password = attr.(string)
	}

	if attr, ok = d.GetOk("webhook_template_id"); ok {
		webhook.WebhookTemplateID = attr.(int)
	}

	return &webhook
}

// buildForemanWebhookResponse constructs a ForemanWebhookResponse struct from a resource
// data reference. The struct's members are populated from the data
// in the resource data. Missing members will be left to the zero value for
// that member's type.
func buildForemanWebhookResponse(d *schema.ResourceData) *api.ForemanWebhookResponse {
	log.Tracef("resource_foreman_webhook.go#buildForemanWebhookResponse")

	webhookResponse := api.ForemanWebhookResponse{}

	obj := buildForemanObject(d)
	webhookResponse.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		webhookResponse.Name = attr.(string)
	}

	if attr, ok = d.GetOk("target_url"); ok {
		webhookResponse.TargetURL = attr.(string)
	}

	if attr, ok = d.GetOk("http_method"); ok {
		webhookResponse.HTTPMethod = attr.(string)
	}

	if attr, ok = d.GetOk("http_content_type"); ok {
		webhookResponse.HTTPContentType = attr.(string)
	}

	if attr, ok = d.GetOk("http_headers"); ok {
		webhookResponse.HTTPHeaders = attr.(string)
	}

	if attr, ok = d.GetOk("event"); ok {
		webhookResponse.Event = attr.(string)
	}

	if attr, ok = d.GetOk("enabled"); ok {
		webhookResponse.Enabled = attr.(bool)
	}

	if attr, ok = d.GetOk("verify_ssl"); ok {
		webhookResponse.VerifySSL = attr.(bool)
	}

	if attr, ok = d.GetOk("ssl_ca_certs"); ok {
		webhookResponse.SSLCACerts = attr.(string)
	}

	if attr, ok = d.GetOk("proxy_authorization"); ok {
		webhookResponse.ProxyAuthorization = attr.(bool)
	}

	if attr, ok = d.GetOk("user"); ok {
		webhookResponse.User = attr.(string)
	}

	if attr, ok = d.GetOk("password_set"); ok {
		webhookResponse.PasswordSet = attr.(bool)
	}

	if attr, ok = d.GetOk("WebhookTemplate"); ok {
		webhookResponse.WebhookTemplate = attr.(api.WebhookTemplate)
	}

	return &webhookResponse
}

// setResourceDataFromForemanWebhook sets a ResourceData's attributes from
// the attributes of the supplied ForemanWebhook struct
func setResourceDataFromForemanWebhook(d *schema.ResourceData, fw *api.ForemanWebhook) {
	log.Tracef("resource_foreman_webhook.go#setResourceDataFromForemanWebhook")

	d.SetId(strconv.Itoa(fw.Id))
	d.Set("name", fw.Name)
	d.Set("target_url", fw.TargetURL)
	d.Set("http_method", fw.HTTPMethod)
	d.Set("http_content_type", fw.HTTPContentType)
	d.Set("http_headers", fw.HTTPHeaders)
	d.Set("event", fw.Event)
	d.Set("enabled", fw.Enabled)
	d.Set("verify_ssl", fw.VerifySSL)
	d.Set("ssl_ca_certs", fw.SSLCACerts)
	d.Set("proxy_authorization", fw.ProxyAuthorization)
	d.Set("user", fw.User)
	d.Set("password", fw.Password)
	d.Set("webhook_template_id", fw.WebhookTemplateID)
}

// setResourceDataFromForemanWebhookResponse sets a ResourceData's attributes from
// the attributes of the supplied ForemanWebhookResponse struct
func setResourceDataFromForemanWebhookResponse(d *schema.ResourceData, fw *api.ForemanWebhookResponse) {
	log.Tracef("resource_foreman_webhook.go#setResourceDataFromForemanWebhookResponse")

	d.SetId(strconv.Itoa(fw.Id))
	d.Set("name", fw.Name)
	d.Set("target_url", fw.TargetURL)
	d.Set("http_method", fw.HTTPMethod)
	d.Set("http_content_type", fw.HTTPContentType)
	d.Set("http_headers", fw.HTTPHeaders)
	d.Set("event", fw.Event)
	d.Set("enabled", fw.Enabled)
	d.Set("verify_ssl", fw.VerifySSL)
	d.Set("ssl_ca_certs", fw.SSLCACerts)
	d.Set("proxy_authorization", fw.ProxyAuthorization)
	d.Set("user", fw.User)
	d.Set("password_set", fw.PasswordSet)
	d.Set("webhook_template", fw.WebhookTemplate)
}

// resourceForemanWebhookCreate creates a ForemanWebhook resource
func resourceForemanWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_webhook.go#Create")

	client := meta.(*api.Client)
	h := buildForemanWebhook(d)

	log.Debugf("ForemanWebhook: [%+v]", h)

	createdWebhook, createErr := client.CreateWebhook(ctx, h)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanWebhook: [%+v]", createdWebhook)

	setResourceDataFromForemanWebhook(d, createdWebhook)

	return nil
}

// resourceForemanWebhookRead reads a ForemanWebhook resource
func resourceForemanWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_webhook.go#Read")

	client := meta.(*api.Client)
	h := buildForemanWebhookResponse(d)

	log.Debugf("ForemanWebhook: [%+v]", h)

	readWebhook, readErr := client.ReadWebhook(ctx, h.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanWebhook: [%+v]", readWebhook)

	setResourceDataFromForemanWebhookResponse(d, readWebhook)
	fmt.Printf("Read ForemanWebhook: [%+v]\n", readWebhook)

	return nil
}

// resourceForemanWebhookUpdate updates a ForemanWebhook resource
func resourceForemanWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_webhook.go#Update")

	client := meta.(*api.Client)
	h := buildForemanWebhook(d)

	log.Debugf("ForemanWebhook: [%+v]", h)

	updatedWebhook, updateErr := client.UpdateWebhook(ctx, h)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanWebhook: [%+v]", updatedWebhook)

	setResourceDataFromForemanWebhook(d, updatedWebhook)

	return nil
}

// resourceForemanWebhookDelete deletes a ForemanWebhook resource
func resourceForemanWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_webhook.go#Delete")

	client := meta.(*api.Client)
	h := buildForemanWebhook(d)

	log.Debugf("ForemanWebhook: [%+v]", h)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return diag.FromErr(api.CheckDeleted(d, client.DeleteWebhook(ctx, h.Id)))
}

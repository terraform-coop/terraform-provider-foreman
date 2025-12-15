package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/conv"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanWebhookTemplate() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanWebhookTemplateCreate,
		ReadContext:   resourceForemanWebhookTemplateRead,
		UpdateContext: resourceForemanWebhookTemplateUpdate,
		DeleteContext: resourceForemanWebhookTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: "Webhook templates allow to configure a payload to send via webhook." +
					autodoc.MetaSummary,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
				Description: fmt.Sprintf(
					"Webhook Template name "+
						"%s \"compute\"",
					autodoc.MetaExample,
				),
			},

			"template": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Content of the webhook template.",
			},

			"snippet": {
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
				Description: "Specifies if webhook template is a snippet.",
			},

			"audit_comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Comment for audits.",
			},

			"locked": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether or not the template is locked for editing.",
			},

			"default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether or not the template is added automatically to new organizations and locations.",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Webhook Template description.",
			},

			"location_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional:    true,
				Description: "List of all locations the webhook template can use.",
			},

			"organization_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional:    true,
				Description: "List of all organizations the webhook template can use.",
			},
		},
	}
}

// buildForemanWebhookTemplate constructs a ForemanWebhookTemplate struct from a resource
// data reference. The struct's members are populated from the data populated
// in the resource data. Missing members will be left to the zero value for
// that member's type.
func buildForemanWebhookTemplate(d *schema.ResourceData) *api.ForemanWebhookTemplate {
	log.Tracef("resource_foreman_webhooktemplate.go#buildForemanWebhookTemplate")

	webhookTemplate := api.ForemanWebhookTemplate{}

	obj := buildForemanObject(d)
	webhookTemplate.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		webhookTemplate.Name = attr.(string)
	}

	if attr, ok = d.GetOk("template"); ok {
		webhookTemplate.Template = attr.(string)
	}

	if attr, ok = d.GetOk("snippet"); ok {
		webhookTemplate.Snippet = attr.(bool)
	}

	if attr, ok = d.GetOk("audit_comment"); ok {
		webhookTemplate.AuditComment = attr.(string)
	}

	if attr, ok = d.GetOk("locked"); ok {
		webhookTemplate.Locked = attr.(bool)
	}

	if attr, ok = d.GetOk("default"); ok {
		webhookTemplate.Default = attr.(bool)
	}

	if attr, ok = d.GetOk("description"); ok {
		webhookTemplate.Description = attr.(string)
	}

	if attr, ok = d.GetOk("location_ids"); ok {
		attrSet := attr.(*schema.Set)
		webhookTemplate.LocationIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	if attr, ok = d.GetOk("organization_ids"); ok {
		attrSet := attr.(*schema.Set)
		webhookTemplate.OrganizationIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	return &webhookTemplate
}

// buildForemanWebhookTemplateResponse constructs a ForemanWebhookTemplateResponse struct from a resource
// data reference. The struct's members are populated from the data populated
// in the resource data. Missing members will be left to the zero value for
// that member's type.
func buildForemanWebhookTemplateResponse(d *schema.ResourceData) *api.ForemanWebhookTemplateResponse {
	log.Tracef("resource_foreman_webhooktemplate.go#buildForemanWebhookTemplate")

	webhookTemplateResponse := api.ForemanWebhookTemplateResponse{}

	obj := buildForemanObject(d)
	webhookTemplateResponse.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		webhookTemplateResponse.Name = attr.(string)
	}

	if attr, ok = d.GetOk("template"); ok {
		webhookTemplateResponse.Template = attr.(string)
	}

	if attr, ok = d.GetOk("snippet"); ok {
		webhookTemplateResponse.Snippet = attr.(bool)
	}

	if attr, ok = d.GetOk("audit_comment"); ok {
		webhookTemplateResponse.AuditComment = attr.(string)
	}

	if attr, ok = d.GetOk("locked"); ok {
		webhookTemplateResponse.Locked = attr.(bool)
	}

	if attr, ok = d.GetOk("default"); ok {
		webhookTemplateResponse.Default = attr.(bool)
	}

	if attr, ok = d.GetOk("description"); ok {
		webhookTemplateResponse.Description = attr.(string)
	}

	if attr, ok = d.GetOk("locations"); ok {
		attrSet := attr.(*schema.Set)
		webhookTemplateResponse.Locations = make([]api.EntityResponse, attrSet.Len())
		for i, v := range attrSet.List() {
			webhookTemplateResponse.Locations[i] = api.EntityResponse{ID: v.(int)}
		}
	}

	if attr, ok = d.GetOk("organizations"); ok {
		attrSet := attr.(*schema.Set)
		webhookTemplateResponse.Organizations = make([]api.EntityResponse, attrSet.Len())
		for i, v := range attrSet.List() {
			webhookTemplateResponse.Organizations[i] = api.EntityResponse{ID: v.(int)}
		}
	}

	return &webhookTemplateResponse
}

// setResourceDataFromForemanWebhookTemplate sets a ResourceData's attributes from
// the attributes of the supplied ForemanWebhookTemplate struct
func setResourceDataFromForemanWebhookTemplate(d *schema.ResourceData, fwt *api.ForemanWebhookTemplate) {
	log.Tracef("resource_foreman_webhooktemplate.go#setResourceDataFromForemanWebhookTemplate")

	d.SetId(strconv.Itoa(fwt.Id))
	d.Set("name", fwt.Name)
	d.Set("template", fwt.Template)
	d.Set("snippet", fwt.Snippet)
	d.Set("audit_comment", fwt.AuditComment)
	d.Set("locked", fwt.Locked)
	d.Set("default", fwt.Default)
	d.Set("description", fwt.Description)
	d.Set("location_ids", fwt.LocationIds)
	d.Set("organization_ids", fwt.OrganizationIds)
}

// setResourceDataFromForemanWebhookTemplateResponse sets a ResourceData's attributes from
// the attributes of the supplied ForemanWebhookTemplateResponse struct
func setResourceDataFromForemanWebhookTemplateResponse(d *schema.ResourceData, fwt *api.ForemanWebhookTemplateResponse) {
	log.Tracef("resource_foreman_webhooktemplate.go#setResourceDataFromForemanWebhookTemplateResponse")

	d.SetId(strconv.Itoa(fwt.Id))
	d.Set("name", fwt.Name)
	d.Set("template", fwt.Template)
	d.Set("snippet", fwt.Snippet)
	d.Set("audit_comment", fwt.AuditComment)
	d.Set("locked", fwt.Locked)
	d.Set("default", fwt.Default)
	d.Set("description", fwt.Description)

	locationIDs := make([]int, 0, len(fwt.Locations))
	for _, location := range fwt.Locations {
		locationIDs = append(locationIDs, location.ID)
	}
	d.Set("location_ids", locationIDs)

	organizationIDs := make([]int, 0, len(fwt.Organizations))
	for _, organization := range fwt.Organizations {
		organizationIDs = append(organizationIDs, organization.ID)
	}
	d.Set("organization_ids", organizationIDs)
}

// resourceForemanWebhookTemplateCreate creates a ForemanWebhookTemplate resource
func resourceForemanWebhookTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_webhooktemplate.go#Create")

	client := meta.(*api.Client)
	h := buildForemanWebhookTemplate(d)

	log.Debugf("ForemanWebhookTemplate: [%+v]", h)

	createdWebhookTemplate, createErr := client.CreateWebhookTemplate(ctx, h)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanWebhookTemplate: [%+v]", createdWebhookTemplate)

	setResourceDataFromForemanWebhookTemplate(d, createdWebhookTemplate)

	return nil
}

// resourceForemanWebhookTemplateRead reads a ForemanWebhookTemplate resource
func resourceForemanWebhookTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_webhooktemplate.go#Read")

	client := meta.(*api.Client)
	h := buildForemanWebhookTemplateResponse(d)

	log.Debugf("ForemanWebhookTemplate: [%+v]", h)

	readWebhookTemplate, readErr := client.ReadWebhookTemplate(ctx, h.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanWebhookTemplate: [%+v]", readWebhookTemplate)

	setResourceDataFromForemanWebhookTemplateResponse(d, readWebhookTemplate)
	fmt.Printf("Read ForemanWebhookTemplate: [%+v]\n", readWebhookTemplate)

	return nil
}

// resourceForemanWebhookTemplateUpdate updates a ForemanWebhookTemplate resource
func resourceForemanWebhookTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_webhooktemplate.go#Update")

	client := meta.(*api.Client)
	h := buildForemanWebhookTemplate(d)

	log.Debugf("ForemanWebhookTemplate: [%+v]", h)

	updatedWebhookTemplate, updateErr := client.UpdateWebhookTemplate(ctx, h)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanWebhookTemplate: [%+v]", updatedWebhookTemplate)

	setResourceDataFromForemanWebhookTemplate(d, updatedWebhookTemplate)

	return nil
}

// resourceForemanWebhookTemplateDelete deletes a ForemanWebhookTemplate resource
func resourceForemanWebhookTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_webhooktemplate.go#Delete")

	client := meta.(*api.Client)
	h := buildForemanWebhookTemplate(d)

	log.Debugf("ForemanWebhookTemplate: [%+v]", h)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return diag.FromErr(api.CheckDeleted(d, client.DeleteWebhookTemplate(ctx, h.Id)))
}

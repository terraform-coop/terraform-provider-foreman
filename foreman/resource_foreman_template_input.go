package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanTemplateInput() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanTemplateInputCreate,
		ReadContext:   resourceForemanTemplateInputRead,
		UpdateContext: resourceForemanTemplateInputUpdate,
		DeleteContext: resourceForemanTemplateInputDelete,

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of a template input.",
					autodoc.MetaSummary,
				),
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the template input",
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"template_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"fact_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"variable_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"puppet_parameter_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"puppet_class_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"advanced": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"default": {
				Type:     schema.TypeString,
				Required: true,
			},

			"hidden_value": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"input_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "user",
				ValidateFunc: validation.StringInSlice([]string{
					"user",
					"fact",
					"variable",
				}, false),
			},

			"value_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "plain",
				ValidateFunc: validation.StringInSlice([]string{
					"plain",
					"search",
					"date",
					"resource",
				}, false),
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					log.Debugf("DiffSuppressFunc for template_input value_type: '%s' '%s' '%s'", k, oldValue, newValue)

					// Using only `d.IsNewResource` as check does not work, so we check the Id value as well
					isNew := d.IsNewResource() || d.Id() == ""

					// If this operation is NOT creation, then ignore the value_type field if Terraform tries to
					// "fix" an empty value in Foreman. The key "value_type" is not returned from the Foreman API
					// and is therefore always empty from the providers POV. The newValue is chosen from the default.
					if oldValue == "" && newValue == "plain" && !isNew {
						return true
					}
					return false
				},
			},

			"resource_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func buildForemanTemplateInput(d *schema.ResourceData) *api.ForemanTemplateInput {
	utils.TraceFunctionCall()

	utils.Debugf("buildForemanTemplateInput schema: %+v", d)

	newObj := api.ForemanTemplateInput{}

	fmObj := buildForemanObject(d)
	newObj.ForemanObject = *fmObj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("template_id"); ok {
		newObj.TemplateId = attr.(int)
	}
	if attr, ok = d.GetOk("fact_name"); ok {
		newObj.FactName = attr.(string)
	}
	if attr, ok = d.GetOk("variable_name"); ok {
		newObj.VariableName = attr.(string)
	}
	if attr, ok = d.GetOk("puppet_parameter_name"); ok {
		newObj.PuppetParameterName = attr.(string)
	}
	if attr, ok = d.GetOk("puppet_class_name"); ok {
		newObj.PuppetClassName = attr.(string)
	}
	if attr, ok = d.GetOk("description"); ok {
		newObj.Description = attr.(string)
	}
	if attr, ok = d.GetOk("required"); ok {
		newObj.Required = attr.(bool)
	}
	if attr, ok = d.GetOk("advanced"); ok {
		newObj.Advanced = attr.(bool)
	}
	if attr, ok = d.GetOk("default"); ok {
		newObj.Default = attr.(string)
	}
	if attr, ok = d.GetOk("hidden_value"); ok {
		newObj.HiddenValue = attr.(bool)
	}
	if attr, ok = d.GetOk("input_type"); ok {
		newObj.InputType = attr.(string)
	}
	if attr, ok = d.GetOk("value_type"); ok {
		newObj.ValueType = attr.(string)
	}
	if attr, ok = d.GetOk("resource_type"); ok {
		newObj.ResourceType = attr.(string)
	}

	log.Debugf("newObj: %+v", newObj)

	return &newObj
}

func setResourceDataFromForemanTemplateInput(resdata *schema.ResourceData, ti *api.ForemanTemplateInput) {
	utils.TraceFunctionCall()

	resdata.SetId(strconv.Itoa(ti.Id))

	resdata.Set("name", ti.Name)
	resdata.Set("description", ti.Description)
	resdata.Set("template_id", ti.TemplateId)
	resdata.Set("fact_name", ti.FactName)
	resdata.Set("variable_name", ti.VariableName)
	resdata.Set("puppet_parameter_name", ti.PuppetParameterName)
	resdata.Set("puppet_class_name", ti.PuppetClassName)
	resdata.Set("required", ti.Required)
	resdata.Set("advanced", ti.Advanced)
	resdata.Set("default", ti.Default)
	resdata.Set("hidden_value", ti.HiddenValue)
	resdata.Set("input_type", ti.InputType)
	resdata.Set("value_type", ti.ValueType)
	resdata.Set("resource_type", ti.ResourceType)

	utils.Debugf("resdata after setResourceDataFromForemanTemplateInput: %+v", resdata)
}

// Resource CRUD Operations

func resourceForemanTemplateInputCreate(ctx context.Context, resdata *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	built := buildForemanTemplateInput(resdata)

	created, err := client.CreateTemplateInput(ctx, built)
	if err != nil {
		return diag.FromErr(err)
	}

	setResourceDataFromForemanTemplateInput(resdata, created)

	return nil
}

func resourceForemanTemplateInputRead(ctx context.Context, resdata *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	built := buildForemanTemplateInput(resdata)

	log.Debugf("ForemanTemplateInput: [%+v]", built)

	read, err := client.ReadTemplateInput(ctx, built)
	if err != nil {
		return diag.FromErr(api.CheckDeleted(resdata, err))
	}

	log.Debugf("Read ForemanTemplateInput: [%+v]", read)

	setResourceDataFromForemanTemplateInput(resdata, read)

	return nil
}

func resourceForemanTemplateInputUpdate(ctx context.Context, resdata *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	c := meta.(*api.Client)
	built := buildForemanTemplateInput(resdata)

	updated, err := c.UpdateTemplateInput(ctx, built)
	if err != nil {
		return diag.FromErr(err)
	}

	setResourceDataFromForemanTemplateInput(resdata, updated)

	return nil
}

func resourceForemanTemplateInputDelete(ctx context.Context, resdata *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	built := buildForemanTemplateInput(resdata)

	err := client.DeleteTemplateInput(ctx, built)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

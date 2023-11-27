package foreman

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceForemanJobTemplate() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanJobTemplateCreate,
		ReadContext:   resourceForemanJobTemplateRead,
		UpdateContext: resourceForemanJobTemplateUpdate,
		DeleteContext: resourceForemanJobTemplateDelete,

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of a job template.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the job template",
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"description_format": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"template": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    false,
				Description: "The template content itself",
			},

			"locked": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"job_category": {
				Type:     schema.TypeString,
				Required: true,
			},

			"provider_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"snippet": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"template_inputs": {
				Optional: true,
				ForceNew: true,
				Type:     schema.TypeList,
				Elem:     resourceForemanTemplateInput(),
			},
		},
	}
}

func buildForemanJobTemplate(d *schema.ResourceData) *api.ForemanJobTemplate {
	utils.TraceFunctionCall()

	jt := api.ForemanJobTemplate{}

	obj := buildForemanObject(d)
	jt.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("description"); ok {
		jt.Description = attr.(string)
	}
	if attr, ok = d.GetOk("description_format"); ok {
		jt.DescriptionFormat = attr.(string)
	}
	if attr, ok = d.GetOk("template"); ok {
		jt.Template = attr.(string)
	}
	if attr, ok = d.GetOk("locked"); ok {
		jt.Locked = attr.(bool)
	}
	if attr, ok = d.GetOk("job_category"); ok {
		jt.JobCategory = attr.(string)
	}
	if attr, ok = d.GetOk("provider_type"); ok {
		jt.ProviderType = attr.(string)
	}
	if attr, ok = d.GetOk("snippet"); ok {
		jt.Snippet = attr.(bool)
	}

	// Template inputs with default values
	if attr, ok = d.GetOk("template_inputs"); ok {
		tiList := attr.([]interface{})
		inputs := make([]api.ForemanTemplateInput, len(tiList))

		// utils.Debug("attr: %#v | %+v", attr, attr)
		// utils.Debug("tiList: %#v | %+v", tiList, tiList)

		for idx, tiMap := range tiList {
			if tiMap == nil {
				log.Fatalf("tiMap is nil: %#v", tiMap)
			}

			tiInterface := tiMap.(map[string]interface{})
			jsonBytes, err := json.Marshal(tiInterface)
			if err != nil {
				utils.Fatal(err)
			}

			newObj := api.ForemanTemplateInput{}
			err = json.Unmarshal(jsonBytes, &newObj)
			if err != nil {
				utils.Fatalf("Error in json.Unmarshal: %s", err)
			}

			inputs[idx] = newObj
		}

		jt.TemplateInputs = inputs
	}

	utils.Debug("jt: %+v", jt)

	return &jt
}

func setResourceDataFromForemanJobTemplate(resdata *schema.ResourceData, jt *api.ForemanJobTemplate) {
	utils.TraceFunctionCall()

	resdata.SetId(strconv.Itoa(jt.Id))
	resdata.Set("name", jt.Name)
	resdata.Set("description", jt.Description)
	resdata.Set("description_format", jt.DescriptionFormat)
	resdata.Set("template", jt.Template)
	resdata.Set("locked", jt.Locked)
	resdata.Set("job_category", jt.JobCategory)
	resdata.Set("provider_type", jt.ProviderType)
	resdata.Set("snippet", jt.Snippet)

	utils.Debug("TemplateInputs: %+v", jt.TemplateInputs)

	// At this point the nested TemplateInputs need to be handled.
	// It would be more natively to use `tiList := []schema.ResourceData{}` as the data type and
	// then read it with `resdata.Set("template_inputs", tiList)`, but this does not work.
	// What does work is passing in a simple map[string]interface, which is parsed by `resdata.Set()`

	var tiList []map[string]interface{}

	for _, inputItem := range jt.TemplateInputs {
		mapData := inputItem.ToResourceDataMap(true)
		utils.Debug("mapData: %#v", mapData)
		tiList = append(tiList, mapData)
	}

	utils.Debug("tiList: %+v", tiList)

	err := resdata.Set("template_inputs", tiList)
	if err != nil {
		log.Fatalf("Error in setting resdata template_input: %s", err)
	}

	utils.Debug("resdata setResourceDataFromForemanJobTemplate: %+v", resdata)

	utils.Debug("resdata template_inputs: %+v", resdata.Get("template_inputs"))
}

// Resource CRUD Operations

func resourceForemanJobTemplateCreate(ctx context.Context, resdata *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	jt := buildForemanJobTemplate(resdata)

	created, err := client.CreateJobTemplate(ctx, jt)
	if err != nil {
		return diag.FromErr(err)
	}

	setResourceDataFromForemanJobTemplate(resdata, created)

	return nil
}

func resourceForemanJobTemplateRead(ctx context.Context, resdata *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	jt := buildForemanJobTemplate(resdata)

	log.Debugf("ForemanJobTemplate: [%+v]", jt)

	readJT, readErr := client.ReadJobTemplate(ctx, jt.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(resdata, readErr))
	}

	log.Debugf("Read ForemanJobTemplate: [%+v]", readJT)

	setResourceDataFromForemanJobTemplate(resdata, readJT)

	return nil
}

func resourceForemanJobTemplateUpdate(ctx context.Context, resdata *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	c := meta.(*api.Client)
	jt := buildForemanJobTemplate(resdata)

	updatedJT, err := c.UpdateJobTemplate(ctx, jt)
	if err != nil {
		return diag.FromErr(err)
	}

	setResourceDataFromForemanJobTemplate(resdata, updatedJT)

	return nil
}

func resourceForemanJobTemplateDelete(ctx context.Context, resdata *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	jt := buildForemanJobTemplate(resdata)

	err := client.DeleteJobTemplate(ctx, jt)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

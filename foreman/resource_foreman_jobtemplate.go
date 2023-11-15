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

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

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

	log.Debugf("ResourceData: %+v", d)

	// Template inputs with default values
	if attr, ok = d.GetOk("template_inputs"); ok {
		tiList := attr.([]interface{})
		inputs := make([]api.ForemanTemplateInput, len(tiList))

		for idx, tiMap := range tiList {
			tiInterface := tiMap.(map[string]interface{})
			log.Debugf("%d, %+v", idx, tiInterface)

			jsonBytes, _ := json.Marshal(tiInterface)

			newObj := api.ForemanTemplateInput{}
			json.Unmarshal(jsonBytes, &newObj)
			log.Debugf("%+v", newObj)

			inputs[idx] = newObj
		}

		jt.TemplateInputs = inputs
	}

	utils.Debug("jt: %+v", jt)

	return &jt
}

func setResourceDataFromForemanJobTemplate(resdata *schema.ResourceData, jt *api.ForemanJobTemplate) {
	resdata.SetId(strconv.Itoa(jt.Id))
	resdata.Set("name", jt.Name)
	resdata.Set("description", jt.Description)
	resdata.Set("description_format", jt.DescriptionFormat)
	resdata.Set("template", jt.Template)
	resdata.Set("locked", jt.Locked)
	resdata.Set("job_category", jt.JobCategory)
	resdata.Set("provider_type", jt.ProviderType)
	resdata.Set("snippet", jt.Snippet)

	resdata.Set("template_inputs", jt.TemplateInputs)
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

	// Create template inputs

	utils.Debug("resdata TIs: %+v", resdata.Get("template_inputs"))

	if tis, ok := resdata.GetOk("template_inputs"); ok {
		// Convert interface to list
		tis := tis.([]interface{})

		var tiList []*api.ForemanTemplateInput

		//

		for _, item := range tis {
			utils.Debug("%#v", item)

			converted_rd_map := item.(map[string]interface{})
			// utils.Debug("%#v", converted_rd_map)
			// new_rd := new(schema.ResourceData)
			rdObj := api.ForemanTemplateInput{}

			data, err := json.Marshal(converted_rd_map)
			if err != nil {
				return diag.FromErr(err)
			}

			err = json.Unmarshal(data, &rdObj)
			if err != nil {
				return diag.FromErr(err)
			}

			tiList = append(tiList, &rdObj)

			// for key, val := range converted_rd_map {
			// 	utils.Debug("%#v %#v", key, val)

			// 	err := new_rd.Set(key, val)
			// 	if err != nil {
			// 		log.Fatalf("Error in building ResourceData: %s", err)
			// 	}
			// }

			// ti := buildForemanTemplateInput(new_rd)
			// utils.Debug("built TI: %+v", ti)
		}
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

	err := client.DeleteJobTemplate(ctx, jt.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

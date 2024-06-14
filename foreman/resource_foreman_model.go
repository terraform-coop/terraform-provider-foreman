package foreman

import (
	"context"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceForemanModel() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanModelCreate,
		ReadContext:   resourceForemanModelRead,
		UpdateContext: resourceForemanModelUpdate,
		DeleteContext: resourceForemanModelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Vendor-specific hardware model.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The name of the hardware model. "+
						"%s \"PowerEdge FC630\"",
					autodoc.MetaExample,
				),
			},

			"info": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional information about this hardware model.",
			},

			"vendor_class": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name or class of the hardware vendor.",
			},

			"hardware_model": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the specific hardware model.",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanModel constructs a ForemanModel struct from a resource data
// reference.  The struct's members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanModel(d *schema.ResourceData) *api.ForemanModel {
	utils.TraceFunctionCall()

	model := api.ForemanModel{}

	obj := buildForemanObject(d)
	model.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("info"); ok {
		model.Info = attr.(string)
	}

	if attr, ok = d.GetOk("vendor_class"); ok {
		model.VendorClass = attr.(string)
	}

	if attr, ok = d.GetOk("hardware_model"); ok {
		model.HardwareModel = attr.(string)
	}

	return &model
}

// setResourceDataFromForemanModel sets a ResourceData's attributes from the
// attributes of the supplied ForemanModel struct
func setResourceDataFromForemanModel(d *schema.ResourceData, fm *api.ForemanModel) {
	utils.TraceFunctionCall()

	d.SetId(strconv.Itoa(fm.Id))
	d.Set("name", fm.Name)
	d.Set("info", fm.Info)
	d.Set("vendor_class", fm.VendorClass)
	d.Set("hardware_model", fm.HardwareModel)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanModelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	utils.Debugf("ForemanModel: [%+v]", m)

	createdModel, createErr := client.CreateModel(ctx, m)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	utils.Debugf("Created ForemanModel: [%+v]", createdModel)

	setResourceDataFromForemanModel(d, createdModel)

	return nil
}

func resourceForemanModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	utils.Debugf("ForemanModel: [%+v]", m)

	readModel, readErr := client.ReadModel(ctx, m.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	utils.Debugf("Read ForemanModel: [%+v]", readModel)

	setResourceDataFromForemanModel(d, readModel)

	return nil
}

func resourceForemanModelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	utils.Debugf("ForemanModel: [%+v]", m)

	updatedModel, updateErr := client.UpdateModel(ctx, m)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	utils.Debugf("Updated ForemanModel: [%+v]", updatedModel)

	setResourceDataFromForemanModel(d, updatedModel)

	return nil
}

func resourceForemanModelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	m := buildForemanModel(d)

	utils.Debugf("ForemanModel: [%+v]", m)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return diag.FromErr(api.CheckDeleted(d, client.DeleteModel(ctx, m.Id)))
}

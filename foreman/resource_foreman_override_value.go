package foreman

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceForemanOverrideValue() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanOverrideValueCreate,
		ReadContext:   resourceForemanOverrideValueRead,
		UpdateContext: resourceForemanOverrideValueUpdate,
		DeleteContext: resourceForemanOverrideValueDelete,

		// TODO - passthrough cannot be used as d.Id() is not sufficient to retrieve the resource
		// Importer: &schema.ResourceImporter{
		// 	StateContext: schema.ImportStatePassthroughContext,
		// },

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Smart class parameter override value.",
					autodoc.MetaSummary,
				),
			},
			"match": &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				Description: fmt.Sprintf(
					"A map containing the match criteria. Must contain two keys: `type` and `value`."+
						"Type can be one of `fqdn`, `hostgroup`, `domain` or `os`"+
						"%s {\n    type = \"hostgroup\"\n    value = \"example_group\"\n  }",
					autodoc.MetaExample,
				),
			},
			"omit": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: fmt.Sprintf(
					"When set to `true` Foreman will not send this parameter in classification output. "+
						"Default value is `false`."+
						"%s false",
					autodoc.MetaExample,
				),
			},
			"value": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Smart parameter override value. Hashes and arrays must be JSON encoded."+
						"%s jsonencode({\n    key = \"value\"\n  })",
					autodoc.MetaExample,
				),
			},
			"smart_class_parameter_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				Description: fmt.Sprintf(
					"ID of the smart class parameter to override."+
						"%s 1",
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
func buildForemanOverrideValue(d *schema.ResourceData) *api.ForemanOverrideValue {
	log.Tracef("resource_foreman_override_value.go#buildForemanOverrideValue")

	override := api.ForemanOverrideValue{}

	obj := buildForemanObject(d)
	override.ForemanObject = *obj

	override.Omit = d.Get("omit").(bool)
	override.Value = d.Get("value").(string)
	override.SmartClassParameterId = d.Get("smart_class_parameter_id").(int)

	if _, ok := d.GetOk("match"); ok {
		matchParams := d.Get("match").(map[string]interface{})
		if val, ok := matchParams["type"]; ok {
			override.MatchType = val.(string)
		} else {
			override.MatchType = ""
		}
		if val, ok := matchParams["value"]; ok {
			override.MatchValue = val.(string)
		} else {
			override.MatchValue = ""
		}
	}

	return &override
}

// setResourceDataFromForemanKatelloProduct sets a ResourceData's attributes from
// the attributes of the supplied ForemanKatelloProduct struct
func setResourceDataFromForemanOverrideValue(d *schema.ResourceData, override *api.ForemanOverrideValue) {
	log.Tracef("resource_foreman_override_value.go#setResourceDataFromForemanOverrideValue")

	d.SetId(strconv.Itoa(override.Id))
	d.Set("omit", override.Omit)
	d.Set("value", override.Value)
	d.Set("smart_class_parameter_id", override.SmartClassParameterId)

	matchMap := map[string]interface{}{}
	matchMap["type"] = override.MatchType
	matchMap["value"] = override.MatchValue
	d.Set("match", matchMap)

}

// Validate that the match map contains the required keys. We have to perform this manually
// as Terraform does not support validation of complex types.
func validateOverrideMatchMap(d *schema.ResourceData) error {
	log.Tracef("resource_foreman_override_value.go#validateOverrideMatchMap")

	matchParams := d.Get("match").(map[string]interface{})

	if _, ok := matchParams["type"].(string); !ok {
		return errors.New("override match type must be set!")
	}
	if _, ok := matchParams["value"].(string); !ok {
		return errors.New("override match value must be set!")
	}

	return nil
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanOverrideValueCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_override_value.go#Create")

	valErr := validateOverrideMatchMap(d)
	if valErr != nil {
		return diag.FromErr(valErr)
	}

	client := meta.(*api.Client)
	override := buildForemanOverrideValue(d)

	log.Debugf("ForemanOverrideValue: [%+v]", override)

	createdOverrideValue, createErr := client.CreateOverrideValue(ctx, override)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanOverrideValue: [%+v]", createdOverrideValue)

	setResourceDataFromForemanOverrideValue(d, createdOverrideValue)

	return nil
}

func resourceForemanOverrideValueRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_override_value.go#Read")

	client := meta.(*api.Client)
	override := buildForemanOverrideValue(d)

	log.Debugf("ForemanOverrideValue: [%+v]", override)

	readOverrideValue, readErr := client.ReadOverrideValue(ctx, override.Id, override.SmartClassParameterId)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanOverrideValue: [%+v]", readOverrideValue)

	setResourceDataFromForemanOverrideValue(d, readOverrideValue)

	return nil
}

func resourceForemanOverrideValueUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_override_value.go#Update")

	valErr := validateOverrideMatchMap(d)
	if valErr != nil {
		return diag.FromErr(valErr)
	}

	client := meta.(*api.Client)
	override := buildForemanOverrideValue(d)

	log.Debugf("ForemanOverrideValue: [%+v]", override)

	updatedOverrideValue, updateErr := client.UpdateOverrideValue(ctx, override)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("ForemanOverrideValue: [%+v]", updatedOverrideValue)

	setResourceDataFromForemanOverrideValue(d, updatedOverrideValue)

	return nil
}

func resourceForemanOverrideValueDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_override_value.go#Delete")

	client := meta.(*api.Client)
	override := buildForemanOverrideValue(d)

	log.Debugf("ForemanOverrideValue: [%+v]", override)

	return diag.FromErr(api.CheckDeleted(d, client.DeleteOverrideValue(ctx, override.Id, override.SmartClassParameterId)))
}

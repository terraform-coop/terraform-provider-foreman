package foreman

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceForemanOverrideValue() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanOverrideValueCreate,
		Read:   resourceForemanOverrideValueRead,
		Update: resourceForemanOverrideValueUpdate,
		Delete: resourceForemanOverrideValueDelete,

		// TODO - passthrough cannot be used as d.Id() is not sufficient to retrieve the resource
		// Importer: &schema.ResourceImporter{
		// 	State: schema.ImportStatePassthrough,
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

	Override := api.ForemanOverrideValue{}

	obj := buildForemanObject(d)
	Override.ForemanObject = *obj

	Override.Omit = d.Get("omit").(bool)
	Override.Value = d.Get("value").(string)
	Override.SmartClassParameterId = d.Get("smart_class_parameter_id").(int)

	if _, ok := d.GetOk("match"); ok {
		matchParams := d.Get("match").(map[string]interface{})
		if val, ok := matchParams["type"]; ok {
			Override.MatchType = val.(string)
		} else {
			Override.MatchType = ""
		}
		if val, ok := matchParams["value"]; ok {
			Override.MatchValue = val.(string)
		} else {
			Override.MatchValue = ""
		}
	}

	return &Override
}

// setResourceDataFromForemanKatelloProduct sets a ResourceData's attributes from
// the attributes of the supplied ForemanKatelloProduct struct
func setResourceDataFromForemanOverrideValue(d *schema.ResourceData, Override *api.ForemanOverrideValue) {
	log.Tracef("resource_foreman_override_value.go#setResourceDataFromForemanOverrideValue")

	d.SetId(strconv.Itoa(Override.Id))
	d.Set("omit", Override.Omit)
	d.Set("value", Override.Value)
	d.Set("smart_class_parameter_id", Override.SmartClassParameterId)

	matchMap := map[string]interface{}{}
	matchMap["type"] = Override.MatchType
	matchMap["value"] = Override.MatchValue
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

func resourceForemanOverrideValueCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_override_value.go#Create")

	valErr := validateOverrideMatchMap(d)
	if valErr != nil {
		return valErr
	}

	client := meta.(*api.Client)
	Override := buildForemanOverrideValue(d)

	log.Debugf("ForemanOverrideValue: [%+v]", Override)

	createdOverrideValue, createErr := client.CreateOverrideValue(Override)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanOverrideValue: [%+v]", createdOverrideValue)

	setResourceDataFromForemanOverrideValue(d, createdOverrideValue)

	return nil
}

func resourceForemanOverrideValueRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_override_value.go#Read")

	client := meta.(*api.Client)
	Override := buildForemanOverrideValue(d)

	log.Debugf("ForemanOverrideValue: [%+v]", Override)

	readOverrideValue, readErr := client.ReadOverrideValue(Override.Id, Override.SmartClassParameterId)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanOverrideValue: [%+v]", readOverrideValue)

	setResourceDataFromForemanOverrideValue(d, readOverrideValue)

	return nil
}

func resourceForemanOverrideValueUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_override_value.go#Update")

	valErr := validateOverrideMatchMap(d)
	if valErr != nil {
		return valErr
	}

	client := meta.(*api.Client)
	Override := buildForemanOverrideValue(d)

	log.Debugf("ForemanOverrideValue: [%+v]", Override)

	updatedOverrideValue, updateErr := client.UpdateOverrideValue(Override)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("ForemanOverrideValue: [%+v]", updatedOverrideValue)

	setResourceDataFromForemanOverrideValue(d, updatedOverrideValue)

	return nil
}

func resourceForemanOverrideValueDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_override_value.go#Delete")

	client := meta.(*api.Client)
	Override := buildForemanOverrideValue(d)

	log.Debugf("ForemanOverrideValue: [%+v]", Override)

	return client.DeleteOverrideValue(Override.Id, Override.SmartClassParameterId)
}

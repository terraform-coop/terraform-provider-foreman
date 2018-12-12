package foreman

import (
	"fmt"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/conv"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceForemanProvisioningTemplate() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanProvisioningTemplateCreate,
		Read:   resourceForemanProvisioningTemplateRead,
		Update: resourceForemanProvisioningTemplateUpdate,
		Delete: resourceForemanProvisioningTemplateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Provisioning templates are scripts used to describe how to "+
						"bootstrap and install the operating system on the host.",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Name of the provisioning template. "+
						"%s \"AutoYaST default\"",
					autodoc.MetaExample,
				),
			},

			"template": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"The markup and code of the provisioning template. "+
						"%s \"void\"",
					autodoc.MetaExample,
				),
			},

			"snippet": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Whether or not the provisioning template is a snippet " +
					"be used by other templates.",
			},

			"audit_comment": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Notes and comments for auditing purposes.",
			},

			"locked": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether or not the template is locked for editing.",
			},

			"template_kind_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "ID of the template kind which categorizes the " +
					"provisioning template. Optional for snippets, otherwise required.",
			},

			// -- Foreign Key Relationships --

			"operatingsystem_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the operating systems associated with this " +
					"provisioning template.",
			},

			"template_combinations_attributes": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     resourceForemanTemplateCombinationsAttributes(),
				Set:      schema.HashResource(resourceForemanTemplateCombinationsAttributes()),
				Description: "How templates are determined:\n\n" +
					"When editing a template, you must assign a list of operating systems " +
					"which this template can be used with.  Optionally, you can restrict " +
					"a template to a list of host groups and/or environments.\n\n" +
					"When a host requests a template, Foreman will select the best match " +
					"from the available templates of that type in the following order:\n\n" +
					"  1. host group and environment\n" +
					"  2. host group only\n" +
					"  3. environment only\n" +
					"  4. operating system default\n\n" +
					"Template combinations attributes contains an array of hostgroup IDs " +
					"and environment ID combinations so they can be used in the " +
					"provisioning template selection described above.",
			},
		},
	}
}

// resourceForemanTemplateCombinationsAttributes is a nested resource that
// represents a valid template combination attribute.  The "id" of this
// resource is computed and assigned by Foreman at the time of creation.
//
// NOTE(ALL): See comments in ResourceData's "template_combinations_attributes"
//   attribute definition above
func resourceForemanTemplateCombinationsAttributes() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Template combination unique identifier.",
			},
			"hostgroup_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "The hostgroup ID for this template combination.",
			},
			"environment_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "The environment ID for this template combination.",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanProvisioningTemplate constructs a ForemanProvisioningTemplate
// struct from a resource data reference.  The struct's members are populated
// with the data populated in the resource data.  Missing members will be left
// to the zero value for that member's type.
func buildForemanProvisioningTemplate(d *schema.ResourceData) *api.ForemanProvisioningTemplate {
	log.Tracef("resource_foreman_provisioningtemplate.go#buildForemanProvisioningTemplate")

	template := api.ForemanProvisioningTemplate{}

	obj := buildForemanObject(d)
	template.ForemanObject = *obj

	var attr interface{}
	var ok bool

	template.Template = d.Get("template").(string)

	if attr, ok = d.GetOk("snippet"); ok {
		template.Snippet = attr.(bool)
	}
	if attr, ok = d.GetOk("audit_comment"); ok {
		template.AuditComment = attr.(string)
	}
	if attr, ok = d.GetOk("locked"); ok {
		template.Locked = attr.(bool)
	}
	if attr, ok = d.GetOk("template_kind_id"); ok {
		template.TemplateKindId = attr.(int)
	}
	if attr, ok = d.GetOk("operatingsystem_ids"); ok {
		attrSet := attr.(*schema.Set)
		template.OperatingSystemIds = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	template.TemplateCombinationsAttributes = buildForemanTemplateCombinationsAttributes(d)

	return &template
}

// buildForemanTemplateCombinationsAttributes constructs an array of
// ForemanTemplateCombinationAttribute structs from a resource data reference.
// The struct's members are populated with the data populated in the resource
// data. Missing members will be left to the zero value for that member's type.
func buildForemanTemplateCombinationsAttributes(d *schema.ResourceData) []api.ForemanTemplateCombinationAttribute {
	log.Tracef("resource_foreman_provisioningtemplate.go#buildForemanTemplateCombinationsAttributes")

	tempComboAttr := []api.ForemanTemplateCombinationAttribute{}
	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("template_combinations_attributes"); !ok {
		return tempComboAttr
	}

	// type assert the underlying *schema.Set and convert to a list
	attrSet := attr.(*schema.Set)
	attrList := attrSet.List()
	attrListLen := len(attrList)
	tempComboAttr = make([]api.ForemanTemplateCombinationAttribute, attrListLen)
	// iterate over each of the map structure entires in the set and convert that
	// to a concrete struct implementation to append to the template combinations
	// attributes list.
	for idx, attrMap := range attrList {
		tempComboAttrMap := attrMap.(map[string]interface{})
		tempComboAttr[idx] = mapToForemanTemplateCombinationAttribute(tempComboAttrMap)
	}

	return tempComboAttr
}

// mapToForemanTemplateCombinationAttribute converts a map[string]interface{}
// to a ForemanTemplateCombinationAttribute struct.  The supplied map comes
// from an entry in the *schema.Set for the "template_combinations_attributes"
// property of the resource, since *schema.Set stores its entries as this map
// structure.
//
// The map should have the following keys. Omitted or invalid map values will
// result in the struct receiving the zero value for that property.
//
//   id (int)
//   hostgroup_id (int)
//   environment_id (int)
//   _destroy (bool)
func mapToForemanTemplateCombinationAttribute(m map[string]interface{}) api.ForemanTemplateCombinationAttribute {
	log.Tracef("mapToForemanTemplateCombinationAttribute")

	tempComboAttr := api.ForemanTemplateCombinationAttribute{}
	var ok bool

	if tempComboAttr.Id, ok = m["id"].(int); !ok {
		tempComboAttr.Id = 0
	}

	if tempComboAttr.HostgroupId, ok = m["hostgroup_id"].(int); !ok {
		tempComboAttr.HostgroupId = 0
	}

	if tempComboAttr.EnvironmentId, ok = m["environment_id"].(int); !ok {
		tempComboAttr.EnvironmentId = 0
	}

	if tempComboAttr.Destroy, ok = m["_destroy"].(bool); !ok {
		tempComboAttr.Destroy = false
	}

	log.Debugf("m: [%v], tempComboAttr: [%+v]", m, tempComboAttr)
	return tempComboAttr
}

// setResourceDataFromForemanProvisioningTemplate sets a ResourceData's
// attributes from the attributes of the supplied ForemanProvisioningTemplate
// struct
func setResourceDataFromForemanProvisioningTemplate(d *schema.ResourceData, ft *api.ForemanProvisioningTemplate) {
	log.Tracef("resource_foreman_provisioningtemplate.go#setResourceDataFromForemanProvisioningTemplate")

	d.SetId(strconv.Itoa(ft.Id))

	d.Set("name", ft.Name)
	d.Set("template", ft.Template)
	d.Set("snippet", ft.Snippet)
	d.Set("audit_comment", ft.AuditComment)
	d.Set("locked", ft.Locked)

	d.Set("template_kind_id", ft.TemplateKindId)
	d.Set("operatingsystem_ids", ft.OperatingSystemIds)

	setResourceDataFromForemanTemplateCombinationsAttributes(d, ft.TemplateCombinationsAttributes)

}

// setResourceDataFromForemanTemplateCombinationsAttributes sets a
// ResourceData's "template_combinations_attributes" attribute to the value of
// the supplied array of ForemanTemplateCombinationAttribute structs
func setResourceDataFromForemanTemplateCombinationsAttributes(d *schema.ResourceData, ftca []api.ForemanTemplateCombinationAttribute) {
	log.Tracef("resource_foreman_provisioningtemplate.go#setResourceDataFromForemanTemplateCombinationsAttriutes")

	// this attribute is a *schema.Set.  In order to construct a set, we need to
	// supply a hash function so the set can differentiate for uniqueness of
	// entries.  The hash function will be based on the resource definition
	hashFunc := schema.HashResource(resourceForemanTemplateCombinationsAttributes())
	// underneath, a *schema.Set stores an array of map[string]interface{} entries.
	// convert each ForemanTemplateCombination struct in the supplied array to a
	// mapstructure and then add it to the set
	ifaceArr := make([]interface{}, len(ftca))
	for idx, val := range ftca {
		// NOTE(ALL): we ommit the "_destroy" property here - this does not need
		//   to be stored by terraform in the state file. That is a hidden key that
		//   is only used in updates.  Anything that exists will always have it
		//   set to "false".
		ifaceMap := map[string]interface{}{
			"id":             val.Id,
			"hostgroup_id":   val.HostgroupId,
			"environment_id": val.EnvironmentId,
		}
		ifaceArr[idx] = ifaceMap
	}
	// with the array set up, create the *schema.Set and set the ResourceData's
	// "template_combinations_attributes" property
	tempComboAttrSet := schema.NewSet(hashFunc, ifaceArr)
	d.Set("template_combinations_attributes", tempComboAttrSet)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanProvisioningTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_provisioningtemplate.go#Create")

	client := meta.(*api.Client)
	t := buildForemanProvisioningTemplate(d)

	log.Debugf("ForemanProvisioningTemplate: [%+v]", t)

	createdTemplate, createErr := client.CreateProvisioningTemplate(t)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanProvisioningTemplate: [%+v]", createdTemplate)

	setResourceDataFromForemanProvisioningTemplate(d, createdTemplate)

	return nil
}

func resourceForemanProvisioningTemplateRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_provisioningtemplate.go#Read")

	client := meta.(*api.Client)
	t := buildForemanProvisioningTemplate(d)

	log.Debugf("ForemanProvisioningTemplate: [%+v]", t)

	readTemplate, readErr := client.ReadProvisioningTemplate(t.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanProvisioningTemplate: [%+v]", readTemplate)

	setResourceDataFromForemanProvisioningTemplate(d, readTemplate)

	return nil
}

func resourceForemanProvisioningTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_provisioningtemplate.go#Update")

	client := meta.(*api.Client)
	t := buildForemanProvisioningTemplate(d)

	log.Debugf("ForemanProvisioningTemplate: [%+v]", t)

	// NOTE(ALL): Handling the removal of a template combination.  See the note
	//   in ForemanTemplateCombinationAttribute's Destroy property
	if d.HasChange("template_combinations_attributes") {
		oldVal, newVal := d.GetChange("template_combinations_attributes")
		oldValSet, newValSet := oldVal.(*schema.Set), newVal.(*schema.Set)

		// NOTE(ALL): The set difference operation is anticommutative (because math)
		//   ie: [A - B] =/= [B - A].
		//
		//   When performing an update, we need to figure out which template
		//   combinations were removed from the set and tag the destroy property
		//   to true and instruct Foreman which ones to delete from the list. We do
		//   this by performing a set difference between the old set and the new
		//   set (ie: [old - new]) which will return the items that used to be in
		//   the set but are no longer included.
		//
		//   The values that were added to the set or remained unchanged are already
		//   part of the template combinations.  They are present in the
		//   ResourceData and already exist from the
		//   buildForemanProvisioningTemplate() call.

		setDiff := oldValSet.Difference(newValSet)
		setDiffList := setDiff.List()
		log.Debugf("setDiffList: [%v]", setDiffList)

		// iterate over the removed items, add them back to the template's
		// combination array, but tag them for removal.
		//
		// each of the set's items is stored as a map[string]interface{} - use
		// type assertion and construct the struct
		for _, rmVal := range setDiffList {
			// construct, tag for deletion from list of combinations
			rmValMap := rmVal.(map[string]interface{})
			rmCombination := mapToForemanTemplateCombinationAttribute(rmValMap)
			rmCombination.Destroy = true
			// append back to template's list
			t.TemplateCombinationsAttributes = append(t.TemplateCombinationsAttributes, rmCombination)
		}

		log.Debugf("ForemanProvisioningTemplate: [%+v]", t)

	} // end HasChange("template_combinations_attributes")

	updatedTemplate, updateErr := client.UpdateProvisioningTemplate(t)
	if updateErr != nil {
		return updateErr
	}

	log.Debugf("Updated ForemanProvisioningTemplate: [%+v]", t)

	setResourceDataFromForemanProvisioningTemplate(d, updatedTemplate)

	return nil
}

func resourceForemanProvisioningTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_provisioningtemplate.go#Delete")

	client := meta.(*api.Client)
	t := buildForemanProvisioningTemplate(d)

	log.Debugf("ForemanProvisioningTemplate: [%+v]", t)

	// NOTE(ALL): The Foreman API will return a '422: Unprocessable Entity' error
	//   if you try to delete a provisioning template with template combinations.
	//   First, you must update the provisioning template to remove the combinations,
	//   then proceed with deletion.
	if len(t.TemplateCombinationsAttributes) > 0 {
		log.Debugf("deleting template that has combinations set")
		// iterate through each of the template combinations and tag them for
		// removal from the list
		for idx, _ := range t.TemplateCombinationsAttributes {
			t.TemplateCombinationsAttributes[idx].Destroy = true
		}
		log.Debugf("ForemanProvisioningTemplate: [%+v]", t)

		updatedTemplate, updateErr := client.UpdateProvisioningTemplate(t)
		if updateErr != nil {
			return updateErr
		}

		log.Debugf("Updated ForemanProvisioningTemplate: [%+v]", updatedTemplate)

		// NOTE(ALL): set the resource data's properties to what comes back from
		//   the update call. This allows us to recover from a partial state if
		//   delete encounters an error after this point - at least the resource's
		//   state will be saved with the correct template combinations.
		setResourceDataFromForemanProvisioningTemplate(d, updatedTemplate)

		log.Debugf("completed the template combination deletion")

	} // end if len(template.TemplateCombinationsAttributes) > 0

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return client.DeleteProvisioningTemplate(t.Id)
}

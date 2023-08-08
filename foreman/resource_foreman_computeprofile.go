package foreman

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
)

func resourceForemanComputeProfile() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanComputeprofileCreate,
		ReadContext:   resourceForemanComputeprofileRead,
		UpdateContext: resourceForemanComputeprofileUpdate,
		DeleteContext: resourceForemanComputeprofileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of a compute profile.",
					autodoc.MetaSummary,
				),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the compute profile",
			},
			"compute_attributes": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of compute attributes",
				Elem:        resourceForemanComputeAttribute(),
			},
		},
	}
}

func resourceForemanComputeAttribute() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"compute_resource_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the compute resource",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
			},
			"vm_attrs": {
				Type:        schema.TypeMap,
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "VM attributes as JSON",
			},
		},
	}
}

// buildForemanComputeProfile constructs a ForemanComputeProfile reference from a
// resource data reference.  The struct's  members are populated from the data
// populated in the resource data.  Missing members will be left to the zero
// value for that member's type.
func buildForemanComputeProfile(d *schema.ResourceData) *api.ForemanComputeProfile {
	log.Tracef("foreman/resource_foreman_computeprofile.go#buildForemanComputeProfile")

	t := api.ForemanComputeProfile{}
	obj := buildForemanObject(d)
	t.ForemanObject = *obj

	// log.Debugf("compute_attributes: %+v", d.Get("compute_attributes"))
	caList := d.Get("compute_attributes").([]interface{})
	var compattrObjList []*api.ForemanComputeAttribute

	// log.Debugf("compattrObjList: %+v", compattrObjList)

	for i := 0; i < len(caList); i++ {
		ca := caList[i].(map[string]interface{})
		caObj := new(api.ForemanComputeAttribute)
		// log.Debugf("caObj: %+v", ca)

		data, err := json.Marshal(ca)
		if err != nil {
			return nil
		}

		err = json.Unmarshal(data, caObj)
		if err != nil {
			log.Warningf("Error during json.Unmarshal: %s", err)
			return nil
		}

		// log.Debugf("caObjStr: %+v", caObj)

		compattrObjList = append(compattrObjList, caObj)
	}

	t.ComputeAttributes = compattrObjList

	// log.Debugf("t.ComputeAttributes: %+v", t.ComputeAttributes)

	return &t
}

// setResourceDataFromForemanComputeProfile sets a ResourceData's attributes from
// the attributes of the supplied ForemanComputeProfile reference
func setResourceDataFromForemanComputeProfile(d *schema.ResourceData, fk *api.ForemanComputeProfile) {
	log.Tracef("foreman/resource_foreman_computeprofile.go#setResourceDataFromForemanComputeProfile")

	d.SetId(strconv.Itoa(fk.Id))
	d.Set("name", fk.Name)

	var caList []interface{}
	for i := 0; i < len(fk.ComputeAttributes); i++ {
		caObj := *fk.ComputeAttributes[i]
		data, err := json.Marshal(caObj)
		if err != nil {
			log.Errorf("Error in json.Marshal: %s", err)
		}

		t := make(map[string]interface{})
		err = json.Unmarshal(data, &t)
		if err != nil {
			log.Errorf("Error in json.Unmarshal: %s", err)
		}
		caList = append(caList, t)
	}

	log.Debugf("setResourceDataFromForemanComputeProfile caList: %+v", caList...)

	d.Set("compute_attributes", caList)
}

func resourceForemanComputeprofileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("foreman/resource_foreman_computeprofile.go#resourceForemanComputeprofileCreate")

	client := meta.(*api.Client)
	p := buildForemanComputeProfile(d)

	createdComputeprofile, createErr := client.CreateComputeprofile(ctx, p)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanComputeprofile [%+v]", createdComputeprofile)

	setResourceDataFromForemanComputeProfile(d, createdComputeprofile)

	return nil
}

func resourceForemanComputeprofileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("foreman/resource_foreman_computeprofile.go#resourceForemanComputeprofileRead")

	client := meta.(*api.Client)
	p := buildForemanComputeProfile(d)

	cp, err := client.ReadComputeProfile(ctx, p.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Debugf("Read compute_profile: %+v", cp)

	setResourceDataFromForemanComputeProfile(d, cp)

	return nil
}

func resourceForemanComputeprofileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("foreman/resource_foreman_computeprofile.go#resourceForemanComputeprofileUpdate")
	return nil
}

func resourceForemanComputeprofileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("foreman/resource_foreman_computeprofile.go#resourceForemanComputeprofileDelete")

	client := meta.(*api.Client)
	p := buildForemanComputeProfile(d)

	err := client.DeleteComputeProfile(ctx, p.Id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanParameter() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanParameterCreate,
		ReadContext:   resourceForemanParameterRead,
		UpdateContext: resourceForemanParameterUpdate,
		DeleteContext: resourceForemanParameterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of parameter. Parameters serve as an "+
						"identification string that defines autonomy, authority, or control "+
						"for a portion of a network.",
					autodoc.MetaSummary,
				),
			},

			"host_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "ID of the host to assign this parameter to",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"hostgroup_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "ID of the host group to assign this parameter to",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"operatingsystem_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "ID of the operating system to assign this parameter to",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"domain_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "ID of the domain to assign this parameter to",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"subnet_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Description:  "ID of the subnet to assign this parameter to",
				ValidateFunc: validation.IntAtLeast(1),
			},
			// -- Actual Content --
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"value": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanParameter constructs a ForemanParameter reference from a resource data
// reference.  The struct's  members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanParameter(d *schema.ResourceData) *api.ForemanParameter {
	log.Tracef("resource_foreman_parameter.go#buildForemanParameter")

	parameter := api.ForemanParameter{}

	obj := buildForemanObject(d)
	parameter.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("host_id"); ok {
		parameter.HostID = attr.(int)
	}
	if attr, ok = d.GetOk("hostgroup_id"); ok {
		parameter.HostGroupID = attr.(int)
	}
	if attr, ok = d.GetOk("domain_id"); ok {
		parameter.DomainID = attr.(int)
	}
	if attr, ok = d.GetOk("operatingsystem_id"); ok {
		parameter.OperatingSystemID = attr.(int)
	}
	if attr, ok = d.GetOk("subnet_id"); ok {
		parameter.SubnetID = attr.(int)
	}
	if attr, ok = d.GetOk("name"); ok {
		parameter.Parameter.Name = attr.(string)
	}
	if attr, ok = d.GetOk("value"); ok {
		parameter.Parameter.Value = attr.(string)
	}
	return &parameter
}

// setResourceDataFromForemanParameter sets a ResourceData's attributes from the
// attributes of the supplied ForemanParameter reference
func setResourceDataFromForemanParameter(d *schema.ResourceData, fd *api.ForemanParameter) {
	log.Tracef("resource_foreman_parameter.go#setResourceDataFromForemanParameter")

	d.SetId(strconv.Itoa(fd.Id))
	d.Set("host_id", fd.HostID)
	d.Set("hostgroup_id", fd.HostGroupID)
	d.Set("domain_id", fd.DomainID)
	d.Set("operatingsystem_id", fd.OperatingSystemID)
	d.Set("subnet_id", fd.SubnetID)
	d.Set("name", fd.Parameter.Name)
	d.Set("value", fd.Parameter.Value)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanParameterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_parameter.go#Create")

	client := meta.(*api.Client)
	p := buildForemanParameter(d)

	log.Debugf("ForemanParameter: [%+v]", d)

	createdParam, createErr := client.CreateParameter(ctx, p)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanParameter: [%+v]", createdParam)

	setResourceDataFromForemanParameter(d, createdParam)

	return nil
}

func resourceForemanParameterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_parameter.go#Read")

	client := meta.(*api.Client)
	parameter := buildForemanParameter(d)

	log.Debugf("ForemanParameter: [%+v]", parameter)

	readParameter, readErr := client.ReadParameter(ctx, parameter, parameter.Id)
	if readErr != nil {
		return diag.FromErr(readErr)
	}

	log.Debugf("Read ForemanParameter: [%+v]", readParameter)

	setResourceDataFromForemanParameter(d, readParameter)

	return nil
}

func resourceForemanParameterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_parameter.go#Update")

	client := meta.(*api.Client)
	p := buildForemanParameter(d)

	log.Debugf("ForemanParameter: [%+v]", p)

	updatedParam, updateErr := client.UpdateParameter(ctx, p, p.Id)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanParameter: [%+v]", updatedParam)

	setResourceDataFromForemanParameter(d, updatedParam)

	return nil
}

func resourceForemanParameterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_parameter.go#Delete")

	client := meta.(*api.Client)
	p := buildForemanParameter(d)

	log.Debugf("ForemanParameter: [%+v]", p)

	return diag.FromErr(client.DeleteParameter(ctx, p, p.Id))
}

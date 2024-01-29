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

func resourceForemanComputeResource() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanComputeResourceCreate,
		ReadContext:   resourceForemanComputeResourceRead,
		UpdateContext: resourceForemanComputeResourceUpdate,
		DeleteContext: resourceForemanComputeResourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of computeresource. ComputeResources serve as an "+
						"identification string that defines autonomy, authority, or control "+
						"for a portion of a network.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the compute resource",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "URL for Libvirt, oVirt, OpenStack and Rackspace",
			},
			"hypervisor": {
				Type:     schema.TypeString,
				Required: true,
				Description: "The HyperVisor/Cloud Provider for this Compute Resource:" +
					"supported providers include \"Libvirt\", \"Ovirt\", \"EC2\"," +
					"\"Vmware\", \"Openstack\", \"Rackspace\", \"GCE\"",
			},
			"displaytype": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "For Libvirt: \"VNC\" or \"SPICE\". For VMWare: \"VNC\" or \"VMRC\"",
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username for oVirt, EC2, VMware, OpenStack. Access Key for EC2.",
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "Password for oVirt, EC2, VMware, OpenStack. Secret key for EC2",
			},
			"datacenter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "For oVirt, VMware Datacenter",
			},
			"server": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "For VMware",
			},
			"setconsolepassword": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "For Libvirt and VMware only",
			},
			"cachingenabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "For VMware only",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanComputeResource constructs a ForemanComputeResource reference from a resource data
// reference.  The struct's  members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanComputeResource(d *schema.ResourceData) *api.ForemanComputeResource {
	utils.TraceFunctionCall()

	computeresource := api.ForemanComputeResource{}

	obj := buildForemanObject(d)
	computeresource.ForemanObject = *obj

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("name"); ok {
		computeresource.Name = attr.(string)
	}
	if attr, ok = d.GetOk("url"); ok {
		computeresource.URL = attr.(string)
	}
	if attr, ok = d.GetOk("hypervisor"); ok {
		computeresource.Provider = attr.(string)
	}
	if attr, ok = d.GetOk("displaytype"); ok {
		computeresource.DisplayType = attr.(string)
	}
	if attr, ok = d.GetOk("user"); ok {
		computeresource.User = attr.(string)
	}
	if attr, ok = d.GetOk("password"); ok {
		computeresource.Password = attr.(string)
	}
	if attr, ok = d.GetOk("datacenter"); ok {
		computeresource.Datacenter = attr.(string)
	}
	if attr, ok = d.GetOk("server"); ok {
		computeresource.Server = attr.(string)
	}
	if attr, ok = d.GetOk("setconsolepassword"); ok {
		computeresource.SetConsolePassword = attr.(bool)
	}
	if attr, ok = d.GetOk("cachingenabled"); ok {
		computeresource.CachingEnabled = attr.(bool)
	}

	return &computeresource
}

// setResourceDataFromForemanComputeResource sets a ResourceData's attributes from the
// attributes of the supplied ForemanComputeResource reference
func setResourceDataFromForemanComputeResource(d *schema.ResourceData, fd *api.ForemanComputeResource) {
	utils.TraceFunctionCall()

	d.SetId(strconv.Itoa(fd.Id))
	d.Set("name", fd.Name)
	d.Set("url", fd.URL)
	d.Set("hypervisor", fd.Provider)
	d.Set("displaytype", fd.DisplayType)
	d.Set("user", fd.User)
	d.Set("password", fd.Password)
	d.Set("datacenter", fd.Datacenter)
	d.Set("server", fd.Server)
	d.Set("setconsolepassword", fd.SetConsolePassword)
	d.Set("cachingenabled", fd.CachingEnabled)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanComputeResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()
	return nil
}

func resourceForemanComputeResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	client := meta.(*api.Client)
	computeresource := buildForemanComputeResource(d)

	utils.Debugf("ForemanComputeResource: [%+v]", computeresource)

	readComputeResource, readErr := client.ReadComputeResource(ctx, computeresource.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	utils.Debugf("Read ForemanComputeResource: [%+v]", readComputeResource)

	setResourceDataFromForemanComputeResource(d, readComputeResource)

	return nil
}

func resourceForemanComputeResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()
	return nil
}

func resourceForemanComputeResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	utils.TraceFunctionCall()

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return nil
}

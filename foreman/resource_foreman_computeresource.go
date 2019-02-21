package foreman

import (
	"fmt"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceForemanComputeResource() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanComputeResourceCreate,
		Read:   resourceForemanComputeResourceRead,
		Update: resourceForemanComputeResourceUpdate,
		Delete: resourceForemanComputeResourceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of computeresource. ComputeResources serve as an "+
						"identification string that defines autonomy, authority, or control "+
						"for a portion of a network.",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the compute resource",
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "URL for Libvirt, oVirt, OpenStack and Rackspace",
			},
			"hypervisor": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: "The HyperVisor/Cloud Provider for this Compute Resource:" +
					"supported providers include \"Libvirt\", \"Ovirt\", \"EC2\"," +
					"\"Vmware\", \"Openstack\", \"Rackspace\", \"GCE\"",
			},
			"displaytype": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "For Libvirt: \"VNC\" or \"SPICE\". For VMWare: \"VNC\" or \"VMRC\"",
			},
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username for oVirt, EC2, VMware, OpenStack. Access Key for EC2.",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "Password for oVirt, EC2, VMware, OpenStack. Secret key for EC2",
			},
			"datacenter": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "For oVirt, VMware Datacenter",
			},
			"server": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "For VMware",
			},
			"setconsolepassword": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "For Libvirt and VMware only",
			},
			"cachingenabled": &schema.Schema{
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
	log.Tracef("resource_foreman_computeresource.go#buildForemanComputeResource")

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
	log.Tracef("resource_foreman_computeresource.go#setResourceDataFromForemanComputeResource")

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

func resourceForemanComputeResourceCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_computeresource.go#Create")
	return nil
}

func resourceForemanComputeResourceRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_computeresource.go#Read")

	client := meta.(*api.Client)
	computeresource := buildForemanComputeResource(d)

	log.Debugf("ForemanComputeResource: [%+v]", computeresource)

	readComputeResource, readErr := client.ReadComputeResource(computeresource.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanComputeResource: [%+v]", readComputeResource)

	setResourceDataFromForemanComputeResource(d, readComputeResource)

	return nil
}

func resourceForemanComputeResourceUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_computeresource.go#Update")
	return nil
}

func resourceForemanComputeResourceDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_computeresource.go#Delete")

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return nil
}

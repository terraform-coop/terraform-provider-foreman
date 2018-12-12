package foreman

import (
	"fmt"
	"strconv"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceForemanSubnet() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanSubnetCreate,
		Read:   resourceForemanSubnetRead,
		Update: resourceForemanSubnetUpdate,
		Delete: resourceForemanSubnetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of a subnetwork.",
					autodoc.MetaSummary,
				),
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Subnet name. "+
						"%s \"10.228.247.0 BO1\"",
					autodoc.MetaExample,
				),
			},

			"network": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.SingleIP(),
				Description: fmt.Sprintf(
					"Subnet network. "+
						"%s \"10.228.247.0\"",
					autodoc.MetaExample,
				),
			},

			"mask": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.SingleIP(),
				Description: fmt.Sprintf(
					"Netmask for this subnet. "+
						"%s \"255.255.255.0\"",
					autodoc.MetaExample,
				),
			},

			"gateway": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.SingleIP(),
				Description: "Gateway server to use when connecting/communicating to " +
					"anything not on the same network.",
			},

			"dns_primary": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.SingleIP(),
				Description:  "Primary DNS server for this subnet.",
			},

			"dns_secondary": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.SingleIP(),
				Description:  "Secondary DNS sever for this subnet.",
			},

			"ipam": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"DHCP",
					"Internal DB",
					"None",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "IP address auto-suggestion for this subnet. Valid " +
					"values include: `\"DHCP\"`, `\"Internal DB\"`, `\"None\"`.",
			},

			"from": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.SingleIP(),
				Description:  "Start IP address for IP auto suggestion.",
			},

			"to": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.SingleIP(),
				Description:  "Ending IP address for IP auto suggestion.",
			},

			"boot_mode": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"Static",
					"DHCP",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "Default boot mode for instances assigned to this subnet. " +
					"Values include: `\"Static\"`, `\"DHCP\"`.",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanSubnet constructs a ForemanSubnet reference from a resource data
// reference.  The struct's  members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanSubnet(d *schema.ResourceData) *api.ForemanSubnet {
	log.Tracef("resource_foreman_subnet.go#buildForemanSubnet")

	s := api.ForemanSubnet{}

	obj := buildForemanObject(d)
	s.ForemanObject = *obj

	s.Network = d.Get("network").(string)
	s.Mask = d.Get("mask").(string)

	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("gateway"); ok {
		s.Gateway = attr.(string)
	}
	if attr, ok = d.GetOk("dns_primary"); ok {
		s.DnsPrimary = attr.(string)
	}
	if attr, ok = d.GetOk("dns_secondary"); ok {
		s.DnsSecondary = attr.(string)
	}
	if attr, ok = d.GetOk("ipam"); ok {
		s.Ipam = attr.(string)
	}
	if attr, ok = d.GetOk("from"); ok {
		s.From = attr.(string)
	}
	if attr, ok = d.GetOk("to"); ok {
		s.To = attr.(string)
	}
	if attr, ok = d.GetOk("boot_mode"); ok {
		s.BootMode = attr.(string)
	}

	return &s
}

// setResourceDataFromForemanSubnet sets a ResourceData's attributes from the
// attributes of the supplied ForemanSubnet reference
func setResourceDataFromForemanSubnet(d *schema.ResourceData, fs *api.ForemanSubnet) {
	log.Tracef("resource_foreman_subnet.go#setResourceDataFromForemanSubnet")

	d.SetId(strconv.Itoa(fs.Id))
	d.Set("name", fs.Name)
	d.Set("network", fs.Network)
	d.Set("mask", fs.Mask)
	d.Set("gateway", fs.Gateway)
	d.Set("dns_primary", fs.DnsPrimary)
	d.Set("dns_secondary", fs.DnsSecondary)
	d.Set("ipam", fs.Ipam)
	d.Set("from", fs.From)
	d.Set("to", fs.To)
	d.Set("boot_mode", fs.BootMode)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_subnet.go#Create")
	return nil
}

func resourceForemanSubnetRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_subnet.go#Read")

	client := meta.(*api.Client)
	s := buildForemanSubnet(d)

	log.Debugf("ForemanSubnet: [%+v]", s)

	readSubnet, readErr := client.ReadSubnet(s.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanSubnet: [%+v]", readSubnet)

	setResourceDataFromForemanSubnet(d, readSubnet)

	return nil
}

func resourceForemanSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_subnet.go#Update")
	return nil
}

func resourceForemanSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_subnet.go#Delete")

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors

	return nil
}

package foreman

import (
	"context"
	"fmt"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/conv"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceForemanSubnet() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanSubnetCreate,
		ReadContext:   resourceForemanSubnetRead,
		UpdateContext: resourceForemanSubnetUpdate,
		DeleteContext: resourceForemanSubnetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s Foreman representation of a subnetwork.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: fmt.Sprintf(
					"Subnet name. "+
						"%s \"10.228.247.0 BO1\"",
					autodoc.MetaExample,
				),
			},

			"network": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPAddress,
				Description: fmt.Sprintf(
					"Subnet network. "+
						"%s \"10.228.247.0\"",
					autodoc.MetaExample,
				),
			},

			"mask": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPAddress,
				Description: fmt.Sprintf(
					"Netmask for this subnet. "+
						"%s \"255.255.255.0\"",
					autodoc.MetaExample,
				),
			},

			"gateway": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
				Description: "Gateway server to use when connecting/communicating to " +
					"anything not on the same network.",
			},

			"dns_primary": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "Primary DNS server for this subnet.",
			},

			"dns_secondary": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "Secondary DNS sever for this subnet.",
			},

			"ipam": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"DHCP",
					"Internal DB",
					"Random DB",
					"None",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "IP address auto-suggestion for this subnet. Valid " +
					"values include: `\"DHCP\"`, `\"Internal DB\"`, `\"Random DB\"`,`\"None\"`.",
			},

			"from": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "Start IP address for IP auto suggestion.",
			},

			"to": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "Ending IP address for IP auto suggestion.",
			},

			"boot_mode": {
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
			"network_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Subnets CIDR in the format 169.254.0.0/16",
			},
			"vlanid": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "VLAN id that is in use in the subnet",
			},
			"mtu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "MTU value for the subnet",
			},
			"template_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Template HTTP(S) Proxy ID to use within this subnet",
			},
			"dhcp_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "DHCP Proxy ID to use within this subnet",
			},
			"bmc_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "BMC Proxy ID to use within this subnet",
			},
			"tftp_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "TFTP Proxy ID to use within this subnet",
			},
			"httpboot_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "HTTPBoot Proxy ID to use within this subnet",
			},
			"domain_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional:    true,
				Description: "Domains in which this subnet is part",
			},
			"network_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"IPv4",
					"IPv6",
				}, false),
				Description: "Type or protocol, IPv4 or IPv6, defaults to IPv4.",
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
	if attr, ok = d.GetOk("network_address"); ok {
		s.NetworkAddress = attr.(string)
	}
	if attr, ok = d.GetOk("vlanid"); ok {
		s.VlanID = attr.(int)
	}
	if attr, ok = d.GetOk("mtu"); ok {
		s.Mtu = attr.(int)
	}
	if attr, ok = d.GetOk("template_id"); ok {
		s.TemplateID = attr.(int)
	}
	if attr, ok = d.GetOk("dhcp_id"); ok {
		s.DhcpID = attr.(int)
	}
	bmcId := d.Get("bmc_id").(int)
	if bmcId != 0 {
		s.BmcID = &bmcId
	}
	if attr, ok = d.GetOk("tftp_id"); ok {
		s.TftpID = attr.(int)
	}
	if attr, ok = d.GetOk("httpboot_id"); ok {
		s.HTTPBootID = attr.(int)
	}
	if attr, ok = d.GetOk("domain_ids"); ok {
		attrSet := attr.(*schema.Set)
		s.DomainIDs = conv.InterfaceSliceToIntSlice(attrSet.List())
	}
	if attr, ok = d.GetOk("network_type"); ok {
		s.NetworkType = attr.(string)
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
	d.Set("network_address", fs.NetworkAddress)
	d.Set("vlanid", fs.VlanID)
	d.Set("mtu", fs.Mtu)
	d.Set("template_id", fs.TemplateID)
	d.Set("dhcp_id", fs.DhcpID)
	d.Set("bmc_id", fs.BmcID)
	d.Set("tftp_id", fs.TftpID)
	d.Set("httpboot_id", fs.HTTPBootID)
	d.Set("domain_ids", fs.DomainIDs)
	d.Set("network_type", fs.NetworkType)
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanSubnetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_subnet.go#Create")

	client := meta.(*api.Client)
	s := buildForemanSubnet(d)

	log.Debugf("ForemanSubnet: [%+v]", s)

	createdSubnet, createErr := client.CreateSubnet(ctx, s)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanSubnet: [%+v]", createdSubnet)

	setResourceDataFromForemanSubnet(d, createdSubnet)

	return nil
}

func resourceForemanSubnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_subnet.go#Read")

	client := meta.(*api.Client)
	s := buildForemanSubnet(d)

	log.Debugf("ForemanSubnet: [%+v]", s)

	readSubnet, readErr := client.ReadSubnet(ctx, s.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanSubnet: [%+v]", readSubnet)

	setResourceDataFromForemanSubnet(d, readSubnet)

	return nil
}

func resourceForemanSubnetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_subnet.go#Update")
	client := meta.(*api.Client)
	s := buildForemanSubnet(d)

	log.Debugf("ForemanSubnet: [%+v]", s)

	updatedSubnet, updateErr := client.UpdateSubnet(ctx, s)
	if updateErr != nil {
		return diag.FromErr(updateErr)
	}

	log.Debugf("Updated ForemanSubnet: [%+v]", updatedSubnet)

	setResourceDataFromForemanSubnet(d, updatedSubnet)

	return nil
}

func resourceForemanSubnetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_subnet.go#Delete")

	client := meta.(*api.Client)
	s := buildForemanSubnet(d)

	log.Debugf("ForemanSubnet: [%+v]", s)

	return diag.FromErr(api.CheckDeleted(d, client.DeleteSubnet(ctx, s.Id)))
}

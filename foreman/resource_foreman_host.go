package foreman

import (
	"fmt"
	"strconv"
	"time"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	"github.com/wayfair/terraform-provider-utils/autodoc"
	"github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceForemanHost() *schema.Resource {
	return &schema.Resource{

		Create: resourceForemanHostCreate,
		Read:   resourceForemanHostRead,
		Update: resourceForemanHostUpdate,
		Delete: resourceForemanHostDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s A host managed by Foreman.",
					autodoc.MetaSummary,
				),
			},

			// -- Required --

			"name": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				Description: fmt.Sprintf(
					"Host fully qualified domain name. "+
						"%s \"compute01.dc1.company.com\"",
					autodoc.MetaExample,
				),
			},

			// -- Optional --

			"comment": &schema.Schema{
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description: fmt.Sprintf("Add additional information about this host." +
					"Note: Changes to this attribute will trigger a host rebuild.",
				),
			},

			"enable_bmc": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Enables PMI/BMC functionality. On create and update " +
					"calls, having this enabled will force a host to poweroff, set next " +
					"boot to PXE and power on. Defaults to `false`.",
			},

			"retry_count": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				Description:  "Number of times to retry on a failed attempt to register a new host in foreman.",
				ValidateFunc: validation.IntAtLeast(1),
			},

			"bmc_success": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: fmt.Sprintf(
					"Tracks the partial state of BMC operations on host "+
						"creation. If these operations fail, the host will be created in "+
						"Foreman and this boolean will remain `false`. On the next "+
						"`terraform apply` will trigger the host update to pick back up "+
						"with the BMC operations. "+
						"%s",
					autodoc.MetaUnexported,
				),
			},

			// -- Foreign Key Relationships --

			"domain_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the domain to assign to the host.",
			},

			"environment_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the environment to assign to the host.",
			},

			"hostgroup_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the hostgroup to assign to the host.",
			},

			"interfaces_attributes": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Elem:        resourceForemanInterfacesAttributes(),
				Set:         schema.HashResource(resourceForemanInterfacesAttributes()),
				Description: "Host interface information.",
			},

			"operatingsystem_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the operating system to put on the host.",
			},
		},
	}
}

// resourceForemanInterfacesAttributes is a nested resource that represents a
// valid interfaces attribute.  The "id" of this resource is computed and
// assigned by Foreman at the time of creation.
//
// NOTE(ALL): See comments in ResourceData's "interfaces_attributes"
//   attribute definition above
func resourceForemanInterfacesAttributes() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Unique identifier for the interface.",
			},

			// -- Optional --

			"primary": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Whether or not this is the primary interface.",
			},
			"ip": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.SingleIP(),
				Description:  "IP address associated with the interface.",
			},
			"mac": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "MAC address associated with the interface.",
			},
			"subnet_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the subnet to associate with this interface.",
			},
			"identifier": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Identifier of this interface local to the host.",
			},
			"managed": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Whether or not this interface is managed by Foreman.",
			},
			"provision": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Whether or not this interface is used to provision the host.",
			},
			"virtual": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				Description: "Whether or not this is a virtual interface.",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "Username used for BMC/IPMI functionality.",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Sensitive:   true,
				ForceNew:    true,
				Optional:    true,
				Description: "Associated password used for BMC/IPMI functionality.",
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"interface",
					"bmc",
					"bond",
					"bridge",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "The type of interface. Values include: `\"interface\"`, " +
					"`\"bmc\"`, `\"bond\"`, `\"bridge\"`.",
			},
			// Provider used for BMC/IPMI calls. (Default: IPMI)
			"provider": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"IPMI",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "Provider used for BMC/IMPI functionality. Values include: " +
					"`\"IPMI\"`",
			},
		},
	}
}

// -----------------------------------------------------------------------------
// Conversion Helpers
// -----------------------------------------------------------------------------

// buildForemanHost constructs a ForemanHost struct from a resource data
// reference.  The struct's members are populated from the data populated in
// the resource data.  Missing members will be left to the zero value for that
// member's type.
func buildForemanHost(d *schema.ResourceData) *api.ForemanHost {
	log.Tracef("resource_foreman_host.go#buildForemanHost")

	host := api.ForemanHost{}

	obj := buildForemanObject(d)
	host.ForemanObject = *obj

	var attr interface{}
	var ok bool

	host.Name = d.Get("name").(string)
	host.Comment = d.Get("comment").(string)

	if attr, ok = d.GetOk("domain_id"); ok {
		host.DomainId = attr.(int)
	}
	if attr, ok = d.GetOk("environment_id"); ok {
		host.EnvironmentId = attr.(int)
	}
	if attr, ok = d.GetOk("hostgroup_id"); ok {
		host.HostgroupId = attr.(int)
	}
	if attr, ok = d.GetOk("operatingsystem_id"); ok {
		host.OperatingSystemId = attr.(int)
	}

	host.InterfacesAttributes = buildForemanInterfacesAttributes(d)

	return &host
}

// buildForemanInterfacesAttributes constructs an array of
// ForemanInterfacesAttribute structs from a resource data reference. The
// struct's members are populated with the data populated in the resource data.
// Missing members will be left to the zero value for that member's type.
func buildForemanInterfacesAttributes(d *schema.ResourceData) []api.ForemanInterfacesAttribute {
	log.Tracef("resource_foreman_host.go#buildForemanInterfacesAttributes")

	tempIntAttr := []api.ForemanInterfacesAttribute{}
	var attr interface{}
	var ok bool

	if attr, ok = d.GetOk("interfaces_attributes"); !ok {
		return tempIntAttr
	}

	// type assert the underlying *schema.Set and convert to a list
	attrSet := attr.(*schema.Set)
	attrList := attrSet.List()
	attrListLen := len(attrList)
	tempIntAttr = make([]api.ForemanInterfacesAttribute, attrListLen)
	// iterate over each of the map structure entires in the set and convert that
	// to a concrete struct implementation to append to the interfaces
	// attributes list.
	for idx, attrMap := range attrList {
		tempIntAttrMap := attrMap.(map[string]interface{})
		tempIntAttr[idx] = mapToForemanInterfacesAttribute(tempIntAttrMap)
	}

	return tempIntAttr
}

// mapToForemanInterfacesAttribute converts a map[string]interface{} to a
// ForemanInterfacesAttribute struct.  The supplied map comes from an entry in
// the *schema.Set for the "interfaces_attributes" property of the resource,
// since *schema.Set stores its entries as this map structure.
//
// The map should have the following keys. Omitted or invalid map values will
// result in the struct receiving the zero value for that property.
//
//   id (int)
//   primary (bool)
//   ip (string)
//   mac (string)
//   name (string)
//   subnet_id (int)
//   identifier (string)
//   managed (bool)
//   provision (bool)
//   virtual (bool)
//   username (string)
//   password (string)
//   type (string)
//   provider (string)
//   _destroy (bool)

func mapToForemanInterfacesAttribute(m map[string]interface{}) api.ForemanInterfacesAttribute {
	log.Tracef("mapToForemanInterfacesAttribute")

	tempIntAttr := api.ForemanInterfacesAttribute{}
	var ok bool

	if tempIntAttr.Id, ok = m["id"].(int); !ok {
		tempIntAttr.Id = 0
	}

	if tempIntAttr.Primary, ok = m["primary"].(bool); !ok {
		tempIntAttr.Primary = false
	}

	if tempIntAttr.IP, ok = m["ip"].(string); !ok {
		tempIntAttr.IP = ""
	}

	if tempIntAttr.Name, ok = m["name"].(string); !ok {
		tempIntAttr.Name = ""
	}

	if tempIntAttr.SubnetId, ok = m["subnet_id"].(int); !ok {
		tempIntAttr.SubnetId = 0
	}

	if tempIntAttr.MAC, ok = m["mac"].(string); !ok {
		tempIntAttr.MAC = ""
	}

	if tempIntAttr.Managed, ok = m["managed"].(bool); !ok {
		tempIntAttr.Managed = false
	}

	if tempIntAttr.Provision, ok = m["provision"].(bool); !ok {
		tempIntAttr.Provision = false
	}

	if tempIntAttr.Virtual, ok = m["virtual"].(bool); !ok {
		tempIntAttr.Virtual = false
	}

	if tempIntAttr.Username, ok = m["username"].(string); !ok {
		tempIntAttr.Username = ""
	}

	if tempIntAttr.Password, ok = m["password"].(string); !ok {
		tempIntAttr.Password = ""
	}

	if tempIntAttr.Identifier, ok = m["identifier"].(string); !ok {
		tempIntAttr.Identifier = ""
	}

	if tempIntAttr.Type, ok = m["type"].(string); !ok {
		tempIntAttr.Type = ""
	}

	if tempIntAttr.Provider, ok = m["provider"].(string); !ok {
		tempIntAttr.Provider = ""
	}

	if tempIntAttr.Destroy, ok = m["_destroy"].(bool); !ok {
		tempIntAttr.Destroy = false
	}

	log.Debugf("m: [%v], tempIntAttr: [%+v]", m, tempIntAttr)
	return tempIntAttr
}

// setResourceDataFromForemanHost sets a ResourceData's attributes from the
// attributes of the supplied ForemanHost struct
func setResourceDataFromForemanHost(d *schema.ResourceData, fh *api.ForemanHost) {
	log.Tracef("resource_foreman_host.go#setResourceDataFromForemanHost")

	d.SetId(strconv.Itoa(fh.Id))

	d.Set("name", fh.Name)
	d.Set("comment", fh.Comment)
	d.Set("domain_id", fh.DomainId)
	d.Set("environment_id", fh.EnvironmentId)
	d.Set("hostgroup_id", fh.HostgroupId)
	d.Set("operatingsystem_id", fh.OperatingSystemId)

	// In partial mode, flag keys below as completed successfully
	d.SetPartial("name")
	d.SetPartial("comment")
	d.SetPartial("domain_id")
	d.SetPartial("environment_id")
	d.SetPartial("hostgroup_id")
	d.SetPartial("operatingsystem_id")
	d.SetPartial("enable_bmc")

	setResourceDataFromForemanInterfacesAttributes(d, fh.InterfacesAttributes)
}

// setResourceDataFromInterfacesAttributes sets a ResourceData's
// "interfaces_attributes" attribute to the value of the supplied array of
// ForemanInterfacesAttribute structs
func setResourceDataFromForemanInterfacesAttributes(d *schema.ResourceData, fhia []api.ForemanInterfacesAttribute) {
	// this attribute is a *schema.Set.  In order to construct a set, we need to
	// supply a hash function so the set can differentiate for uniqueness of
	// entries.  The hash function will be based on the resource definition
	hashFunc := schema.HashResource(resourceForemanInterfacesAttributes())
	// underneath, a *schema.Set stores an array of map[string]interface{} entries.
	// convert each ForemanInterfaces struct in the supplied array to a
	// mapstructure and then add it to the set
	ifaceArr := make([]interface{}, len(fhia))
	for idx, val := range fhia {
		// NOTE(ALL): we ommit the "_destroy" property here - this does not need
		//   to be stored by terraform in the state file. That is a hidden key that
		//   is only used in updates.  Anything that exists will always have it
		//   set to "false".
		ifaceMap := map[string]interface{}{
			"id":        val.Id,
			"ip":        val.IP,
			"mac":       val.MAC,
			"name":      val.Name,
			"subnet_id": val.SubnetId,
			"primary":   val.Primary,
			"managed":   val.Managed,
			"provision": val.Provision,
			"virtual":   val.Virtual,
			"type":      val.Type,
			"provider":  val.Provider,
			"username":  val.Username,
			"password":  val.Password,
		}
		ifaceArr[idx] = ifaceMap
	}
	// with the array set up, create the *schema.Set and set the ResourceData's
	// "interfaces_attributes" property
	tempIntAttrSet := schema.NewSet(hashFunc, ifaceArr)
	d.Set("interfaces_attributes", tempIntAttrSet)

	// For partial state, passing a prefix will flag all nested keys as successful
	d.SetPartial("interfaces_attributes")
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanHostCreate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_host.go#Create")

	client := meta.(*api.Client)
	h := buildForemanHost(d)

	// NOTE(ALL): Set the build flag to true on host create
	h.Build = true

	log.Debugf("ForemanHost: [%+v]", h)
	hostRetryCount := d.Get("retry_count").(int)

	createdHost, createErr := client.CreateHost(h, hostRetryCount)
	if createErr != nil {
		return createErr
	}

	log.Debugf("Created ForemanHost: [%+v]", createdHost)

	// Enables partial state mode in the event of failure of one of API calls required for host creation
	// This requires you to call the SetPartial function for each changed key.
	// Only changes enabled with SetPartial are merged in.
	d.Partial(true)

	setResourceDataFromForemanHost(d, createdHost)

	enablebmc := d.Get("enable_bmc").(bool)

	var powerCmds []interface{}
	// If enable_bmc is true, perform required power off, pxe boot and power on BMC functions
	if enablebmc {
		log.Debugf("Calling BMC Reboot/PXE Functions")
		// List of BMC Actions to perform
		powerCmds = []interface{}{
			api.Power{
				PowerAction: api.PowerOff,
			},
			api.BMCBoot{
				Device: api.BootPxe,
			},
			api.Power{
				PowerAction: api.PowerOn,
			},
		}
	} else {
		log.Debugf("Using default Foreman behaviour for startup")
		powerCmds = []interface{}{
			api.Power{
				PowerAction: api.PowerOn,
			},
		}
	}

	// Loop through each of the above BMC Operations and execute.
	// In the event fo any failure, exit with error
	for _, cmd := range powerCmds {
		sendErr := client.SendPowerCommand(createdHost, cmd, hostRetryCount)
		if sendErr != nil {
			return sendErr
		}
		// Sleep for 3 seconds between chained BMC calls
		duration := time.Duration(3) * time.Second
		time.Sleep(duration)
	}
	// When the BMC Operations succeed, set the `bmc_success` key to true.
	d.Set("bmc_success", true)
	// Set the `bmc_success` key as successful in partial mode
	d.SetPartial("bmc_success")

	// Disable partial mode
	d.Partial(false)

	return nil
}

func resourceForemanHostRead(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_host.go#Read")

	client := meta.(*api.Client)
	h := buildForemanHost(d)

	log.Debugf("ForemanHost: [%+v]", h)

	readHost, readErr := client.ReadHost(h.Id)
	if readErr != nil {
		return readErr
	}

	log.Debugf("Read ForemanHost: [%+v]", readHost)

	setResourceDataFromForemanHost(d, readHost)

	return nil
}

func resourceForemanHostUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_host.go#Update")

	client := meta.(*api.Client)
	h := buildForemanHost(d)

	// NOTE(ALL): Set the build flag to true on host create
	h.Build = true

	log.Debugf("ForemanHost: [%+v]", h)

	// Enable partial mode in the event of failure of one of API calls required for host update
	d.Partial(true)

	// NOTE(ALL): Handling the removal of a Interfaces.  See the note
	//   in ForemanInterfacesAttribute's Destroy property
	if d.HasChange("interfaces_attributes") {
		oldVal, newVal := d.GetChange("interfaces_attributes")
		oldValSet, newValSet := oldVal.(*schema.Set), newVal.(*schema.Set)

		// NOTE(ALL): The set difference operation is anticommutative (because math)
		//   ie: [A - B] =/= [B - A].
		//
		//   When performing an update, we need to figure out which interfaces
		//   were removed from the set and tag the destroy property
		//   to true and instruct Foreman which ones to delete from the list. We do
		//   this by performing a set difference between the old set and the new
		//   set (ie: [old - new]) which will return the items that used to be in
		//   the set but are no longer included.
		//
		//   The values that were added to the set or remained unchanged are already
		//   part of the interfaces.  They are present in the
		//   ResourceData and already exist from the
		//   buildForemanHost() call.

		setDiff := oldValSet.Difference(newValSet)
		setDiffList := setDiff.List()
		log.Debugf("setDiffList: [%v]", setDiffList)

		// iterate over the removed items, add them back to the interface's
		// array, but tag them for removal.
		//
		// each of the set's items is stored as a map[string]interface{} - use
		// type assertion and construct the struct
		for _, rmVal := range setDiffList {
			// construct, tag for deletion from list of interfaces
			rmValMap := rmVal.(map[string]interface{})
			rmInterface := mapToForemanInterfacesAttribute(rmValMap)
			rmInterface.Destroy = true
			// append back to interface's list
			h.InterfacesAttributes = append(h.InterfacesAttributes, rmInterface)
		}

	} // end HasChange("interfaces_attributes")

	hostRetryCount := d.Get("retry_count").(int)

	// We need to test whether a call to update the host is necessary based on what has changed.
	// Otherwise, a detected update caused by a unsuccessful BMC operation will cause a 422 on update.
	if d.HasChange("name") ||
		d.HasChange("comment") ||
		d.HasChange("domain_id") ||
		d.HasChange("environment_id") ||
		d.HasChange("hostgroup_id") ||
		d.HasChange("operatingsystem_id") ||
		d.HasChange("interfaces_attributes") {

		log.Debugf("host: [%+v]", h)

		updatedHost, updateErr := client.UpdateHost(h, hostRetryCount)
		if updateErr != nil {
			return updateErr
		}

		log.Debugf("Updated FormanHost: [%+v]", updatedHost)

		setResourceDataFromForemanHost(d, updatedHost)
	} // end HasChange("name")

	// Perform BMC operations on update only if the bmc_success boolean has a change
	if d.HasChange("bmc_success") {
		enablebmc := d.Get("enable_bmc").(bool)

		var powerCmds []interface{}
		// If enable_bmc is true, perform required power off, pxe boot and power on BMC functions
		if enablebmc {
			log.Debugf("Calling BMC Reboot/PXE Functions")
			// List of BMC Actions to perform
			powerCmds = []interface{}{
				api.Power{
					PowerAction: api.PowerOff,
				},
				api.BMCBoot{
					Device: api.BootPxe,
				},
				api.Power{
					PowerAction: api.PowerOn,
				},
			}
		} else {
			powerCmds = []interface{}{
				api.Power{
					PowerAction: api.PowerOn,
				},
			}
		}

		for _, cmd := range powerCmds {
			sendErr := client.SendPowerCommand(h, cmd, hostRetryCount)
			if sendErr != nil {
				return sendErr
			}
			// Sleep for 3 seconds between chained BMC calls
			duration := time.Duration(3) * time.Second
			time.Sleep(duration)
		}
		d.Set("bmc_success", true)
		d.SetPartial("bmc_success")

	} // end HasChange("bmc_success")
	// Use partial state mode in the event of failure of one of API calls required for host creation
	d.Partial(false)

	return nil
}

func resourceForemanHostDelete(d *schema.ResourceData, meta interface{}) error {
	log.Tracef("resource_foreman_host.go#Delete")

	client := meta.(*api.Client)
	h := buildForemanHost(d)

	log.Debugf("ForemanHost: [%+v]", h)
	hostRetryCount := d.Get("retry_count").(int)

	if len(h.InterfacesAttributes) > 0 {
		log.Debugf("deleting host that has interfaces set")
		// iterate through each of the host interfaces and tag them for
		// removal from the list
		for idx := range h.InterfacesAttributes {
			h.InterfacesAttributes[idx].Destroy = true
		}
		log.Debugf("host: [%+v]", h)

		updatedHost, updateErr := client.UpdateHost(h, hostRetryCount)
		if updateErr != nil {
			return updateErr
		}

		log.Debugf("updated Host: [%+v]", updatedHost)

		// NOTE(ALL): set the resource data's properties to what comes back from
		//   the update call. This allows us to recover from a partial state if
		//   delete encounters an error after this point - at least the resource's
		//   state will be saved with the correct interfaces.
		setResourceDataFromForemanHost(d, updatedHost)

		log.Debugf("completed the interface deletion")

	} // end if len(host.InterfacesAttributes) > 0

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	return client.DeleteHost(h.Id)
}

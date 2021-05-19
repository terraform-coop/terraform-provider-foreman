package foreman

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/imdario/mergo"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceForemanHost() *schema.Resource {
	return &schema.Resource{

		Create:        resourceForemanHostCreate,
		Read:          resourceForemanHostRead,
		Update:        resourceForemanHostUpdate,
		Delete:        resourceForemanHostDelete,
		CustomizeDiff: resourceForemanHostCustomizeDiff,

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

			"method": &schema.Schema{
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "build",
				ValidateFunc: validation.StringInSlice([]string{
					"build",
					"image",
				}, false),
				Description: "Chooses a method with which to provision the Host" +
					"Options are \"build\" and \"image\"",
			},

			"comment": &schema.Schema{
				Type:         schema.TypeString,
				ForceNew:     false,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description: fmt.Sprintf("Add additional information about this host." +
					"Note: Changes to this attribute will trigger a host rebuild.",
				),
			},
			"parameters": &schema.Schema{
				Type:     schema.TypeMap,
				ForceNew: false,
				Optional: true,
				Description: "A map of parameters that will be saved as host parameters " +
					"in the machine config.",
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
				Description:  "Number of times to retry on a failed attempt to register or delete a host in foreman.",
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

			"owner_type": &schema.Schema{
				Type:         schema.TypeString,
				ForceNew:     false,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description:  fmt.Sprintf("Owner of the host, must be either User ot Usergroup"),
			},

			// -- Foreign Key Relationships --

			"owner_id": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				//Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the user or usergroup that owns the host.",
			},

			"domain_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the domain to assign to the host.",
			},

			"environment_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the environment to assign to the host.",
			},
			"operatingsystem_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the operating system to put on the host.",
			},
			"medium_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the medium mounted on the host.",
			},
			"hostgroup_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the hostgroup to assign to the host.",
			},
			"image_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of an image to be used as base for this host when cloning",
			},
			"model_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the hardware model if applicable",
			},
			"compute_resource_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"compute_profile_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     false,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"compute_attributes": &schema.Schema{
				Type:             schema.TypeString,
				ValidateFunc:     validation.ValidateJsonString,
				Optional:         true,
				Computed:         true,
				Description:      "Hypervisor specific VM options. Must be a JSON string, as every compute provider has different attributes schema",
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},

			// -- Key Components --
			"interfaces_attributes": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        resourceForemanInterfacesAttributes(),
				Description: "Host interface information.",
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
				Default:     false,
				Description: "Whether or not this is the primary interface.",
			},
			"ip": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.SingleIP(),
				Description:  "IP address associated with the interface.",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Name of the interface",
			},
			"mac": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "MAC address associated with the interface.",
			},
			"subnet_id": &schema.Schema{
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the subnet to associate with this interface.",
			},
			"identifier": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier of this interface local to the host.",
			},
			"managed": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not this interface is managed by Foreman.",
			},
			"provision": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not this interface is used to provision the host.",
			},
			"virtual": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not this is a virtual interface.",
			},
			"attached_to": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier of the interface to which this interface belongs.",
			},
			"attached_devices": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifiers of attached interfaces, e.g. 'eth1', 'eth2' as comma-separated list",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username used for BMC/IPMI functionality.",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "Associated password used for BMC/IPMI functionality.",
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
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
			"bmc_provider": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"IPMI",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "Provider used for BMC/IMPI functionality. Values include: " +
					"`\"IPMI\"`",
			},
			"compute_attributes": &schema.Schema{
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Hypervisor specific interface options",
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
	host.OwnerType = d.Get("owner_type").(string)
	host.Method = d.Get("method").(string)
	host.OwnerId = d.Get("owner_id").(int)

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
	if attr, ok = d.GetOk("medium_id"); ok {
		host.MediumId = attr.(int)
	}
	if attr, ok = d.GetOk("image_id"); ok {
		host.ImageId = attr.(int)
	}
	if attr, ok = d.GetOk("model_id"); ok {
		host.ModelId = attr.(int)
	}
	if attr, ok = d.GetOk("compute_resource_id"); ok {
		host.ComputeResourceId = attr.(int)
	}
	if attr, ok = d.GetOk("compute_profile_id"); ok {
		host.ComputeProfileId = attr.(int)
	}
	if attr, ok = d.GetOk("parameters"); ok {
		hostTags := d.Get("parameters").(map[string]interface{})
		for key, value := range hostTags {
			host.HostParameters = append(host.HostParameters, api.ForemanKVParameter{
				Name:  key,
				Value: value.(string),
			})
		}
	}

	if attr, ok = d.GetOk("compute_attributes"); ok {
		host.ComputeAttributes = expandComputeAttributes(attr)
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
	attrList := attr.([]interface{})
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
//   bmc_provider (string)
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

	if tempIntAttr.Provider, ok = m["bmc_provider"].(string); !ok {
		tempIntAttr.Provider = ""
	}

	if tempIntAttr.AttachedTo, ok = m["attached_to"].(string); !ok {
		tempIntAttr.AttachedTo = ""
	}

	if tempIntAttr.AttachedDevices, ok = m["attached_devices"].(string); !ok {
		tempIntAttr.AttachedDevices = ""
	}

	if tempIntAttr.ComputeAttributes, ok = m["compute_attributes"].(map[string]interface{}); !ok {
		tempIntAttr.ComputeAttributes = nil
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

	host_parameters := make(map[string]string)
	for _, parameter := range fh.HostParameters {
		host_parameters[parameter.Name] = parameter.Value
	}

	d.Set("name", fh.Name)
	d.Set("comment", fh.Comment)
	d.Set("parameters", host_parameters)

	if err := d.Set("compute_attributes", flattenComputeAttributes(fh.ComputeAttributes)); err != nil {
		log.Printf("[WARN] error setting compute attributes: %s", err)
	}

	d.Set("domain_id", fh.DomainId)
	d.Set("environment_id", fh.EnvironmentId)
	d.Set("owner_id", fh.OwnerId)
	d.Set("owner_type", fh.OwnerType)
	d.Set("hostgroup_id", fh.HostgroupId)
	d.Set("compute_resource_id", fh.ComputeResourceId)
	d.Set("compute_profile_id", fh.ComputeProfileId)
	d.Set("operatingsystem_id", fh.OperatingSystemId)
	d.Set("medium_id", fh.MediumId)
	d.Set("image_id", fh.ImageId)
	d.Set("model_id", fh.ModelId)

	// In partial mode, flag keys below as completed successfully
	d.SetPartial("name")
	d.SetPartial("comment")
	d.SetPartial("parameters")
	d.SetPartial("compute_attributes")
	d.SetPartial("domain_id")
	d.SetPartial("environment_id")
	d.SetPartial("owner_id")
	d.SetPartial("owner_type")
	d.SetPartial("hostgroup_id")
	d.SetPartial("compute_resource_id")
	d.SetPartial("compute_profile_id")
	d.SetPartial("operatingsystem_id")
	d.SetPartial("medium_id")
	d.SetPartial("image_id")
	d.SetPartial("model_id")
	d.SetPartial("enable_bmc")

	setResourceDataFromForemanInterfacesAttributes(d, fh)
}

// setResourceDataFromInterfacesAttributes sets a ResourceData's
// "interfaces_attributes" attribute to the value of the supplied array of
// ForemanInterfacesAttribute structs
func setResourceDataFromForemanInterfacesAttributes(d *schema.ResourceData, fh *api.ForemanHost) {
	// this attribute is a *schema.Set.  In order to construct a set, we need to
	// supply a hash function so the set can differentiate for uniqueness of
	// entries.  The hash function will be based on the resource definition
	//hashFunc := schema.HashResource(resourceForemanInterfacesAttributes())
	// underneath, a *schema.Set stores an array of map[string]interface{} entries.
	// convert each ForemanInterfaces struct in the supplied array to a
	// mapstructure and then add it to the set
	fhia := fh.InterfacesAttributes
	interfaces_compute_attributes := make(map[string]interface{})
	var ifs interface{}
	var ok bool

	if ifs, ok = fh.ComputeAttributes.(map[string]interface{})["interfaces_attributes"]; ok {
		for _, attrs := range ifs.(map[string]interface{}) {
			a := attrs.(map[string]interface{})
			interfaces_compute_attributes[a["mac"].(string)] = a["compute_attributes"]
		}
	}

	ifaceArr := make([]interface{}, len(fhia))
	for idx, val := range fhia {
		// NOTE(ALL): we ommit the "_destroy" property here - this does not need
		//   to be stored by terraform in the state file. That is a hidden key that
		//   is only used in updates.  Anything that exists will always have it
		//   set to "false".
		ifaceMap := map[string]interface{}{
			"id":           val.Id,
			"ip":           val.IP,
			"mac":          val.MAC,
			"name":         val.Name,
			"subnet_id":    val.SubnetId,
			"primary":      val.Primary,
			"managed":      val.Managed,
			"identifier":   val.Identifier,
			"provision":    val.Provision,
			"virtual":      val.Virtual,
			"type":         val.Type,
			"bmc_provider": val.Provider,
			"username":     val.Username,
			"password":     val.Password,

			"attached_devices": val.AttachedDevices,
			"attached_to":      val.AttachedTo,
		}

		// NOTE(ALL): These settings only apply to virtual machines
		if ifaceMap["compute_attributes"], ok = interfaces_compute_attributes[val.MAC]; !ok {
			ifaceMap["compute_attributes"] = val.ComputeAttributes
		}

		ifaceArr[idx] = ifaceMap
	}
	// with the array set up, create the *schema.Set and set the ResourceData's
	// "interfaces_attributes" property
	d.Set("interfaces_attributes", ifaceArr)

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
	if h.Method == "build" {
		h.Build = true
	}

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
			api.BMCBoot{
				Device: api.BootPxe,
			},
			api.Power{
				PowerAction: api.PowerCycle,
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

	log.Debugf("ForemanHost: [%+v]", h)

	// Enable partial mode in the event of failure of one of API calls required for host update
	d.Partial(true)

	// NOTE(ALL): Do not make requests to compute provider if no changes to compute attributes are needed
	if !d.HasChange("compute_attributes") {
		h.ComputeAttributes = nil
	}

	// NOTE(ALL): Handling the removal of a Interfaces.  See the note
	//   in ForemanInterfacesAttribute's Destroy property
	if d.HasChange("interfaces_attributes") {
		oldVal, newVal := d.GetChange("interfaces_attributes")
		oldValList, newValList := oldVal.([]interface{}), newVal.([]interface{})

		// iterate over the removed items, add them back to the interface's
		// array, but tag them for removal.
		for idx, rmVal := range oldValList {
			if idx+1 > len(newValList) {
				// construct, tag for deletion from list of interfaces
				rmValMap := rmVal.(map[string]interface{})
				rmInterface := mapToForemanInterfacesAttribute(rmValMap)
				rmInterface.Destroy = true
				// append back to interface's list
				h.InterfacesAttributes = append(h.InterfacesAttributes, rmInterface)
			}
		}

	} // end HasChange("interfaces_attributes")

	hostRetryCount := d.Get("retry_count").(int)

	// We need to test whether a call to update the host is necessary based on what has changed.
	// Otherwise, a detected update caused by a unsuccessful BMC operation will cause a 422 on update.
	if d.HasChange("name") ||
		d.HasChange("comment") ||
		d.HasChange("parameters") ||
		d.HasChange("compute_attributes") ||
		d.HasChange("domain_id") ||
		d.HasChange("environment_id") ||
		d.HasChange("owner_id") ||
		d.HasChange("owner_type") ||
		d.HasChange("hostgroup_id") ||
		d.HasChange("compute_resource_id") ||
		d.HasChange("compute_profile_id") ||
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
	returnDelete := client.DeleteHost(h.Id)
	if returnDelete != nil {
		return returnDelete
	}
	retry := 0
	for retry < hostRetryCount {
		log.Debugf("ForemanHostDelete: Waiting for deletion #[%d]", retry)
		_, deleting := client.ReadHost(h.Id)
		if deleting == nil {
			retry++
			time.Sleep(2 * time.Second)
		} else {
			return nil
		}
	}
	return fmt.Errorf("Failed to delete host in retry_count* 2 seconds")
}

func expandComputeAttributes(v interface{}) interface{} {
	var attrs interface{}
	if v == "" {
		v = "{}"
	}

	if err := json.Unmarshal([]byte(v.(string)), &attrs); err != nil {
		log.Printf("[ERROR] Could not unmarshal compute attributes %s: %v", v.(string), err)
		return nil
	}

	return attrs
}

func flattenComputeAttributes(attrs interface{}) interface{} {
	json, err := json.Marshal(attrs)
	if err != nil {
		log.Printf("[ERROR] Could not marshal compute attributes %s: %v", attrs.(string), err)
		return nil
	}
	return string(json)
}

func resourceForemanHostCustomizeDiff(d *schema.ResourceDiff, m interface{}) error {

	oldVal, newVal := d.GetChange("compute_attributes")

	oldMap := expandComputeAttributes(oldVal).(map[string]interface{})
	newMap := expandComputeAttributes(newVal).(map[string]interface{})

	log.Printf("OLD: [%v]", newMap)
	err := mergo.Merge(&oldMap, newMap, mergo.WithOverride)

	if err != nil {
		log.Printf("[ERROR]: Could not merge defined and existing compute attributes, [%v]", err)
	}

	d.SetNew("compute_attributes", flattenComputeAttributes(oldMap))
	return nil
}

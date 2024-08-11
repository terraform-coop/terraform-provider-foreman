package foreman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/HanseMerkur/terraform-provider-utils/autodoc"
	"github.com/HanseMerkur/terraform-provider-utils/conv"
	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/imdario/mergo"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	DEFAULT_RETRY_COUNT = 2
)

func resourceForemanHostV0() *schema.Resource {
	return &schema.Resource{

		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s A host managed by Foreman.",
					autodoc.MetaSummary,
				),
			},

			// -- Required --

			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				Description: fmt.Sprintf(
					"Host fully qualified domain name. "+
						"%s \"compute01.dc1.company.com\"",
					autodoc.MetaExample,
				),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					domainName := d.Get("domain_name").(string)
					if domainName == "" || !(strings.Contains(new, domainName) || strings.Contains(old, domainName)) {
						return false
					}
					return strings.Replace(old, "."+domainName, "", 1) == strings.Replace(new, "."+domainName, "", 1)
				},
			},

			// -- Optional --

			"method": {
				Type:       schema.TypeString,
				ForceNew:   true,
				Optional:   true,
				Default:    "build",
				Deprecated: "The argument is handled by build instead",
				ValidateFunc: validation.StringInSlice([]string{
					"build",
					"image",
				}, false),
				Description: "REMOVED - use build argument instead to manage build flag of host.",
			},

			"comment": {
				Type:         schema.TypeString,
				ForceNew:     false,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description: fmt.Sprintf("Add additional information about this host." +
					"Note: Changes to this attribute will trigger a host rebuild.",
				),
			},

			"parameters": {
				Type:     schema.TypeMap,
				ForceNew: false,
				Computed: true,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of parameters that will be saved as host parameters " +
					"in the machine config.",
			},

			"enable_bmc": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Enables PMI/BMC functionality. On create and update " +
					"calls, having this enabled will force a host to poweroff, set next " +
					"boot to PXE and power on. Defaults to `false`.",
			},

			"manage_build": {
				Type:       schema.TypeBool,
				Optional:   true,
				Default:    true,
				Deprecated: "The feature was merged into the new key managed",
				Description: "REMOVED, please use the new 'managed' key instead." +
					" Create host only, don't set build status or manage power states",
			},

			"managed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "Whether or not this host is managed by Foreman." +
					" Create host only, don't set build status or manage power states.",
			},
			"build": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "Whether or not this host's build flag will be enabled in Foreman. Default is true, " +
					"which means host will be built at next boot.",
			},
			"manage_power_operations": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Manage power operations, e.g. power on, if host's build flag will be enabled.",
			},
			"retry_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      DEFAULT_RETRY_COUNT,
				Description:  "Number of times to retry on a failed attempt to register or delete a host in foreman.",
				ValidateFunc: validation.IntAtLeast(1),
			},

			"bmc_success": {
				Type:       schema.TypeBool,
				Optional:   true,
				Default:    true,
				Deprecated: "The feature no longer exists",
				Description: fmt.Sprintf(
					"REMOVED - Tracks the partial state of BMC operations on host "+
						"creation. If these operations fail, the host will be created in "+
						"Foreman and this boolean will remain `false`. On the next "+
						"`terraform apply` will trigger the host update to pick back up "+
						"with the BMC operations. "+
						"%s",
					autodoc.MetaUnexported,
				),
			},

			"owner_type": {
				Type:         schema.TypeString,
				ForceNew:     false,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description:  "Owner of the host, must be either User ot Usergroup",
			},

			// -- Foreign Key Relationships --

			"owner_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     false,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the user or usergroup that owns the host.",
			},

			"domain_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the domain to assign to the host.",
			},

			"domain_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The domain name of the host.",
			},

			"environment_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the environment to assign to the host.",
			},
			"operatingsystem_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the operating system to put on the host.",
			},
			"medium_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the medium mounted on the host.",
			},
			"hostgroup_id": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the hostgroup to assign to the host.",
			},
			"image_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of an image to be used as base for this host when cloning",
			},
			"model_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the hardware model if applicable",
			},
			"puppet_class_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the applied puppet classes.",
			},
			"compute_resource_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"compute_profile_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     false,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"compute_attributes": {
				Type:             schema.TypeString,
				ValidateFunc:     validation.StringIsJSON,
				Optional:         true,
				Computed:         true,
				Description:      "Hypervisor specific VM options. Must be a JSON string, as every compute provider has different attributes schema",
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},

			// -- Key Components --
			"interfaces_attributes": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem:        resourceForemanInterfacesAttributes(),
				Description: "Host interface information.",
			},
		},
	}
}

func resourceForemanHostStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	rawState["build"] = rawState["method"] == "build"
	rawState["managed"] = rawState["manage_build"]

	return rawState, nil
}

func resourceForemanHost() *schema.Resource {
	return &schema.Resource{

		CreateContext: resourceForemanHostCreate,
		ReadContext:   resourceForemanHostRead,
		UpdateContext: resourceForemanHostUpdate,
		DeleteContext: resourceForemanHostDelete,

		CustomizeDiff: customdiff.All(
			resourceForemanHostCustomizeDiffComputeAttributes,
		),

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceForemanHostV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceForemanHostStateUpgradeV0,
				Version: 0,
			},
		},
		Schema: map[string]*schema.Schema{

			autodoc.MetaAttribute: {
				Type:     schema.TypeBool,
				Computed: true,
				Description: fmt.Sprintf(
					"%s A host managed by Foreman.",
					autodoc.MetaSummary,
				),
			},

			"name": {
				Type:             schema.TypeString,
				Computed:         true,
				Optional:         true,
				ForceNew:         true,
				Required:         false,
				Description:      "Name of this host as stored in Foreman. Can be short name or FQDN, depending on your Foreman settings (especially the setting 'append_domain_name_for_hosts').",
				DiffSuppressFunc: resourceForemanHostNameDiffSuppressFunc,
			},

			"shortname": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Computed:    true,
				Optional:    true,
				Required:    false,
				Description: "The short name of this host. Example: when the FQDN is 'host01.example.org', then 'host01' is the short name.",
				ValidateDiagFunc: func(value interface{}, p cty.Path) diag.Diagnostics {
					var diags diag.Diagnostics
					if strings.Count(value.(string), ".") > 0 {
						diag := diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "Shortname is not allowed to contain dots",
							Detail:   fmt.Sprintf("The shortname %q contains dots, but this is not allowed, since the shortname is not the FQDN.", value),
						}
						diags = append(diags, diag)
					}
					return diags
				},
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					name := d.Get("name").(string)

					// Use case: manifest has no "shortname" defined, but uses "name". Then an apply fetches the shortname,
					// but a new apply would think it should be re-set to "".
					if name != "" && newValue == "" && strings.HasPrefix(name, oldValue) {
						return true
					}
					return false
				},
			},

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true, // Read-only value
				Optional: false,
				Required: false,
				Description: fmt.Sprintf(
					"Host fully qualified domain name. Read-only value to be used in variables. %s \"compute01.dc1.company.com\"",
					autodoc.MetaExample,
				),
			},

			// -- Optional --

			"root_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 256),
				Description:  "Default root password",
			},

			"provision_method": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "build",
				ValidateFunc: validation.StringInSlice([]string{
					"build", // build = Network Based
					"image", // image = Image Based
				}, false),
				Description: "Sets the provision method in Foreman for this host: either network-based ('build') or image-based ('image')",
			},

			"comment": {
				Type:         schema.TypeString,
				ForceNew:     false,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description: fmt.Sprintf("Add additional information about this host." +
					"Note: Changes to this attribute will trigger a host rebuild.",
				),
			},

			"parameters": {
				Type:     schema.TypeMap,
				ForceNew: false,
				Computed: true,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A map of parameters that will be saved as host parameters " +
					"in the machine config.",
			},

			"enable_bmc": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Enables PMI/BMC functionality. On create and update " +
					"calls, having this enabled will force a host to poweroff, set next " +
					"boot to PXE and power on. Defaults to `false`.",
			},

			"managed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "Whether or not this host is managed by Foreman." +
					" Create host only, don't set build status or manage power states.",
			},

			"set_build_flag": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Sets the Foreman-internal 'build' flag on this host - even if it is already built completely.",
			},

			"manage_power_operations": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Manage power operations, e.g. power on, if host's build flag will be enabled.",
			},

			"retry_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      2,
				Description:  "Number of times to retry on a failed attempt to register or delete a host in foreman.",
				ValidateFunc: validation.IntAtLeast(1),
			},

			"bmc_success": {
				Type:       schema.TypeBool,
				Optional:   true,
				Deprecated: "The feature no longer exists",
				Description: fmt.Sprintf(
					"REMOVED - Tracks the partial state of BMC operations on host "+
						"creation. If these operations fail, the host will be created in "+
						"Foreman and this boolean will remain `false`. On the next "+
						"`terraform apply` will trigger the host update to pick back up "+
						"with the BMC operations. "+
						"%s",
					autodoc.MetaUnexported,
				),
			},

			"owner_type": {
				Type:         schema.TypeString,
				ForceNew:     false,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
				Description:  "Owner of the host, must be either User ot Usergroup",
			},

			// -- Foreign Key Relationships --

			"owner_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     false,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the user or usergroup that owns the host.",
			},

			"domain_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the domain to assign to the host.",
			},

			"domain_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The domain name of the host.",
			},

			"environment_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the environment to assign to the host.",
			},

			"architecture_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the architecture of this host",
			},

			"subnet_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the subnet the host should be placed in",
			},

			"ptable_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the partition table the host should use",
			},

			"operatingsystem_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the operating system to put on the host.",
			},
			"medium_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the medium mounted on the host.",
			},
			"hostgroup_id": {
				Type:         schema.TypeInt,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the hostgroup to assign to the host.",
			},
			"image_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of an image to be used as base for this host when cloning",
			},
			"model_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the hardware model if applicable",
			},
			"puppet_class_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the applied puppet classes.",
			},
			"config_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Description: "IDs of the applied config groups.",
			},
			"compute_resource_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"compute_profile_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     false,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"compute_attributes": {
				Type:             schema.TypeString,
				ValidateFunc:     validation.StringIsJSON,
				Optional:         true,
				Computed:         true,
				Description:      "Hypervisor specific VM options. Must be a JSON string, as every compute provider has different attributes schema",
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},

			"token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Build token. Can be used to signal to Foreman that a host build is complete.",
			},

			// -- Key Components --
			"interfaces_attributes": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
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
//
//	attribute definition above
func resourceForemanInterfacesAttributes() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Unique identifier for the interface.",
			},

			// -- Optional --

			"primary": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not this is the primary interface.",
			},
			"ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPAddress,
				Description:  "IP address associated with the interface.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Name of the interface",
			},
			"mac": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "MAC address associated with the interface.",
			},
			"subnet_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "ID of the subnet to associate with this interface.",
			},
			"identifier": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier of this interface local to the host.",
			},
			"managed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not this interface is managed by Foreman.",
			},
			"provision": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not this interface is used to provision the host.",
			},
			"virtual": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether or not this is a virtual interface.",
			},
			"attached_to": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier of the interface to which this interface belongs.",
			},
			"attached_devices": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifiers of attached interfaces, e.g. 'eth1', 'eth2' as comma-separated list",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username used for BMC/IPMI functionality.",
			},
			"password": {
				Type:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "Associated password used for BMC/IPMI functionality.",
			},
			"type": {
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
			"bmc_provider": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"IPMI",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "Provider used for BMC/IMPI functionality. Values include: " +
					"`\"IPMI\"`",
			},
			"compute_attributes": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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

	host.Shortname = d.Get("shortname").(string) // Required
	host.DomainName = d.Get("domain_name").(string)

	// Set name for Foreman
	if host.Name == "" {
		if host.DomainName != "" {
			// Construct full FQDN
			log.Infof("Host %s name was set to FQDN", host.Shortname)
			host.Name = host.Shortname + "." + host.DomainName
		} else {
			// Fall back to short name, omitting the domain part
			log.Infof("Host %s name was set to shortname, because domainname was missing", host.Shortname)
			host.Name = host.Shortname
		}
	}

	host.Comment = d.Get("comment").(string)
	host.OwnerType = d.Get("owner_type").(string)
	host.ProvisionMethod = d.Get("provision_method").(string)
	host.Managed = d.Get("managed").(bool)
	host.Build = d.Get("set_build_flag").(bool)
	host.Token = d.Get("token").(string)

	ownerId := d.Get("owner_id").(int)
	if ownerId != 0 {
		host.OwnerId = &ownerId
	}
	domainId := d.Get("domain_id").(int)
	if domainId != 0 {
		host.DomainId = &domainId
	}
	environmentId := d.Get("environment_id").(int)
	if environmentId != 0 {
		host.EnvironmentId = &environmentId
	}
	hostgroupId := d.Get("hostgroup_id").(int)
	if hostgroupId != 0 {
		host.HostgroupId = &hostgroupId
	}
	architectureId := d.Get("architecture_id").(int)
	if architectureId != 0 {
		host.ArchitectureId = &architectureId
	}
	subnetId := d.Get("subnet_id").(int)
	if subnetId != 0 {
		host.SubnetId = &subnetId
	}
	ptableId := d.Get("ptable_id").(int)
	if ptableId != 0 {
		host.PtableId = &ptableId
	}
	operatingSystemId := d.Get("operatingsystem_id").(int)
	if operatingSystemId != 0 {
		host.OperatingSystemId = &operatingSystemId
	}
	mediumId := d.Get("medium_id").(int)
	if mediumId != 0 {
		host.MediumId = &mediumId
	}

	// TODO: How is this parameter used?
	// VMware-backed providers need the UUID instead of the Foreman-internal ID
	imageId := d.Get("image_id").(int)
	if imageId != 0 {
		host.ImageId = &imageId
	}

	modelId := d.Get("model_id").(int)
	if modelId != 0 {
		host.ModelId = &modelId
	}
	computeResourceId := d.Get("compute_resource_id").(int)
	if computeResourceId != 0 {
		host.ComputeResourceId = &computeResourceId
	}
	computeProfileId := d.Get("compute_profile_id").(int)
	if computeProfileId != 0 {
		host.ComputeProfileId = &computeProfileId
	}
	computeAttributes := expandComputeAttributes(d.Get("compute_attributes").(string))
	if len(computeAttributes) > 0 {
		host.ComputeAttributes = computeAttributes
	}

	if attr, ok = d.GetOk("puppet_class_ids"); ok {
		attrSet := attr.(*schema.Set)
		host.PuppetClassIds = conv.InterfaceSliceToIntSlice(attrSet.List())
		host.PuppetAttributes.Puppetclass_ids = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	if attr, ok = d.GetOk("config_group_ids"); ok {
		attrSet := attr.(*schema.Set)
		host.ConfigGroupIds = conv.InterfaceSliceToIntSlice(attrSet.List())
		host.PuppetAttributes.ConfigGroup_ids = conv.InterfaceSliceToIntSlice(attrSet.List())
	}

	if attr, ok = d.GetOk("parameters"); ok {
		host.HostParameters = api.ToKV(attr.(map[string]interface{}))
	}

	if attr, ok = d.GetOk("root_password"); ok {
		host.RootPassword = attr.(string)
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
func setResourceDataFromForemanHost(d *schema.ResourceData, fh *api.ForemanHost) error {
	log.Tracef("resource_foreman_host.go#setResourceDataFromForemanHost")

	d.SetId(strconv.Itoa(fh.Id))

	log.Debugf("ForemanHost: %+v", fh)

	d.Set("name", fh.Name) // internal name from Foreman meta struct

	// Foreman has a setting called "append_domain_name_for_hosts" which might result in the
	// "name" field being only the shortname. To present consistent values to Terraform, the
	// attributes "fqdn" and "shortname" were introduced.

	// To ensure consistency in the fqdn attribute, handle adding the domain part if needed.
	// This attribute should be used instead of "name".
	if fh.DomainName != "" && !strings.Contains(fh.Name, fh.DomainName) {
		d.Set("fqdn", fmt.Sprintf("%s.%s", fh.Name, fh.DomainName))
	} else {
		d.Set("fqdn", fh.Name)
	}

	// The shortname is created in foreman/api/host.go#constructShortname
	d.Set("shortname", fh.Shortname)

	d.Set("comment", fh.Comment)
	d.Set("parameters", api.FromKV(fh.HostParameters))

	if err := d.Set("compute_attributes", flattenComputeAttributes(fh.ComputeAttributes)); err != nil {
		log.Printf("[WARN] error setting compute attributes: %s", err)
	}

	// See issue #115 for "Build" attribute
	d.Set("managed", fh.Managed)
	d.Set("provision_method", fh.ProvisionMethod)

	d.Set("domain_id", fh.DomainId)
	d.Set("domain_name", fh.DomainName)
	d.Set("environment_id", fh.EnvironmentId)
	d.Set("owner_id", fh.OwnerId)
	d.Set("owner_type", fh.OwnerType)
	d.Set("hostgroup_id", fh.HostgroupId)
	d.Set("architecture_id", fh.ArchitectureId)
	d.Set("ptable_id", fh.PtableId)
	d.Set("subnet_id", fh.SubnetId)
	d.Set("compute_resource_id", fh.ComputeResourceId)
	d.Set("compute_profile_id", fh.ComputeProfileId)
	d.Set("operatingsystem_id", fh.OperatingSystemId)
	d.Set("medium_id", fh.MediumId)
	d.Set("image_id", fh.ImageId)
	d.Set("model_id", fh.ModelId)
	d.Set("puppet_class_ids", fh.PuppetClassIds)
	d.Set("config_group_ids", fh.ConfigGroupIds)
	d.Set("token", fh.Token)

	return setResourceDataFromForemanInterfacesAttributes(d, fh)
}

// setResourceDataFromInterfacesAttributes sets a ResourceData's
// "interfaces_attributes" attribute to the value of the supplied array of
// ForemanInterfacesAttribute structs
func setResourceDataFromForemanInterfacesAttributes(d *schema.ResourceData, fh *api.ForemanHost) error {
	log.Tracef("resource_foreman_host.go#setResourceDataFromForemanInterfacesAttributes")

	// this attribute is a *schema.Set.  In order to construct a set, we need to
	// supply a hash function so the set can differentiate for uniqueness of
	// entries.  The hash function will be based on the resource definition
	// hashFunc := schema.HashResource(resourceForemanInterfacesAttributes())
	// underneath, a *schema.Set stores an array of map[string]interface{} entries.
	// convert each ForemanInterfaces struct in the supplied array to a
	// mapstructure and then add it to the set
	fhia := fh.InterfacesAttributes
	interfaces_compute_attributes := make(map[string]interface{})

	if fh.ComputeAttributes != nil {
		var ifs interface{}
		var ok bool
		if ifs, ok = fh.ComputeAttributes["interfaces_attributes"]; ok {
			ifs_, ok := ifs.(map[string]interface{})
			if !ok {
				return errors.New(
					fmt.Sprintf("could not convert interface attributes to map[string]interface{}. Host: %s."+
						"Original value %+v is of type %T", fh.Name, ifs, ifs),
				)
			}

			for _, attrs := range ifs_ {
				attrs_, ok := attrs.(map[string]interface{})
				if !ok {
					return errors.New(fmt.Sprintf("could not convert attribute to map[string]interface{}. "+
						"Original value %+v is of type %T", attrs, attrs))
				}

				// Error handling to catch https://github.com/terraform-coop/terraform-provider-foreman/issues/160.
				// Check if "mac" exists as key
				if _, ok := attrs_["mac"]; ok {
					// If yes, be extra careful with conversion errors
					mac, ok := attrs_["mac"].(string)
					if !ok {
						return errors.New(fmt.Sprintf("could not convert attribute 'mac' to map[string]interface{}. "))
					}
					interfaces_compute_attributes[mac] = attrs_["compute_attributes"]
				}
			}
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
		var ok bool
		if ifaceMap["compute_attributes"], ok = interfaces_compute_attributes[val.MAC]; !ok {
			ifaceMap["compute_attributes"] = val.ComputeAttributes
		}

		ifaceArr[idx] = ifaceMap
	}
	// with the array set up, create the *schema.Set and set the ResourceData's
	// "interfaces_attributes" property
	d.Set("interfaces_attributes", ifaceArr)

	return nil
}

// -----------------------------------------------------------------------------
// Resource CRUD Operations
// -----------------------------------------------------------------------------

func resourceForemanHostCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_host.go#resourceForemanHostCreate")
	var diags diag.Diagnostics

	client := meta.(*api.Client)
	h := buildForemanHost(d)

	// NOTE(ALL): Set the build flag to true on host create
	if h.ProvisionMethod == "build" && h.Managed {
		h.Build = true
	}
	// Another way to set this flag is to use the "set_build_flag" argument in Terraform

	// Here, commit 7dad08886079b82672eee33f9e1247c5ca60bb77 used a query against the settings API to check
	// the "append_domain_name_for_hosts" setting. In case of true, a shortname will be expanded to
	// a FQDN, resulting in inconsistent plans. Maybe this issue will arise again, then handle it here.

	log.Debugf("ForemanHost: [%+v]", h)
	hostRetryCount := d.Get("retry_count").(int)

	// See commit ad2b5890f09645513b520f12291546f26b812c96 for an experimental implementation
	// for checks of the "computeAttributes" field, when using ProvisionMethod=image.
	// The feature was removed because it was VMware-specific and the test on the backend provider
	// could not yet be implemented (via client.ReadComputeResource -> computeResource.Provider)

	createdHost, createErr := client.CreateHost(ctx, h, hostRetryCount)
	if createErr != nil {
		return diag.FromErr(createErr)
	}

	log.Debugf("Created ForemanHost: [%+v]", createdHost)

	// Enables partial state mode in the event of failure of one of API calls required for host creation
	// This requires you to call the SetPartial function for each changed key.
	// Only changes enabled with SetPartial are merged in.
	d.Partial(true)

	err := setResourceDataFromForemanHost(d, createdHost)
	if err != nil {
		return diag.FromErr(err)
	}

	ManagePowerOperations := d.Get("manage_power_operations").(bool)

	// Manage power operations only if needed, default is true
	if ManagePowerOperations {
		var powerCmds []interface{}
		// If enable_bmc is true, perform required power off, pxe boot and power on BMC functions
		// Don't modify power state at all if we're not managing the build
		if h.EnableBMC {
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
		} else if h.Managed {
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
			sendErr := client.SendPowerCommand(ctx, createdHost, cmd, hostRetryCount)
			if sendErr != nil {
				return diag.FromErr(sendErr)
			}
			// Sleep for 3 seconds between chained BMC calls
			duration := time.Duration(3) * time.Second
			time.Sleep(duration)
		}
	}

	// Disable partial mode
	d.Partial(false)

	return diags
}

func resourceForemanHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_host.go#resourceForemanHostRead")

	client := meta.(*api.Client)
	h := buildForemanHost(d)

	log.Debugf("ForemanHost: [%+v]", h)

	readHost, readErr := client.ReadHost(ctx, h.Id)
	if readErr != nil {
		return diag.FromErr(api.CheckDeleted(d, readErr))
	}

	log.Debugf("Read ForemanHost: [%+v]", readHost)

	err := setResourceDataFromForemanHost(d, readHost)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("retry_count").(int) == 0 {
		d.Set("retry_count", DEFAULT_RETRY_COUNT)
	}

	return nil
}

func resourceForemanHostUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_host.go#resourceForemanHostUpdate")

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
	// Otherwise, a detected update caused by an unsuccessful BMC operation will cause a 422 on update.
	if d.HasChange("name") ||
		d.HasChange("shortname") ||
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
		d.HasChange("interfaces_attributes") ||
		d.HasChange("build") ||
		d.HasChange("puppet_class_ids") ||
		d.HasChange("config_group_ids") ||
		d.Get("managed") == false {

		log.Debugf("host: [%+v]", h)

		updatedHost, updateErr := client.UpdateHost(ctx, h, hostRetryCount)
		if updateErr != nil {
			return diag.FromErr(updateErr)
		}

		log.Debugf("Updated FormanHost: [%+v]", updatedHost)

		err := setResourceDataFromForemanHost(d, updatedHost)
		if err != nil {
			return diag.FromErr(err)
		}
	} // end HasChange("name")

	// Use partial state mode in the event of failure of one of API calls required for host creation
	d.Partial(false)

	return nil
}

func resourceForemanHostDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Tracef("resource_foreman_host.go#Delete")

	client := meta.(*api.Client)
	h := buildForemanHost(d)

	log.Debugf("ForemanHost: [%+v]", h)
	hostRetryCount := d.Get("retry_count").(int)

	// NOTE(ALL): d.SetId("") is automatically called by terraform assuming delete
	//   returns no errors
	returnDelete := client.DeleteHost(ctx, h.Id)
	if returnDelete != nil {
		return diag.FromErr(api.CheckDeleted(d, returnDelete))
	}
	retry := 0
	for retry < hostRetryCount {
		log.Debugf("ForemanHostDelete: Waiting for deletion #[%d]", retry)
		_, deleting := client.ReadHost(ctx, h.Id)
		if deleting == nil {
			retry++
			time.Sleep(2 * time.Second)
		} else {
			return nil
		}
	}
	return diag.Errorf("Failed to delete host in retry_count* 2 seconds")
}

func expandComputeAttributes(v string) map[string]interface{} {
	var attrs map[string]interface{}

	// If Foreman fails to connect to compute provider, it might just return null
	if v == "" || v == "null" {
		v = "{}"
	}

	if err := json.Unmarshal([]byte(v), &attrs); err != nil {
		log.Printf("[ERROR] Could not unmarshal compute attributes %s: %v", v, err)
		return nil
	}

	return attrs
}

func flattenComputeAttributes(attrs map[string]interface{}) string {
	if len(attrs) == 0 {
		return ""
	}
	json, err := json.Marshal(attrs)
	if err != nil {
		log.Printf("[ERROR] Could not marshal compute attributes %v: %v", attrs, err)
		return ""
	}
	return string(json)
}

func resourceForemanHostCustomizeDiffComputeAttributes(ctx context.Context, d *schema.ResourceDiff, i interface{}) error {
	oldVal, newVal := d.GetChange("compute_attributes")

	oldMap := expandComputeAttributes(oldVal.(string))
	newMap := expandComputeAttributes(newVal.(string))

	err := mergo.Merge(&oldMap, newMap, mergo.WithOverride)

	if err != nil {
		log.Printf("[ERROR]: Could not merge defined and existing compute attributes, [%v]", err)
	}

	d.SetNew("compute_attributes", flattenComputeAttributes(oldMap))
	return nil
}

func resourceForemanHostNameDiffSuppressFunc(k, oldValue, newValue string, d *schema.ResourceData) bool {
	domainName := d.Get("domain_name").(string)
	if domainName == "" {
		// If domainName not given, no comparison is possible
		return false
	}

	beforeOld, afterOld, foundOld := strings.Cut(oldValue, ".")
	beforeNew, afterNew, foundNew := strings.Cut(newValue, ".")

	if !foundOld && foundNew {
		if oldValue == beforeNew {
			// Shortname was expanded to FQDN
			return true
		}

		if afterNew == domainName {
			return true
		}
	}

	if (!foundOld && !foundNew) && (beforeOld != beforeNew) {
		// Neither value has a dot but values differ, so no suppression
		return false
	}

	if (foundOld && foundNew) && (afterOld != afterNew) {
		// Domains differ
		return false
	}

	// Use case: We pass in a shortname, Foreman expands it and the next "terraform apply"
	// thinks the shortname is a new value.
	if (foundOld && !foundNew) && strings.HasPrefix(oldValue, newValue) {
		return true
	}

	return false
}

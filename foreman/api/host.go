package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HanseMerkur/terraform-provider-utils/log"
)

const (
	// HostEndpointPrefix : Prefix appended to API url for hosts
	HostEndpointPrefix = "hosts"
	// PowerSuffix : Suffix appended to API url for power operations
	PowerSuffix = "power"
	// ComputeAttributesSuffix : Suffix appended to API url for getting the VM attributes
	ComputeAttributesSuffix = "vm_compute_attributes"
	// PowerOn : Power on operation
	PowerOn = "on"
	// PowerOff : Power off operation
	PowerOff = "off"
	// PowerSoft : Power reboot operation (soft)
	PowerSoft = "soft"
	// PowerCycle : Power reset operation (hard)
	PowerCycle = "cycle"
	// PowerState : Power state check operation
	PowerState = "state"
	// BootSuffix : Suffix appended to API url for power operations
	BootSuffix = "boot"
	// BootDisk : Boot to Disk
	BootDisk = "disk"
	// BootCdrom : Boot to CDROM
	BootCdrom = "cdrom"
	// BootPxe : Boot to PXE
	BootPxe = "pxe"
	// PowerBios : Boot to BIOS
	PowerBios = "bios"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanHost API model represents a host managed by Foreman
type ForemanHost struct {
	// Inherits the base object's attributes
	ForemanObject

	// Whether or not to rebuild the host on reboot
	Build bool `json:"build"`
	// Describes the way this host will be provisioned by Foreman
	Method string `json:"provision_method,omitempty"`
	// ID of the domain to assign the host
	DomainId *int `json:"domain_id,omitempty"`
	// Name of the Domain. To substract from the Machine name
	DomainName string `json:"domain_name,omitempty"`
	// ID of the owner user or group to assign the host
	OwnerId *int `json:"owner_id,omitempty"`
	// Type of the owner, either user or group
	OwnerType string `json:"owner_type,omitempty"`
	// ID of the environment to assign the host
	EnvironmentId *int `json:"environment_id,omitempty"`
	// ID of the hostgroup to assign the host
	HostgroupId *int `json:"hostgroup_id,omitempty"`
	// ID of the operating system to put on the host
	OperatingSystemId *int `json:"operatingsystem_id,omitempty"`
	// ID of the medium that should be mounted
	MediumId *int `json:"medium_id,omitempty"`
	// ID of the image that should be cloned for this host
	ImageId *int `json:"image_id,omitempty"`
	// ID of the hardware model
	ModelId *int `json:"model_id,omitempty"`
	// Whether or not to Enable BMC Functionality on this host
	EnableBMC bool `json:"-"`
	// Boolean to track success of BMC Calls
	BMCSuccess bool `json:"-"`
	// Whether or not the host is managed by foreman
	Managed bool `json:"managed"`
	// Additional information about this host
	Comment string `json:"comment"`
	// Nested struct defining any interfaces associated with the Host
	InterfacesAttributes []ForemanInterfacesAttribute `json:"interfaces_attributes,omitempty"`
	// Map of HostParameters
	HostParameters []ForemanKVParameter `json:"host_parameters_attributes,omitempty"`
	// NOTE(ALL): These settings only apply to virtual machines
	// Hypervisor specific map of ComputeAttributes
	ComputeAttributes map[string]interface{} `json:"compute_attributes,omitempty"`
	// ComputeResourceId specifies the Hypervisor to deploy on
	ComputeResourceId *int `json:"compute_resource_id,omitempty"`
	// ComputeProfileId specifies the Attributes via the Profile Id on the Hypervisor
	ComputeProfileId *int `json:"compute_profile_id,omitempty"`
	// IDs of the puppet classes applied to the host
	PuppetClassIds []int `json:"puppet_class_ids,omitempty"`
	// Build token
	Token string `json:"token,omitempty"`
	// List of config groups to apply to the hostg
	ConfigGroupIds []int `json:"config_group_ids"`
	// The puppetattributes object is only used for create and update, it's not populated on read, hence the duplication
	PuppetAttributes PuppetAttribute `json:"puppet_attributes"`
}

// ForemanInterfacesAttribute representing a hosts defined network interfaces
type ForemanInterfacesAttribute struct {
	Id         int    `json:"id,omitempty"`
	SubnetId   int    `json:"subnet_id,omitempty"`
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	Managed    bool   `json:"managed"`
	Provision  bool   `json:"provision"`
	Virtual    bool   `json:"virtual"`
	Primary    bool   `json:"primary"`
	IP         string `json:"ip"`
	MAC        string `json:"mac"`
	Type       string `json:"type"`
	Provider   string `json:"provider"`

	AttachedDevices string `json:"attached_devices,omitempty"`
	AttachedTo      string `json:"attached_to,omitempty"`

	// NOTE(ALL): These settings only apply to virtual machines
	// ComputeAttributes are hypervisor specific features
	ComputeAttributes map[string]interface{} `json:"compute_attributes,omitempty"`

	// NOTE(ALL): Each of the interfaces receives a unique identifier
	//   on creation. To modify the list of interfaces, the supplied
	//   list to the API does NOT perform a replace operation. Adding new
	//   interfaces to the list is rather trivial and just involves sending the
	//   new values to receive an ID.  When removing one of the combinations from
	//   the set, a secret flag "_destroy" must be supplied as part of that
	//   combination.  This is not documented as part of the Foreman API.  We
	//   omit empty here, because we only want to pass the flag when "_destroy"
	//   is "true" to signal an item removal.
	Destroy bool `json:"_destroy,omitempty"`
}

// foremanHostDecode struct used for JSON decode.
type foremanHostDecode struct {
	ForemanHost
	InterfacesAttributesDecode []ForemanInterfacesAttribute `json:"interfaces"`
	PuppetClassesDecode        []ForemanObject              `json:"puppetclasses"`
	ConfigGroupsDecode         []ForemanObject              `json:"config_groups"`
	HostParametersDecode       []ForemanKVParameter         `json:"parameters"`
}

// Power struct for marshal/unmarshal of power state
// valid states are on, off, soft, cycle, state
// `omitempty` lets use the same struct for power operations.Command
type Power struct {
	PowerAction string `json:"power_action,omitempty"`
	Power       bool   `json:"power,omitempty"`
}

// BMCBoot struct used for marshal/unmarshal of BMC boot device
// valid boot devices are disk, cdrom, pxe, bios
// `omitempty` lets use the same struct for boot operations.BMCCommand
type BMCBoot struct {
	Device string `json:"device,omitempty"`
	Boot   struct {
		Action string `json:"action,omitempty"`
		Result bool   `json:"result,omitempty"`
	} `json:"boot,omitempty"`
}

// SendPowerCommand sends provided Action and State to foreman.  This
// performs an IPMI action against the provided host Expects Power or
// BMCBoot type struct populated with an action
//
// Example: https://<foreman>/api/hosts/<hostname>/boot
func (c *Client) SendPowerCommand(ctx context.Context, h *ForemanHost, cmd interface{}, retryCount int) error {
	// Initialize suffix variable,
	suffix := ""

	// Defines the suffix to append to the URL per operation type
	// Switch-Case against interface type to determine URL suffix
	switch v := cmd.(type) {
	case Power:
		suffix = PowerSuffix
	case BMCBoot:
		suffix = BootSuffix
	default:
		return fmt.Errorf("Invalid Operation: [%v]", v)
	}

	reqHost := fmt.Sprintf("/%s/%d/%s", HostEndpointPrefix, h.Id, suffix)

	JSONBytes, jsonEncErr := json.Marshal(cmd)
	if jsonEncErr != nil {
		return jsonEncErr
	}
	log.Debugf("JSONBytes: [%s]", JSONBytes)

	req, reqErr := c.NewRequestWithContext(ctx, http.MethodPut, reqHost, bytes.NewBuffer(JSONBytes))
	if reqErr != nil {
		return reqErr
	}

	retry := 0
	var sendErr error
	// retry until the successful Operation
	// or until # of allowed retries is reached
	for retry < retryCount {
		log.Debugf("SendPower: Retry #[%d]", retry)
		sendErr = c.SendAndParse(req, &cmd)
		if sendErr != nil {
			retry++
		} else {
			break
		}
	}

	if sendErr != nil {
		return sendErr
	}

	// Type Assertion to access map fields for Power and BMCBoot types
	powerMap, _ := cmd.(map[string]interface{})
	bootMap, _ := cmd.(map[string]map[string]interface{})

	log.Debugf("Power Response: [%+v]", cmd)

	// Test operation and return an error if result is false
	if powerMap[PowerSuffix] == false || bootMap[BootSuffix]["result"] == false {
		return fmt.Errorf("Failed Power Operation")
	}
	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateHost creates a new ForemanHost with the attributes of the supplied
// ForemanHost reference and returns the created ForemanHost reference.  The
// returned reference will have its ID and other API default values set by this
// function.
func (c *Client) CreateHost(ctx context.Context, h *ForemanHost, retryCount int) (*ForemanHost, error) {
	log.Tracef("foreman/api/host.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", HostEndpointPrefix)

	hJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("host", h)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("hJSONBytes: [%s]", hJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(hJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdHost foremanHostDecode

	retry := 0
	var sendErr error
	// retry until successful Host creation
	// or until # of allowed retries is reached
	for retry < retryCount {
		log.Debugf("CreatedHost: Retry #[%d]", retry)
		sendErr = c.SendAndParse(req, &createdHost)
		if sendErr != nil {
			retry++
		} else {
			break
		}
	}

	if sendErr != nil {
		return nil, sendErr
	}

	createdHost.InterfacesAttributes = createdHost.InterfacesAttributesDecode
	createdHost.PuppetClassIds = foremanObjectArrayToIdIntArray(createdHost.PuppetClassesDecode)
	createdHost.ConfigGroupIds = foremanObjectArrayToIdIntArray(createdHost.ConfigGroupsDecode)
	createdHost.HostParameters = createdHost.HostParametersDecode

	computeAttributes, _ := c.readComputeAttributes(ctx, createdHost.Id)
	if len(computeAttributes) > 0 {
		createdHost.ComputeAttributes = computeAttributes
	}

	log.Debugf("createdHost: [%+v]", createdHost)

	return &createdHost.ForemanHost, nil
}

// ReadHost reads the attributes of a ForemanHost identified by the supplied ID
// and returns a ForemanHost reference.
func (c *Client) ReadHost(ctx context.Context, id int) (*ForemanHost, error) {
	log.Tracef("foreman/api/host.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", HostEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readHost foremanHostDecode
	sendErr := c.SendAndParse(req, &readHost)
	if sendErr != nil {
		return nil, sendErr
	}

	computeAttributes, _ := c.readComputeAttributes(ctx, id)
	if len(computeAttributes) > 0 {
		readHost.ComputeAttributes = computeAttributes
	}
	readHost.InterfacesAttributes = readHost.InterfacesAttributesDecode
	readHost.PuppetClassIds = foremanObjectArrayToIdIntArray(readHost.PuppetClassesDecode)
	readHost.ConfigGroupIds = foremanObjectArrayToIdIntArray(readHost.ConfigGroupsDecode)
	readHost.HostParameters = readHost.HostParametersDecode

	return &readHost.ForemanHost, nil
}

// UpdateHost updates a ForemanHost's attributes.  The host with the ID of the
// supplied ForemanHost will be updated. A new ForemanHost reference is
// returned with the attributes from the result of the update operation.
func (c *Client) UpdateHost(ctx context.Context, h *ForemanHost, retryCount int) (*ForemanHost, error) {
	log.Tracef("foreman/api/host.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", HostEndpointPrefix, h.Id)

	hJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("host", h)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("hostJSONBytes: [%s]", hJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(hJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedHost foremanHostDecode
	retry := 0
	var sendErr error
	// retry until the successful Host Update
	// or until # of allowed retries is reached
	for retry < retryCount {
		log.Debugf("UpdateHost: Retry #[%d]", retry)
		sendErr = c.SendAndParse(req, &updatedHost)
		if sendErr != nil {
			retry++
		} else {
			break
		}
	}

	if sendErr != nil {
		return nil, sendErr
	}

	computeAttributes, _ := c.readComputeAttributes(ctx, h.Id)
	if len(computeAttributes) > 0 {
		updatedHost.ComputeAttributes = computeAttributes
	}
	updatedHost.InterfacesAttributes = updatedHost.InterfacesAttributesDecode
	updatedHost.PuppetClassIds = foremanObjectArrayToIdIntArray(updatedHost.PuppetClassesDecode)
	updatedHost.ConfigGroupIds = foremanObjectArrayToIdIntArray(updatedHost.ConfigGroupsDecode)
	updatedHost.HostParameters = updatedHost.HostParametersDecode
	log.Debugf("updatedHost: [%+v]", updatedHost)

	return &updatedHost.ForemanHost, nil
}

// DeleteHost deletes the ForemanHost identified by the supplied ID
func (c *Client) DeleteHost(ctx context.Context, id int) error {
	log.Tracef("foreman/api/host.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", HostEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return reqErr
	}

	return c.SendAndParse(req, nil)
}

// Compute Attributes are only available via dedicated API endpoint. readComputeAttributes gets this endpoint.
func (c *Client) readComputeAttributes(ctx context.Context, id int) (map[string]interface{}, error) {

	reqEndpoint := fmt.Sprintf("/%s/%d/%s", HostEndpointPrefix, id, ComputeAttributesSuffix)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readVmAttributes map[string]interface{}
	sendErr := c.SendAndParse(req, &readVmAttributes)
	if sendErr != nil {
		return nil, sendErr
	}

	readVmAttributesStr := make(map[string]interface{}, len(readVmAttributes))

	for idx, val := range readVmAttributes {
		readVmAttributesStr[idx] = val
	}

	return readVmAttributesStr, nil
}

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	// HostEndpointPrefix : Prefix appended to API url for hosts
	HostEndpointPrefix = "hosts"
	// PowerSuffix : Suffix appended to API url for power operations
	PowerSuffix = "power"
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
	// ID of the domain to assign the host
	DomainId int `json:"domain_id"`
	// ID of the environment to assign the host
	EnvironmentId int `json:"environment_id"`
	// ID of the hostgroup to assign the host
	HostgroupId int `json:"hostgroup_id"`
	// ID of the operating system to put on the host
	OperatingSystemId int `json:"operatingsystem_id"`
	// Whether or not to Enable BMC Functionality on this host
	EnableBMC bool
	// Boolean to track success of BMC Calls
	BMCSuccess bool
	// Additional information about this host
	Comment string `json:"comment"`
	// Nested struct defining any interfaces associated with the Host
	InterfacesAttributes []ForemanInterfacesAttribute `json:"interfaces_attributes"`
}

// ForemanInterfacesAttribute representing a hosts defined network interfaces
type ForemanInterfacesAttribute struct {
	Id         int    `json:"id,omitempty"`
	SubnetId   int    `json:"subnet_id"`
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Managed    bool   `json:"managed"`
	Provision  bool   `json:"provision"`
	Virtual    bool   `json:"virtual"`
	Primary    bool   `json:"primary"`
	IP         string `json:"ip"`
	MAC        string `json:"mac"`
	Type       string `json:"type"`
	Provider   string `json:"provider"`
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

// foremanHostJSON struct used for JSON decode.
type foremanHostJSON struct {
	InterfacesAttributes []ForemanInterfacesAttribute `json:"interfaces"`
}

// Power struct for marshal/unmarshal of power state
// valid states are on, off, soft, cycle, state
// `omitempty`` lets use the same struct for power operations.Command
type Power struct {
	PowerAction string `json:"power_action,omitempty"`
	Power       bool   `json:"power,omitempty"`
}

// BMCBoot struct used for marshal/unmarshal of BMC boot device
// valid boot devices are disk, cdrom, pxe, bios
// `omitempty`` lets use the same struct for boot operations.BMCCommand
type BMCBoot struct {
	Device string `json:"device,omitempty"`
	Boot   struct {
		Action string `json:"action,omitempty"`
		Result bool   `json:"result,omitempty"`
	} `json:"boot,omitempty"`
}

// Implement the Marshaler interface
func (fh ForemanHost) MarshalJSON() ([]byte, error) {
	log.Tracef("foreman/api/host.go#MarshalJSON")

	fhMap := map[string]interface{}{}

	fhMap["name"] = fh.Name
	fhMap["comment"] = fh.Comment
	fhMap["build"] = fh.Build
	fhMap["domain_id"] = intIdToJSONString(fh.DomainId)
	fhMap["operatingsystem_id"] = intIdToJSONString(fh.OperatingSystemId)
	fhMap["hostgroup_id"] = intIdToJSONString(fh.HostgroupId)
	fhMap["environment_id"] = intIdToJSONString(fh.EnvironmentId)
	if len(fh.InterfacesAttributes) > 0 {
		fhMap["interfaces_attributes"] = fh.InterfacesAttributes
	}
	log.Debugf("fhMap: [%+v]", fhMap)

	return json.Marshal(fhMap)
}

// Custom JSON unmarshal function. Unmarshal to the unexported JSON struct
// and then convert over to a ForemanHost struct.
func (fh *ForemanHost) UnmarshalJSON(b []byte) error {
	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fh.ForemanObject = fo

	// Unmarshal to temporary JSON struct to get the properties with differently
	// named keys
	var fhJSON foremanHostJSON
	jsonDecErr = json.Unmarshal(b, &fhJSON)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fh.InterfacesAttributes = fhJSON.InterfacesAttributes

	// Unmarshal into mapstructure and set the rest of the struct properties
	// NOTE(ALL): Properties unmarshalled are of type float64 as opposed to int, hence the below testing
	// Without this, properties will define as default values in state file.
	var fhMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &fhMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	log.Debugf("fhMap: [%v]", fhMap)
	var ok bool
	if fh.Build, ok = fhMap["build"].(bool); !ok {
		fh.Build = false
	}
	if fh.Comment, ok = fhMap["comment"].(string); !ok {
		fh.Comment = ""
	}
	if _, ok = fhMap["domain_id"].(float64); !ok {
		fh.DomainId = 0
	} else {
		fh.DomainId = int(fhMap["domain_id"].(float64))
	}
	if _, ok = fhMap["environment_id"].(float64); !ok {
		fh.EnvironmentId = 0
	} else {
		fh.EnvironmentId = int(fhMap["environment_id"].(float64))
	}
	if _, ok = fhMap["hostgroup_id"].(float64); !ok {
		fh.HostgroupId = 0
	} else {
		fh.HostgroupId = int(fhMap["hostgroup_id"].(float64))
	}
	if _, ok = fhMap["operatingsystem_id"].(float64); !ok {
		fh.OperatingSystemId = 0
	} else {
		fh.OperatingSystemId = int(fhMap["operatingsystem_id"].(float64))
	}

	return nil
}

// SendPowerCommand sends provided Action and State to foreman.  This
// performs an IPMI action against the provided host Expects Power or
// BMCBoot type struct populated with an action
//
// Example: https://<foreman>/api/hosts/<hostname>/boot
func (c *Client) SendPowerCommand(h *ForemanHost, cmd interface{}, retryCount int) error {
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

	reqHost := fmt.Sprintf("/%s/%s/%s", HostEndpointPrefix, h.Name, suffix)

	JSONBytes, jsonEncErr := json.Marshal(cmd)
	if jsonEncErr != nil {
		return jsonEncErr
	}
	log.Debugf("JSONBytes: [%s]", JSONBytes)

	req, reqErr := c.NewRequest(http.MethodPut, reqHost, bytes.NewBuffer(JSONBytes))
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
func (c *Client) CreateHost(h *ForemanHost, retryCount int) (*ForemanHost, error) {
	log.Tracef("foreman/api/host.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", HostEndpointPrefix)

	hJSONBytes, jsonEncErr := json.Marshal(h)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("hJSONBytes: [%s]", hJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(hJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdHost ForemanHost

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

	log.Debugf("createdHost: [%+v]", createdHost)

	return &createdHost, nil
}

// ReadHost reads the attributes of a ForemanHost identified by the supplied ID
// and returns a ForemanHost reference.
func (c *Client) ReadHost(id int) (*ForemanHost, error) {
	log.Tracef("foreman/api/host.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", HostEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readHost ForemanHost
	sendErr := c.SendAndParse(req, &readHost)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readHost: [%+v]", readHost)

	return &readHost, nil
}

// UpdateHost updates a ForemanHost's attributes.  The host with the ID of the
// supplied ForemanHost will be updated. A new ForemanHost reference is
// returned with the attributes from the result of the update operation.
func (c *Client) UpdateHost(h *ForemanHost, retryCount int) (*ForemanHost, error) {
	log.Tracef("foreman/api/host.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", HostEndpointPrefix, h.Id)

	hJSONBytes, jsonEncErr := json.Marshal(h)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("hostJSONBytes: [%s]", hJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(hJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedHost ForemanHost
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

	log.Debugf("updatedHost: [%+v]", updatedHost)

	return &updatedHost, nil
}

// DeleteHost deletes the ForemanHost identified by the supplied ID
func (c *Client) DeleteHost(id int) error {
	log.Tracef("foreman/api/host.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", HostEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodDelete,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return reqErr
	}

	return c.SendAndParse(req, nil)
}

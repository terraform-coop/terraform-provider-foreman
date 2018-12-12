package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	HostgroupEndpointPrefix = "hostgroups"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanHostgroup API model represents a hostgroup.  Hostgroups are
// organized in a tree-like structure and can inherit the values of their
// parent hostgroups.  The relationship is maintained through the parent_id
// attribute.
//
// When hosts get associated with a hostgroup, it will inherit attributes
// from the hostgroup. This allows for easy shared configuration of various
// hosts based on common attributes.
type ForemanHostgroup struct {
	// Inherits the base object's attributes
	ForemanObject

	// The title is a computed property representing the fullname of the
	// hostgroup.  A hostgroup's title is a path-like string from the head
	// of the hostgroup tree down to this hostgroup.  The title will be
	// in the form of: "<parent 1>/<parent 2>/.../<name>"
	Title string `json:"title"`
	// ID of the architecture associated with this hostgroup
	ArchitectureId int `json:"architecture_id"`
	// ID of the compute profile associated with this hostgroup
	ComputeProfileId int `json:"compute_profile_id"`
	// ID of the domain associated with this hostgroup
	DomainId int `json:"domain_id"`
	// ID of the environment associated with this hostgroup
	EnvironmentId int `json:"environment_id"`
	// ID of the media associated with this hostgroup
	MediaId int `json:"medium_id"`
	// ID of the operating system associated with this hostgroup
	OperatingSystemId int `json:"operatingsystem_id"`
	// ID of this hostgroup's parent hostgroup
	ParentId int `json:"parent_id"`
	// ID of the partition table to use with this hostgroup
	PartitionTableId int `json:"ptable_id"`
	// ID of the smart proxy acting as the puppet certificate authority
	// server for the hostgroup
	PuppetCAProxyId int `json:"puppet_ca_proxy_id"`
	// ID of the smart proxy acting as the puppet proxy server for the
	// hostgroup
	PuppetProxyId int `json:"puppet_proxy_id"`
	// ID of the realm associated with the hostgroup
	RealmId int `json:"realm_id"`
	// ID of the subnet associated with the hostgroup
	SubnetId int `json:"subnet_id"`
}

// Implement the Marshaler interface
func (fh ForemanHostgroup) MarshalJSON() ([]byte, error) {
	log.Tracef("foreman/api/hostgroup.go#MarshalJSON")

	// NOTE(ALL): omit the "title" property from the JSON marshal since it is a
	//   computed value

	fhMap := map[string]interface{}{}

	fhMap["name"] = fh.Name

	fhMap["architecture_id"] = intIdToJSONString(fh.ArchitectureId)
	fhMap["compute_profile_id"] = intIdToJSONString(fh.ComputeProfileId)
	fhMap["domain_id"] = intIdToJSONString(fh.DomainId)
	fhMap["environment_id"] = intIdToJSONString(fh.EnvironmentId)
	fhMap["medium_id"] = intIdToJSONString(fh.MediaId)
	fhMap["operatingsystem_id"] = intIdToJSONString(fh.OperatingSystemId)
	fhMap["parent_id"] = intIdToJSONString(fh.ParentId)
	fhMap["ptable_id"] = intIdToJSONString(fh.PartitionTableId)
	fhMap["puppet_ca_proxy_id"] = intIdToJSONString(fh.PuppetCAProxyId)
	fhMap["puppet_proxy_id"] = intIdToJSONString(fh.PuppetProxyId)
	fhMap["realm_id"] = intIdToJSONString(fh.RealmId)
	fhMap["subnet_id"] = intIdToJSONString(fh.SubnetId)

	log.Debugf("fhMap: [%v]", fhMap)

	return json.Marshal(fhMap)
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateHostgroup creates a new ForemanHostgroup with the attributes of the
// supplied ForemanHostgroup reference and returns the created ForemanHostgroup
// reference.  The returned reference will have its ID and other API default
// values set by this function.
func (c *Client) CreateHostgroup(h *ForemanHostgroup) (*ForemanHostgroup, error) {
	log.Tracef("foreman/api/hostgroup.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", HostgroupEndpointPrefix)

	hJSONBytes, jsonEncErr := json.Marshal(h)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("hostgroupJSONBytes: [%s]", hJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(hJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdHostgroup ForemanHostgroup
	sendErr := c.SendAndParse(req, &createdHostgroup)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdHostgroup: [%+v]", createdHostgroup)

	return &createdHostgroup, nil
}

// ReadHostgroup reads the attributes of a ForemanHostgroup identified by the
// supplied ID and returns a ForemanHostgroup reference.
func (c *Client) ReadHostgroup(id int) (*ForemanHostgroup, error) {
	log.Tracef("foreman/api/hostgroup.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", HostgroupEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readHostgroup ForemanHostgroup
	sendErr := c.SendAndParse(req, &readHostgroup)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readHostgroup: [%+v]", readHostgroup)

	return &readHostgroup, nil
}

// UpdateHostgroup updates a ForemanHostgroup's attributes.  The hostgroup with
// the ID of the supplied ForemanHostgroup will be updated. A new
// ForemanHostgroup reference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateHostgroup(h *ForemanHostgroup) (*ForemanHostgroup, error) {
	log.Tracef("foreman/api/hostgroup.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", HostgroupEndpointPrefix, h.Id)

	hJSONBytes, jsonEncErr := json.Marshal(h)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("hostgroupJSONBytes: [%s]", hJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(hJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedHostgroup ForemanHostgroup
	sendErr := c.SendAndParse(req, &updatedHostgroup)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedHostgroup: [%+v]", updatedHostgroup)

	return &updatedHostgroup, nil
}

// DeleteHostgroup deletes the ForemanHostgroup identified by the supplied ID
func (c *Client) DeleteHostgroup(id int) error {
	log.Tracef("foreman/api/hostgroup.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", HostgroupEndpointPrefix, id)

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

// -----------------------------------------------------------------------------
// Query Implementation
// -----------------------------------------------------------------------------

// QueryHostgroup queries for a ForemanHostgroup based on the attributes of the
// supplied ForemanHostgroup reference and returns a QueryResponse struct
// containing query/response metadata and the matching hostgroups.
func (c *Client) QueryHostgroup(h *ForemanHostgroup) (QueryResponse, error) {
	log.Tracef("foreman/api/hostgroup.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", HostgroupEndpointPrefix)
	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return queryResponse, reqErr
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	title := `"` + h.Title + `"`
	reqQuery.Set("search", "title="+title)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanHostgroup for
	// the results
	results := []ForemanHostgroup{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanHostgroup to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

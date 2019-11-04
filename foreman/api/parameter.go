package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	ParameterEndpointPrefix = "/%s/%d/parameters"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanParameter API model represents the parameter name. Parameters serve as an
// identification string that defines autonomy, authority, or control for
// a portion of a network.
type ForemanParameter struct {
	// Inherits the base object's attributes
	ForemanObject

	// Subject tells us what this parameter is actually referencing
	// Currently supports "host", "hostgroup", "domain", "operatingsystem" "subnet"
	Subject string

	// One of the ID fields should be set
	HostID            int `json:"host_id,omitempty"`
	HostGroupID       int `json:"hostgroup_id,omitempty"`
	DomainID          int `json:"parameter_id,omitempty"`
	OperatingSystemID int `json:"operatingsystem_id,omitempty"`
	SubnetID          int `json:"subnet_id,omitempty"`
	// The Parameter we actually send
	Parameter ForemanKVParameter `json:"parameter"`
}

func (fp *ForemanParameter) apiEndpoint() (string, int) {
	if fp.HostID != 0 {
		return "hosts", fp.HostID
	} else if fp.HostGroupID != 0 {
		return "hostgroups", fp.HostGroupID
	} else if fp.DomainID != 0 {
		return "domains", fp.DomainID
	} else if fp.OperatingSystemID != 0 {
		return "operatingsystems", fp.OperatingSystemID
	} else if fp.SubnetID != 0 {
		return "subnets", fp.SubnetID
	}
	return "", -1
}

func (fp *ForemanParameter) UnmarshalJSON(b []byte) error {
	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fp.ForemanObject = fo

	var fpMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &fpMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}

	var ok bool
	if fp.Parameter.Name, ok = fpMap["name"].(string); !ok {
		fp.Parameter.Name = ""
	}
	if fp.Parameter.Value, ok = fpMap["value"].(string); !ok {
		fp.Parameter.Value = ""
	}

	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateParameter creates a new ForemanParameter with the attributes of the supplied
// ForemanParameter reference and returns the created ForemanParameter reference.
// The returned reference will have its ID and other API default values set by
// this function.
func (c *Client) CreateParameter(d *ForemanParameter) (*ForemanParameter, error) {
	log.Tracef("foreman/api/parameter.go#Create")

	selEndA, selEndB := d.apiEndpoint()
	reqEndpoint := fmt.Sprintf(ParameterEndpointPrefix, selEndA, selEndB)

	// All parameters are send individually. Yeay for that
	var createdParameter ForemanParameter
	parameterJSONBytes, jsonEncErr := json.Marshal(d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("parameterJSONBytes: [%s]", parameterJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(parameterJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	sendErr := c.SendAndParse(req, &createdParameter)
	if sendErr != nil {
		return nil, sendErr
	}
	log.Debugf("createdParameter: [%+v]", createdParameter)

	d.Id = createdParameter.Id
	d.Parameter = createdParameter.Parameter
	return d, nil
}

// ReadParameter reads the attributes of a ForemanParameter identified by the
// supplied ID and returns a ForemanParameter reference.
func (c *Client) ReadParameter(d *ForemanParameter, id int) (*ForemanParameter, error) {
	log.Tracef("foreman/api/parameter.go#Read")

	selEndA, selEndB := d.apiEndpoint()
	reqEndpoint := fmt.Sprintf(ParameterEndpointPrefix+"/%d", selEndA, selEndB, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readParameter ForemanParameter
	sendErr := c.SendAndParse(req, &readParameter)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readParameter: [%+v]", readParameter)

	d.Id = readParameter.Id
	d.Parameter = readParameter.Parameter
	return d, nil
}

// UpdateParameter deletes all parameters for the subject resource and re-creates them
// as we look at them differently on either side this is the safest way to reach sync
func (c *Client) UpdateParameter(d *ForemanParameter, id int) (*ForemanParameter, error) {
	log.Tracef("foreman/api/parameter.go#Update")

	selEndA, selEndB := d.apiEndpoint()
	reqEndpoint := fmt.Sprintf(ParameterEndpointPrefix+"/%d", selEndA, selEndB, id)

	parameterJSONBytes, jsonEncErr := json.Marshal(d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("parameterJSONBytes: [%s]", parameterJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(parameterJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedParameter ForemanParameter
	sendErr := c.SendAndParse(req, &updatedParameter)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedParameter: [%+v]", updatedParameter)

	d.Id = updatedParameter.Id
	d.Parameter = updatedParameter.Parameter
	return d, nil
}

// DeleteParameter deletes the ForemanParameters for the given resource
func (c *Client) DeleteParameter(d *ForemanParameter, id int) error {
	log.Tracef("foreman/api/parameter.go#Delete")

	selEndA, selEndB := d.apiEndpoint()
	reqEndpoint := fmt.Sprintf(ParameterEndpointPrefix+"/%d", selEndA, selEndB, id)

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

// QueryParameter queries for a ForemanParameter based on the attributes of the
// supplied ForemanParameter reference and returns a QueryResponse struct
// containing query/response metadata and the matching parameters.
func (c *Client) QueryParameter(d *ForemanParameter) (QueryResponse, error) {
	log.Tracef("foreman/api/parameter.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", ParameterEndpointPrefix)
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
	name := `"` + d.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanParameter for
	// the results
	results := []ForemanParameter{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanParameter to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

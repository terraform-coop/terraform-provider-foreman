package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	ComputeResourceEndpointPrefix = "compute_resources"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanComputeResource API model represents the computeresource name. ComputeResources serve as an
// identification string that defines autonomy, authority, or control for
// a portion of a network.

type ForemanComputeResource struct {
	// Inherits the base object's attributes
	ForemanObject

	Description string `json:"description"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	DisplayType string `json:"display_type"`
	// VMWare specific
	User       string `json:"user,omitempty"`
	Password   string `json:"password,omitempty"`
	Datacenter string `json:"datacenter,omitempty"`
	Server     string `json:"server,omitempty"`
	// VMWare and Libvirt
	SetConsolePassword bool `json:"set_console_password,omitempty"`
	CachingEnabled     bool `json:"caching_enabled,omitempty"`
}

// Custom JSON unmarshal function. Unmarshal to the unexported JSON struct
// and then convert over to a ForemanComputeResource struct.
func (fcr *ForemanComputeResource) UnmarshalJSON(b []byte) error {
	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fcr.ForemanObject = fo

	// Unmarshal into mapstructure and set the rest of the struct properties
	// NOTE(ALL): Properties unmarshalled are of type float64 as opposed to int, hence the below testing
	// Without this, properties will define as default values in state file.
	var fcrMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &fcrMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	log.Debugf("fcrMap: [%v]", fcrMap)
	var ok bool
	if fcr.Description, ok = fcrMap["description"].(string); !ok {
		fcr.Description = ""
	}
	if fcr.URL, ok = fcrMap["url"].(string); !ok {
		fcr.URL = ""
	}
	if fcr.Name, ok = fcrMap["name"].(string); !ok {
		fcr.Name = ""
	}
	if fcr.Provider, ok = fcrMap["provider"].(string); !ok {
		fcr.Provider = ""
	}
	if fcr.DisplayType, ok = fcrMap["displaytype"].(string); !ok {
		fcr.DisplayType = ""
	}
	if fcr.User, ok = fcrMap["user"].(string); !ok {
		fcr.User = ""
	}
	if fcr.Password, ok = fcrMap["password"].(string); !ok {
		fcr.Password = ""
	}
	if fcr.Datacenter, ok = fcrMap["datacenter"].(string); !ok {
		fcr.Datacenter = ""
	}
	if fcr.Server, ok = fcrMap["server"].(string); !ok {
		fcr.Server = ""
	}
	if fcr.SetConsolePassword, ok = fcrMap["set_console_password"].(bool); !ok {
		fcr.SetConsolePassword = false
	}
	if fcr.CachingEnabled, ok = fcrMap["caching_enabled"].(bool); !ok {
		fcr.CachingEnabled = false
	}

	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateComputeResource creates a new ForemanComputeResource with the attributes of the supplied
// ForemanComputeResource reference and returns the created ForemanComputeResource reference.
// The returned reference will have its ID and other API default values set by
// this function.
func (c *Client) CreateComputeResource(d *ForemanComputeResource) (*ForemanComputeResource, error) {
	log.Tracef("foreman/api/computeresource.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", ComputeResourceEndpointPrefix)

	computeresourceJSONBytes, jsonEncErr := WrapJson("compute_resource", d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("computeresourceJSONBytes: [%s]", computeresourceJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(computeresourceJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdComputeResource ForemanComputeResource
	sendErr := c.SendAndParse(req, &createdComputeResource)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdComputeResource: [%+v]", createdComputeResource)

	return &createdComputeResource, nil
}

// ReadComputeResource reads the attributes of a ForemanComputeResource identified by the
// supplied ID and returns a ForemanComputeResource reference.
func (c *Client) ReadComputeResource(id int) (*ForemanComputeResource, error) {
	log.Tracef("foreman/api/computeresource.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", ComputeResourceEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readComputeResource ForemanComputeResource
	sendErr := c.SendAndParse(req, &readComputeResource)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readComputeResource: [%+v]", readComputeResource)

	return &readComputeResource, nil
}

// UpdateComputeResource updates a ForemanComputeResource's attributes.  The computeresource with the ID
// of the supplied ForemanComputeResource will be updated. A new ForemanComputeResource reference
// is returned with the attributes from the result of the update operation.
func (c *Client) UpdateComputeResource(d *ForemanComputeResource) (*ForemanComputeResource, error) {
	log.Tracef("foreman/api/computeresource.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", ComputeResourceEndpointPrefix, d.Id)

	computeresourceJSONBytes, jsonEncErr := WrapJson("compute_resource", d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("computeresourceJSONBytes: [%s]", computeresourceJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(computeresourceJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedComputeResource ForemanComputeResource
	sendErr := c.SendAndParse(req, &updatedComputeResource)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedComputeResource: [%+v]", updatedComputeResource)

	return &updatedComputeResource, nil
}

// DeleteComputeResource deletes the ForemanComputeResource identified by the supplied ID
func (c *Client) DeleteComputeResource(id int) error {
	log.Tracef("foreman/api/computeresource.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", ComputeResourceEndpointPrefix, id)

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

// QueryComputeResource queries for a ForemanComputeResource based on the attributes of the
// supplied ForemanComputeResource reference and returns a QueryResponse struct
// containing query/response metadata and the matching computeresources.
func (c *Client) QueryComputeResource(d *ForemanComputeResource) (QueryResponse, error) {
	log.Tracef("foreman/api/computeresource.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", ComputeResourceEndpointPrefix)
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
	// Encode back to JSON, then Unmarshal into []ForemanComputeResource for
	// the results
	results := []ForemanComputeResource{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanComputeResource to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

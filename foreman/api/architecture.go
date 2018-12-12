package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	ArchitectureEndpointPrefix = "architectures"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanArchitecture API model represents an instruction set
// architecture (ISA)
type ForemanArchitecture struct {
	// Inherits the base object's attributes
	ForemanObject

	// Array of ForemanOperatingSystem IDs associated with this architecture
	OperatingSystemIds []int `json:"operatingsystem_ids"`
}

// ForemanArchitecture struct used for JSON decode.  Foreman API returns
// the operating system ids back as a list of ForemanObjects with some of
// the attributes of an operating system.  However, we are only interested in
// the IDs returned.
type foremanArchitectureJSON struct {
	OperatingSystems []ForemanObject `json:"operatingsystems"`
}

// Custom JSON unmarshal function.  Unmarshal to the unexported JSON struct
// and then convert over to a ForemanArchitecture struct.
func (fa *ForemanArchitecture) UnmarshalJSON(b []byte) error {
	log.Tracef("foreman/api/architecture.go#UnmarshalJSON")

	var jsonDecErr error

	// decode base forman object
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fa.ForemanObject = fo

	// decode special JSON struct for keys that changed names
	var faJSON foremanArchitectureJSON
	jsonDecErr = json.Unmarshal(b, &faJSON)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fa.OperatingSystemIds = foremanObjectArrayToIdIntArray(faJSON.OperatingSystems)

	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateArchitecture creates a new ForemanArchitecture with the attributes of
// the supplied ForemanArchitecture reference and returns the created
// ForemanArchitecture reference.  The returned reference will have its ID and
// other API default values set by this function.
func (c *Client) CreateArchitecture(a *ForemanArchitecture) (*ForemanArchitecture, error) {
	log.Tracef("foreman/api/architecture.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", ArchitectureEndpointPrefix)

	archJSONBytes, jsonEncErr := json.Marshal(a)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("archJSONBytes: [%s]", archJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(archJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdArch ForemanArchitecture
	sendErr := c.SendAndParse(req, &createdArch)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdArch: [%+v]", createdArch)

	return &createdArch, nil
}

// ReadArchitecture reads the attributes of a ForemanArchitecture identified by
// the supplied ID and returns a ForemanArchitecture reference.
func (c *Client) ReadArchitecture(id int) (*ForemanArchitecture, error) {
	log.Tracef("foreman/api/architecture.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", ArchitectureEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readArch ForemanArchitecture
	sendErr := c.SendAndParse(req, &readArch)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readArch: [%+v]", readArch)

	return &readArch, nil
}

// UpdateArchitecture updates a ForemanArchitecture's attributes.  The
// architecture with the ID of the supplied ForemanArchitecture will be
// updated. A new ForemanArchitecture reference is returned with the attributes
// from the result of the update operation.
func (c *Client) UpdateArchitecture(a *ForemanArchitecture) (*ForemanArchitecture, error) {
	log.Tracef("foreman/api/architecture.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", ArchitectureEndpointPrefix, a.Id)

	archJSONBytes, jsonEncErr := json.Marshal(a)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("archJSONBytes: [%s]", archJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(archJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedArch ForemanArchitecture
	sendErr := c.SendAndParse(req, &updatedArch)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedArch: [%+v]", updatedArch)

	return &updatedArch, nil
}

// DeleteArchitecture deletes the ForemanArchitecture identified by the
// supplied ID
func (c *Client) DeleteArchitecture(id int) error {
	log.Tracef("foreman/api/architecture.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", ArchitectureEndpointPrefix, id)

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

// QueryArchitecture queries for a ForemanArchitecture based on the attributes
// of the supplied ForemanArchitecture reference and returns a QueryResponse
// struct containing query/response metadata and the matching architectures.
func (c *Client) QueryArchitecture(a *ForemanArchitecture) (QueryResponse, error) {
	log.Tracef("foreman/api/architecture.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", ArchitectureEndpointPrefix)
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
	name := `"` + a.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanArchitecture for
	// the results
	results := []ForemanArchitecture{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanArchitecture to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"

	"github.com/HanseMerkur/terraform-provider-utils/log"
)

const (
	UsergroupEndpointPrefix = "usergroups"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanUsergroup API model represents a usergroup.
type ForemanUsergroup struct {
	// Inherits the base object's attributes
	ForemanObject

	// enables or disables admin access for group members, Must be one of: true, false, 1, 0.
	Admin bool `json:"admin"`
}

// Implement the Marshaler interface
func (fh ForemanUsergroup) MarshalJSON() ([]byte, error) {
	utils.TraceFunctionCall()

	// NOTE(ALL): omit the "name" property from the JSON marshal since it is a
	//   computed value

	fhMap := map[string]interface{}{}

	fhMap["name"] = fh.Name
	fhMap["admin"] = fh.Admin

	log.Debugf("fhMap: [%v]", fhMap)

	return json.Marshal(fhMap)
}

func (fh *ForemanUsergroup) UnmarshalJSON(b []byte) error {
	utils.TraceFunctionCall()

	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fh.ForemanObject = fo

	// Unmarshal into mapstructure and set the rest of the struct properties
	var fhMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &fhMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}

	var ok bool
	if fh.Admin, ok = fhMap["admin"].(bool); !ok {
		fh.Admin = false
	}

	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateUsergroup creates a new ForemanUsergroup with the attributes of the
// supplied ForemanUsergroup reference and returns the created ForemanUsergroup
// reference.  The returned reference will have its ID and other API default
// values set by this function.
func (c *Client) CreateUsergroup(ctx context.Context, h *ForemanUsergroup) (*ForemanUsergroup, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s", UsergroupEndpointPrefix)

	hJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("usergroup", h)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("usergroupJSONBytes: [%s]", hJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(hJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdUsergroup ForemanUsergroup
	sendErr := c.SendAndParse(req, &createdUsergroup)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdUsergroup: [%+v]", createdUsergroup)

	return &createdUsergroup, nil
}

// ReadUsergroup reads the attributes of a ForemanUsergroup identified by the
// supplied ID and returns a ForemanUsergroup reference.
func (c *Client) ReadUsergroup(ctx context.Context, id int) (*ForemanUsergroup, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", UsergroupEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readUsergroup ForemanUsergroup
	sendErr := c.SendAndParse(req, &readUsergroup)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readUsergroup: [%+v]", readUsergroup)

	return &readUsergroup, nil
}

// UpdateUsergroup updates a ForemanUsergroup's attributes.  The usergroup with
// the ID of the supplied ForemanUsergroup will be updated. A new
// ForemanUsergroup reference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateUsergroup(ctx context.Context, h *ForemanUsergroup) (*ForemanUsergroup, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", UsergroupEndpointPrefix, h.Id)

	hJSONBytes, jsonEncErr := c.WrapJSON("usergroup", h)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("usergroupJSONBytes: [%s]", hJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(hJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedUsergroup ForemanUsergroup
	sendErr := c.SendAndParse(req, &updatedUsergroup)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedUsergroup: [%+v]", updatedUsergroup)

	return &updatedUsergroup, nil
}

// DeleteUsergroup deletes the ForemanUsergroup identified by the supplied ID
func (c *Client) DeleteUsergroup(ctx context.Context, id int) error {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", UsergroupEndpointPrefix, id)

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

// -----------------------------------------------------------------------------
// Query Implementation
// -----------------------------------------------------------------------------

// QueryUsergroup queries for a ForemanUsergroup based on the attributes of the
// supplied ForemanUsergroup reference and returns a QueryResponse struct
// containing query/response metadata and the matching usergroups.
func (c *Client) QueryUsergroup(ctx context.Context, u *ForemanUsergroup) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", UsergroupEndpointPrefix)
	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return queryResponse, reqErr
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	name := `"` + u.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanUsergroup for
	// the results
	results := []ForemanUsergroup{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanUsergroup to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"
)

const (
	// KatelloContentCredentialEndpointPrefix api endpoint prefix for katello content_credentials
	// 'katello/ will be removed, it's a marker to detect talking with katello api
	KatelloContentCredentialEndpointPrefix = "katello/content_credentials"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// ForemanKatelloContentCredential API model representing a content credential.
// A content credential is used to sign a repository in katello.
type ForemanKatelloContentCredential struct {
	// Inherits the base object's attributes
	ForemanObject

	// Public key block in DER encoding
	Content string `json:"content"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateKatelloContentCredential creates a new ForemanKatelloContentCredential with the attributes of the
// supplied ForemanKatelloContentCredential reference and returns the created
// ForemanKatelloContentCredential reference.  The returned reference will have its ID and
// other API default values set by this function.
func (c *Client) CreateKatelloContentCredential(ctx context.Context, s *ForemanKatelloContentCredential) (*ForemanKatelloContentCredential, error) {
	utils.TraceFunctionCall()

	sJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy(nil, s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	utils.Debugf("KatelloContentCredentialJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		KatelloContentCredentialEndpointPrefix,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdKatelloContentCredential ForemanKatelloContentCredential
	sendErr := c.SendAndParse(req, &createdKatelloContentCredential)
	if sendErr != nil {
		return nil, sendErr
	}

	utils.Debugf("createdKatelloContentCredential: [%+v]", createdKatelloContentCredential)

	return &createdKatelloContentCredential, nil
}

// ReadKatelloContentCredential reads the attributes of a ForemanKatelloContentCredential identified by the
// supplied ID and returns a ForemanKatelloContentCredential reference.
func (c *Client) ReadKatelloContentCredential(ctx context.Context, id int) (*ForemanKatelloContentCredential, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloContentCredentialEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readKatelloContentCredential ForemanKatelloContentCredential
	sendErr := c.SendAndParse(req, &readKatelloContentCredential)
	if sendErr != nil {
		return nil, sendErr
	}

	utils.Debugf("readKatelloContentCredential: [%+v]", readKatelloContentCredential)

	return &readKatelloContentCredential, nil
}

// UpdateKatelloContentCredential updates a ForemanKatelloContentCredential's attributes.  The smart proxy
// with the ID of the supplied ForemanKatelloContentCredential will be updated. A new
// ForemanKatelloContentCredential reference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateKatelloContentCredential(ctx context.Context, s *ForemanKatelloContentCredential) (*ForemanKatelloContentCredential, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloContentCredentialEndpointPrefix, s.Id)

	sJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy(nil, s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	utils.Debugf("KatelloContentCredentialJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedKatelloContentCredential ForemanKatelloContentCredential
	sendErr := c.SendAndParse(req, &updatedKatelloContentCredential)
	if sendErr != nil {
		return nil, sendErr
	}

	utils.Debugf("updatedKatelloContentCredential: [%+v]", updatedKatelloContentCredential)

	return &updatedKatelloContentCredential, nil
}

// DeleteKatelloContentCredential deletes the ForemanKatelloContentCredential identified by the supplied ID
func (c *Client) DeleteKatelloContentCredential(ctx context.Context, id int) error {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloContentCredentialEndpointPrefix, id)

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

// QueryKatelloContentCredential queries for a ForemanKatelloContentCredential based on the attributes of
// the supplied ForemanKatelloContentCredential reference and returns a QueryResponse struct
// containing query/response metadata and the matching smart proxy.
func (c *Client) QueryKatelloContentCredential(ctx context.Context, s *ForemanKatelloContentCredential) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	sJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy(nil, s)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}

	utils.Debugf("KatelloContentCredentialJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		KatelloContentCredentialEndpointPrefix,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return queryResponse, reqErr
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	name := `"` + s.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	utils.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanKatelloContentCredential for
	// the results
	results := []ForemanKatelloContentCredential{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanKatelloContentCredential to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

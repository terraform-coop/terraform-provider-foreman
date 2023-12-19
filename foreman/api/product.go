package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/log"
)

const (
	// KatelloProductEndpointPrefix api endpoint prefix for katello products
	// 'katello/ will be removed, it's a marker to detect talking with katello api
	KatelloProductEndpointPrefix = "katello/products"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// ForemanKatelloProduct API model representing a product.
type ForemanKatelloProduct struct {
	// Inherits the base object's attributes
	ForemanObject

	Description     string `json:"description"`
	GpgKeyId        int    `json:"gpg_key_id,omitempty"`
	SslCaCertId     int    `json:"ssl_ca_cert_id,omitempty"`
	SslClientCertId int    `json:"ssl_client_cert_id"`
	SslClientKeyId  int    `json:"ssl_client_key_id"`
	SyncPlanId      int    `json:"sync_plan_id"`
	Label           string `json:"label"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateKatelloProduct creates a new ForemanKatelloProduct with the attributes of the
// supplied ForemanKatelloProduct reference and returns the created
// ForemanKatelloProduct reference. The returned reference will have its ID and
// other API default values set by this function.
func (c *Client) CreateKatelloProduct(ctx context.Context, p *ForemanKatelloProduct) (*ForemanKatelloProduct, error) {
	utils.TraceFunctionCall()

	sJSONBytes, jsonEncErr := c.WrapJSON(nil, p)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("KatelloProductJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		KatelloProductEndpointPrefix,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	// organization_id is a required parameter
	reqQuery := req.URL.Query()
	orgId := strconv.Itoa(c.clientConfig.OrganizationID)
	reqQuery.Set("organization_id", orgId)
	req.URL.RawQuery = reqQuery.Encode()

	var createdKatelloProduct ForemanKatelloProduct
	sendErr := c.SendAndParse(req, &createdKatelloProduct)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdKatelloProduct: [%+v]", createdKatelloProduct)

	return &createdKatelloProduct, nil
}

// ReadKatelloProduct reads the attributes of a ForemanKatelloProduct identified by the
// supplied ID and returns a ForemanKatelloProduct reference.
func (c *Client) ReadKatelloProduct(ctx context.Context, id int) (*ForemanKatelloProduct, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloProductEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readKatelloProduct ForemanKatelloProduct
	sendErr := c.SendAndParse(req, &readKatelloProduct)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readKatelloProduct: [%+v]", readKatelloProduct)

	return &readKatelloProduct, nil
}

// UpdateKatelloProduct updates a ForemanKatelloProduct's attributes.  The sync plan
// with the ID of the supplied ForemanKatelloProduct will be updated. A new
// ForemanKatelloProduct reference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateKatelloProduct(ctx context.Context, p *ForemanKatelloProduct) (*ForemanKatelloProduct, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloProductEndpointPrefix, p.Id)

	sJSONBytes, jsonEncErr := c.WrapJSON(nil, p)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("KatelloProductJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedKatelloProduct ForemanKatelloProduct
	sendErr := c.SendAndParse(req, &updatedKatelloProduct)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedKatelloProduct: [%+v]", updatedKatelloProduct)

	return &updatedKatelloProduct, nil
}

// DeleteKatelloProduct deletes the ForemanKatelloProduct identified by the supplied ID
func (c *Client) DeleteKatelloProduct(ctx context.Context, id int) error {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloProductEndpointPrefix, id)

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

// QueryKatelloProduct queries for a ForemanKatelloProduct based on the attributes of
// the supplied ForemanKatelloProduct reference and returns a QueryResponse struct
// containing query/response metadata and the matching sync plan.
func (c *Client) QueryKatelloProduct(ctx context.Context, p *ForemanKatelloProduct) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		KatelloProductEndpointPrefix,
		nil,
	)
	if reqErr != nil {
		return queryResponse, reqErr
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	name := `"` + p.Name + `"`
	reqQuery.Set("search", "name="+name)

	// organization_id is a required parameter
	orgId := strconv.Itoa(c.clientConfig.OrganizationID)
	reqQuery.Set("organization_id", orgId)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanKatelloProduct for
	// the results
	results := []ForemanKatelloProduct{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanKatelloProduct to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

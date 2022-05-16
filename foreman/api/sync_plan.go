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
	// KatelloSyncPlanEndpointPrefix api endpoint prefix for katello sync_plans
	// 'katello/ will be removed, it's a marker to detect talking with katello api
	// %d will be replaced with organization_id
	KatelloSyncPlanEndpointPrefix = "katello/organizations/%d/sync_plans"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// ForemanKatelloSyncPlan API model representing a sync plan.
// A sync plan is used to schedule a synchronization of a product in katello
type ForemanKatelloSyncPlan struct {
	// Inherits the base object's attributes
	ForemanObject

	// must be one of: hourly, daily, weekly, custom cron.
	Interval string `json:"interval"`
	// start datetime of synchronization
	SyncDate string `json:"sync_date"`
	// sync plan description
	Description string `json:"description"`
	// enables or disables synchronization, Must be one of: true, false, 1, 0.
	Enabled bool `json:"enabled"`
	// custom cron logic for sync plan
	CronExpression string `json:"cron_expression"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateKatelloSyncPlan creates a new ForemanKatelloSyncPlan with the attributes of the
// supplied ForemanKatelloSyncPlan reference and returns the created
// ForemanKatelloSyncPlan reference.  The returned reference will have its ID and
// other API default values set by this function.
func (c *Client) CreateKatelloSyncPlan(ctx context.Context, sp *ForemanKatelloSyncPlan) (*ForemanKatelloSyncPlan, error) {
	log.Tracef("foreman/api/sync_plan.go#Create")

	reqEndpoint := fmt.Sprintf(KatelloSyncPlanEndpointPrefix, c.clientConfig.OrganizationID)

	sJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy(nil, sp)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("KatelloSyncPlanJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdKatelloSyncPlan ForemanKatelloSyncPlan
	sendErr := c.SendAndParse(req, &createdKatelloSyncPlan)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdKatelloSyncPlan: [%+v]", createdKatelloSyncPlan)

	return &createdKatelloSyncPlan, nil
}

// ReadKatelloSyncPlan reads the attributes of a ForemanKatelloSyncPlan identified by the
// supplied ID and returns a ForemanKatelloSyncPlan reference.
func (c *Client) ReadKatelloSyncPlan(ctx context.Context, id int) (*ForemanKatelloSyncPlan, error) {
	log.Tracef("foreman/api/sync_plan.go#Read")

	reqEndpoint := fmt.Sprintf(KatelloSyncPlanEndpointPrefix, c.clientConfig.OrganizationID)
	log.Debugf("readKatelloSyncPlan reqEndpoint: [%+v]", reqEndpoint)
	reqEndpoint = fmt.Sprintf("%s/%d", reqEndpoint, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readKatelloSyncPlan ForemanKatelloSyncPlan
	sendErr := c.SendAndParse(req, &readKatelloSyncPlan)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readKatelloSyncPlan: [%+v]", readKatelloSyncPlan)

	return &readKatelloSyncPlan, nil
}

// UpdateKatelloSyncPlan updates a ForemanKatelloSyncPlan's attributes.  The sync plan
// with the ID of the supplied ForemanKatelloSyncPlan will be updated. A new
// ForemanKatelloSyncPlan reference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateKatelloSyncPlan(ctx context.Context, sp *ForemanKatelloSyncPlan) (*ForemanKatelloSyncPlan, error) {
	log.Tracef("foreman/api/sync_plan.go#Update")

	reqEndpoint := fmt.Sprintf(KatelloSyncPlanEndpointPrefix, c.clientConfig.OrganizationID)
	reqEndpoint = fmt.Sprintf("%s/%d", reqEndpoint, sp.Id)

	sJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy(nil, sp)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("KatelloSyncPlanJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedKatelloSyncPlan ForemanKatelloSyncPlan
	sendErr := c.SendAndParse(req, &updatedKatelloSyncPlan)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedKatelloSyncPlan: [%+v]", updatedKatelloSyncPlan)

	return &updatedKatelloSyncPlan, nil
}

// DeleteKatelloSyncPlan deletes the ForemanKatelloSyncPlan identified by the supplied ID
func (c *Client) DeleteKatelloSyncPlan(ctx context.Context, id int) error {
	log.Tracef("foreman/api/sync_plan.go#Delete")

	reqEndpoint := fmt.Sprintf(KatelloSyncPlanEndpointPrefix, c.clientConfig.OrganizationID)
	reqEndpoint = fmt.Sprintf("%s/%d", reqEndpoint, id)

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

// QueryKatelloSyncPlan queries for a ForemanKatelloSyncPlan based on the attributes of
// the supplied ForemanKatelloSyncPlan reference and returns a QueryResponse struct
// containing query/response metadata and the matching sync plan.
func (c *Client) QueryKatelloSyncPlan(ctx context.Context, sp *ForemanKatelloSyncPlan) (QueryResponse, error) {
	log.Tracef("foreman/api/sync_plan.go#Search")

	reqEndpoint := fmt.Sprintf(KatelloSyncPlanEndpointPrefix, c.clientConfig.OrganizationID)

	queryResponse := QueryResponse{}

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
	name := `"` + sp.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanKatelloSyncPlan for
	// the results
	results := []ForemanKatelloSyncPlan{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanKatelloSyncPlan to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

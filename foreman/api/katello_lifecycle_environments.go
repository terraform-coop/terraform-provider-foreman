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
	LifecycleEnvironmentEndpointPrefix = "/katello/api/environments"
	LifecycleEnvironmentById           = LifecycleEnvironmentEndpointPrefix + "/%d"         // :id
	LifecycleEnvironmentPathsByOrg     = "/katello/api/organizations/%d/environments/paths" // :organization_id
)

type ContentViews struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

type Organization struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Id    int    `json:"id"`
}

type LifecycleEnvironment struct {
	ForemanObject

	Label          string       `json:"label"`
	Description    string       `json:"description"`
	OrganizationId int          `json:"organization_id"`
	Organization   Organization `json:"organization"`

	// Is this LifecycleEnvironment a library?
	Library bool `json:"library"`

	// Container Image Registry related
	RegistryNamePattern         string `json:"registry_name_pattern"`
	RegistryUnauthenticatedPull bool   `json:"registry_unauthenticated_pull"`

	Prior struct {
		Name string `json:"name"`
		Id   int    `json:"id"`
	} `json:"prior"`

	Successor struct {
		Name string `json:"name"`
		Id   int    `json:"id"`
	} `json:"successor"`

	Counts struct {
		ContentHosts int `json:"content_hosts"`
		ContentViews int `json:"content_views"`
	} `json:"counts"`

	ContentViews []ContentViews `json:"content_views"`
}

func (lce *LifecycleEnvironment) MarshalJSON() ([]byte, error) {
	jsonMap := map[string]interface{}{
		"id":              lce.Id,
		"name":            lce.Name,
		"description":     lce.Description,
		"organization_id": lce.OrganizationId,
		"label":           lce.Label,
		"prior_id":        lce.Prior.Id,
	}
	return json.Marshal(jsonMap)
}

func (c *Client) QueryLifecycleEnvironment(ctx context.Context, d *LifecycleEnvironment) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	endpoint := LifecycleEnvironmentEndpointPrefix
	req, err := c.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return queryResponse, err
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	name := `"` + d.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	err = c.SendAndParse(req, &queryResponse)
	if err != nil {
		return queryResponse, err
	}

	utils.Debugf("queryResponse: %+v", queryResponse)

	var results []LifecycleEnvironment
	resultsBytes, err := json.Marshal(queryResponse.Results)
	if err != nil {
		return queryResponse, err
	}
	err = json.Unmarshal(resultsBytes, &results)
	if err != nil {
		return queryResponse, err
	}

	// convert the search results from []ForemanImage to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

func (c *Client) CreateKatelloLifecycleEnvironment(ctx context.Context, lce *LifecycleEnvironment) (*LifecycleEnvironment, error) {
	utils.TraceFunctionCall()

	endpoint := LifecycleEnvironmentEndpointPrefix

	jsonBytes, err := c.WrapJSONWithTaxonomy(nil, lce)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	var createdLce LifecycleEnvironment
	err = c.SendAndParse(req, &createdLce)
	if err != nil {
		return nil, err
	}

	utils.Debugf("createdLce: %+v", createdLce)

	return &createdLce, nil
}

func (c *Client) ReadKatelloLifecycleEnvironment(ctx context.Context, d *LifecycleEnvironment) (*LifecycleEnvironment, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf(LifecycleEnvironmentById, d.Id)
	var lce LifecycleEnvironment

	req, err := c.NewRequestWithContext(ctx, http.MethodGet, reqEndpoint, nil)
	if err != nil {
		return nil, err
	}

	err = c.SendAndParse(req, &lce)
	if err != nil {
		return nil, err
	}

	utils.Debugf("read LifecycleEnv: %+v", lce)

	return &lce, nil
}

func (c *Client) UpdateKatelloLifecycleEnvironment(ctx context.Context, lce *LifecycleEnvironment) (*LifecycleEnvironment, error) {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf(LifecycleEnvironmentById, lce.Id)

	jsonBytes, err := c.WrapJSONWithTaxonomy(nil, lce)
	if err != nil {
		return nil, err
	}

	utils.Debugf("jsonBytes: %s", jsonBytes)

	req, err := c.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	var updatedLce LifecycleEnvironment
	err = c.SendAndParse(req, &updatedLce)
	if err != nil {
		return nil, err
	}

	utils.Debugf("updatedLce: %+v", updatedLce)

	return &updatedLce, nil
}

func (c *Client) DeleteKatelloLifecycleEnvironment(ctx context.Context, id int) error {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf(LifecycleEnvironmentById, id)

	req, err := c.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.SendAndParse(req, nil)
}

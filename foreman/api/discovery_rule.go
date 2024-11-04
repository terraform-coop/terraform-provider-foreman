package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/HanseMerkur/terraform-provider-utils/log"
)

const (
	DiscoveryRuleEndpointPrefix = "/v2/discovery_rules/"
)

type ForemanDiscoveryRule struct {
	ForemanObject
	Name                  string `json:"name"`
	Search                string `json:"search,omitempty"`
	HostGroupId           int    `json:"hostgroup_id,omitempty"`
	Hostname              string `json:"hostname,omitempty"`
	HostsLimitMaxCount    int    `json:"max_count,omitempty"`
	Priority              int    `json:"priority"`
	Enabled               bool   `json:"enabled"`
	LocationIds           []int  `json:"location_ids,omitempty"`
	OrganizationIds       []int  `json:"organization_ids,omitempty"`
	DefaultLocationId     int    `json:"location_id,omitempty"`
	DefaultOrganizationId int    `json:"organization_id,omitempty"`
}

type ForemanDiscoveryRuleResponse struct {
	ForemanObject
	Name               string           `json:"name"`
	Search             string           `json:"search,omitempty"`
	HostGroupId        int              `json:"hostgroup_id,omitempty"`
	Hostname           string           `json:"hostname,omitempty"`
	Priority           int              `json:"priority"`
	Enabled            bool             `json:"enabled"`
	HostsLimitMaxCount int              `json:"hosts_limit,omitempty"`
	Locations          []EntityResponse `json:"locations,omitempty"`
	Organizations      []EntityResponse `json:"organizations,omitempty"`
}

type EntityResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description any    `json:"description"`
}

// CreateDiscoveryRule creates a new ForemanDiscoveryRule
func (c *Client) CreateDiscoveryRule(ctx context.Context, d *ForemanDiscoveryRule) (*ForemanDiscoveryRule, error) {
	log.Tracef("foreman/api/discovery_rule.go#Create")

	if d.DefaultLocationId == 0 {
		d.DefaultLocationId = c.clientConfig.LocationID
	}

	if d.DefaultOrganizationId == 0 {
		d.DefaultOrganizationId = c.clientConfig.OrganizationID
	}

	dJSONBytes, err := c.WrapJSON("discovery_rule", d)
	if err != nil {
		return nil, err
	}

	log.Debugf("discoveryruleJSONBytes: [%s]", dJSONBytes)

	req, err := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		DiscoveryRuleEndpointPrefix,
		bytes.NewBuffer(dJSONBytes),
	)
	if err != nil {
		return nil, err
	}

	var createdDiscoveryRule ForemanDiscoveryRule
	if err := c.SendAndParse(req, &createdDiscoveryRule); err != nil {
		return nil, err
	}

	log.Debugf("createdDiscoveryRule: [%+v]", createdDiscoveryRule)

	return &createdDiscoveryRule, nil
}

// ReadDiscoveryRule reads the ForemanDiscoveryRule identified by the supplied ID
func (c *Client) ReadDiscoveryRule(ctx context.Context, id int) (*ForemanDiscoveryRuleResponse, error) {
	log.Tracef("foreman/api/discovery_rule.go#Read")

	reqEndpoint := path.Join(DiscoveryRuleEndpointPrefix, strconv.Itoa(id))
	req, err := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var readDiscoveryRule ForemanDiscoveryRuleResponse
	if err := c.SendAndParse(req, &readDiscoveryRule); err != nil {
		return nil, err
	}

	log.Debugf("readDiscoveryRule: [%+v]", readDiscoveryRule)

	return &readDiscoveryRule, nil
}

// UpdateDiscoveryRule updates the ForemanDiscoveryRule identified by the supplied ForemanDiscoveryRule
func (c *Client) UpdateDiscoveryRule(ctx context.Context, d *ForemanDiscoveryRule) (*ForemanDiscoveryRule, error) {
	log.Tracef("foreman/api/discovery_rule.go#Update")

	reqEndpoint := path.Join(DiscoveryRuleEndpointPrefix, strconv.Itoa(d.Id))

	discoveryruleJSONBytes, err := c.WrapJSON("discovery_rule", d)
	if err != nil {
		return nil, err
	}

	log.Debugf("discoveryruleJSONBytes: [%s]", discoveryruleJSONBytes)

	req, err := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(discoveryruleJSONBytes),
	)
	if err != nil {
		return nil, err
	}

	var updatedDiscoveryRule ForemanDiscoveryRule

	if err := c.SendAndParse(req, &updatedDiscoveryRule); err != nil {
		return nil, err
	}

	log.Debugf("updatedDiscoveryRule: [%+v]", updatedDiscoveryRule)

	return &updatedDiscoveryRule, nil
}

// DeleteDiscoveryRule deletes the ForemanDiscoveryRule identified by the supplied ID
func (c *Client) DeleteDiscoveryRule(ctx context.Context, id int) error {
	log.Tracef("foreman/api/discovery_rule.go#Delete")

	reqEndpoint := path.Join(DiscoveryRuleEndpointPrefix, strconv.Itoa(id))

	req, err := c.NewRequestWithContext(
		ctx,
		http.MethodDelete,
		reqEndpoint,
		nil,
	)
	if err != nil {
		return err
	}

	return c.SendAndParse(req, nil)
}

// QueryDiscoveryRule queries the ForemanDiscoveryRule identified by the supplied ForemanDiscoveryRule
func (c *Client) QueryDiscoveryRule(ctx context.Context, d *ForemanDiscoveryRule) (QueryResponse, error) {
	log.Tracef("foreman/api/discovery_rule.go#Search")

	queryResponse := QueryResponse{}

	req, err := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		DiscoveryRuleEndpointPrefix,
		nil,
	)
	if err != nil {
		return queryResponse, err
	}

	reqQuery := req.URL.Query()
	reqQuery.Set("search", fmt.Sprintf("name=\"%s\"", d.Name))

	req.URL.RawQuery = reqQuery.Encode()
	if err := c.SendAndParse(req, &queryResponse); err != nil {
		return queryResponse, err
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	results := []ForemanDiscoveryRule{}
	resultsBytes, err := json.Marshal(queryResponse.Results)
	if err != nil {
		return queryResponse, err
	}

	if err := json.Unmarshal(resultsBytes, &results); err != nil {
		return queryResponse, err
	}

	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

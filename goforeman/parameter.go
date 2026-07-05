package goforeman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ParameterParentTypes maps the Terraform-facing "<parent>_id" attribute
// name to the Foreman API's URL path segment for that parent resource type.
// Foreman's plain "parameter" resource has no bare /api/parameters route:
// every method is scoped under exactly one of these parent types (confirmed
// against apidoc/v2.json), which the generic single-ParentEndpoint
// generator mechanism can't express.
var ParameterParentTypes = map[string]string{
	"host_id":            "hosts",
	"hostgroup_id":       "hostgroups",
	"domain_id":          "domains",
	"operatingsystem_id": "operatingsystems",
	"subnet_id":          "subnets",
	"location_id":        "locations",
	"organization_id":    "organizations",
}

type ParameterRequest struct {
	Name          string `json:"name,omitempty"`
	Value         string `json:"value,omitempty"`
	ParameterType string `json:"parameter_type,omitempty"`
	// *bool WITH omitempty, not without: a nil pointer must omit the key
	// entirely (hidden_value is NOT NULL like almost every boolean
	// column), while a non-nil pointer (even to false) still always sends
	// it - see katello_sync_plan.go's Enabled field for the full reasoning.
	HiddenValue *bool `json:"hidden_value,omitempty"`
}

type Parameter struct {
	Base
	// Value is decoded as json.RawMessage, not string: Foreman parameters
	// are user-typed (string/boolean/integer/array/hash/yaml/json), so a
	// non-string value would otherwise fail json.Unmarshal for the whole
	// struct. Rendered via parameterValueToString, matching
	// common_parameters/smart_class_parameters.
	Value         json.RawMessage `json:"value,omitempty"`
	ParameterType string          `json:"parameter_type"`
	HiddenValue   bool            `json:"hidden_value"`
}

func (c *Client) CreateParameter(ctx context.Context, parentType string, parentID int, req *ParameterRequest) (*Parameter, error) {
	var resp Parameter
	err := c.Post(ctx, fmt.Sprintf("%s/%d/parameters", parentType, parentID), "parameter", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ReadParameter(ctx context.Context, parentType string, parentID, id int) (*Parameter, error) {
	var resp Parameter
	err := c.Get(ctx, fmt.Sprintf("%s/%d/parameters/%d", parentType, parentID, id), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateParameter(ctx context.Context, parentType string, parentID, id int, req *ParameterRequest) (*Parameter, error) {
	var resp Parameter
	err := c.Put(ctx, fmt.Sprintf("%s/%d/parameters/%d", parentType, parentID, id), "parameter", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteParameter(ctx context.Context, parentType string, parentID, id int) error {
	return c.Delete(ctx, fmt.Sprintf("%s/%d/parameters/%d", parentType, parentID, id))
}

func (c *Client) FindParameterByName(ctx context.Context, parentType string, parentID int, name string) (*Parameter, error) {
	var response QueryResponse
	err := c.Get(ctx, fmt.Sprintf("%s/%d/parameters?search=name=\"%s\"", parentType, parentID, url.QueryEscape(name)), &response)
	if err != nil {
		return nil, err
	}
	if len(response.Results) == 0 {
		return nil, nil
	}
	var obj Parameter
	if err := json.Unmarshal(response.Results[0], &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

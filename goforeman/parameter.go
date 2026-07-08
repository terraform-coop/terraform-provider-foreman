package goforeman

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
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

// ParameterParentFields lists ParameterParentTypes' keys in a stable order,
// for building candidate lists and error messages.
var ParameterParentFields = []string{
	"host_id", "hostgroup_id", "domain_id", "operatingsystem_id",
	"subnet_id", "location_id", "organization_id",
}

// ResolveParameterParent picks the single parent a parameter is scoped
// under from a field-name-to-ID map of the caller's set values (zero/absent
// entries are ignored). Foreman requires exactly one: every parameter
// endpoint lives under exactly one parent resource, so zero set parents
// means there is no endpoint to call and several set parents is ambiguous -
// both return an error naming the valid fields.
func ResolveParameterParent(setIDs map[string]int64) (parentType string, parentID int64, err error) {
	var found []string
	for _, field := range ParameterParentFields {
		if id := setIDs[field]; id != 0 {
			found = append(found, field)
			parentType, parentID = ParameterParentTypes[field], id
		}
	}
	if len(found) != 1 {
		return "", 0, fmt.Errorf("exactly one of %s must be set; got %d", strings.Join(ParameterParentFields, ", "), len(found))
	}
	return parentType, parentID, nil
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
	// struct. Rendered via RawValueString, matching
	// common_parameters/smart_class_parameters.
	Value         json.RawMessage `json:"value,omitempty"`
	ParameterType string          `json:"parameter_type"`
	// "hidden_value?" (literal question mark), not "hidden_value":
	// parameter-family responses carry the boolean under the ?-suffixed
	// key and reuse the plain key for the masked VALUE string ("*****").
	HiddenValue bool `json:"hidden_value?"`
}

// RawValueString renders a Foreman parameter's raw JSON "value" as plain
// text: a JSON string decodes to its unquoted contents (the common case),
// while a boolean/number/array/object's JSON text is used as-is (e.g.
// "true", "123", ["a","b"]) so any user-typed parameter value
// (parameter_type: string/boolean/integer/real/array/hash/yaml/json)
// round-trips losslessly through a string representation instead of
// failing to decode.
func RawValueString(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// ParseParameters decodes Foreman's standard parameter-collection
// convention - a JSON array of {"name": ..., "value": ...} objects, as
// returned for host_parameters_attributes, group_parameters_attributes,
// os_parameters_attributes, and the like - into a name-to-value map.
// Values are rendered via RawValueString, since each one is user-typed
// and not necessarily a JSON string. A missing/null collection returns
// (nil, nil).
func ParseParameters(raw json.RawMessage) (map[string]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var params []struct {
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(params))
	for _, p := range params {
		out[p.Name] = RawValueString(p.Value)
	}
	return out, nil
}

// BuildParameters converts a name-to-value map into the
// [{"name": ..., "value": ...}] array shape Foreman expects on write for
// its parameter-collection attributes (the inverse of ParseParameters).
// Returns nil for an empty map.
func BuildParameters(params map[string]string) []map[string]interface{} {
	if len(params) == 0 {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(params))
	for name, value := range params {
		out = append(out, map[string]interface{}{
			"name":  name,
			"value": value,
		})
	}
	return out
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

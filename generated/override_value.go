package generated

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const overrideValueEndpointPrefix = "smart_class_parameters/%d/override_values"

// ForemanOverrideValueRequest is the request payload for override values.
type ForemanOverrideValueRequest struct {
	Match                 string `json:"match"`
	Value                 string `json:"value"`
	Omit                  bool   `json:"omit"`
	SmartClassParameterID int    `json:"-"`
}

// ForemanOverrideValue is the entity type for override values.
type ForemanOverrideValue struct {
	ForemanObject
	MatchType             string `json:"-"`
	MatchValue            string `json:"-"`
	Omit                  bool   `json:"omit"`
	SmartClassParameterID int    `json:"-"`
	Value                 string `json:"-"`
}

// MarshalJSON implements custom JSON marshaling for override values.
func (ov ForemanOverrideValue) MarshalJSON() ([]byte, error) {
	ovMap := map[string]interface{}{}
	ovMap["omit"] = ov.Omit
	ovMap["match"] = ov.MatchType + "=" + ov.MatchValue

	// Attempt to parse as int -> float -> bool, fallback to string
	var err error
	ovMap["value"], err = json.Number(ov.Value).Int64()
	if err != nil {
		ovMap["value"], err = json.Number(ov.Value).Float64()
	}
	if err != nil {
		ovMap["value"], err = parseBool(ov.Value)
	}
	if err != nil {
		ovMap["value"] = ov.Value
	}

	return json.Marshal(ovMap)
}

// UnmarshalJSON implements custom JSON unmarshaling for override values.
func (ov *ForemanOverrideValue) UnmarshalJSON(b []byte) error {
	var fo ForemanObject
	if err := json.Unmarshal(b, &fo); err != nil {
		return err
	}
	ov.ForemanObject = fo

	var tmpMap map[string]interface{}
	if err := json.Unmarshal(b, &tmpMap); err != nil {
		return err
	}

	match, _ := tmpMap["match"].(string)
	if strings.HasPrefix(match, "fqdn=") {
		ov.MatchType = "fqdn"
		ov.MatchValue = strings.TrimPrefix(match, "fqdn=")
	} else if strings.HasPrefix(match, "hostgroup=") {
		ov.MatchType = "hostgroup"
		ov.MatchValue = strings.TrimPrefix(match, "hostgroup=")
	} else if strings.HasPrefix(match, "domain=") {
		ov.MatchType = "domain"
		ov.MatchValue = strings.TrimPrefix(match, "domain=")
	} else if strings.HasPrefix(match, "os=") {
		ov.MatchType = "os"
		ov.MatchValue = strings.TrimPrefix(match, "os=")
	}

	ov.Omit, _ = tmpMap["omit"].(bool)

	if v, ok := tmpMap["value"].(string); ok {
		ov.Value = v
	} else {
		vb, _ := json.Marshal(tmpMap["value"])
		ov.Value = string(vb)
	}

	return nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		return false, fmt.Errorf("not a bool: %s", s)
	}
}

// CreateForemanOverrideValue creates a new override value.
func (c *ForemanClient) CreateForemanOverrideValue(ctx context.Context, scpID int, req *ForemanOverrideValueRequest) (*ForemanOverrideValue, error) {
	body := map[string]interface{}{
		"override_value": map[string]interface{}{
			"match": req.Match,
			"value": req.Value,
			"omit":  req.Omit,
		},
	}

	var resp ForemanOverrideValue
	endpoint := fmt.Sprintf(overrideValueEndpointPrefix, scpID)
	if err := c.Post(ctx, endpoint, "", body, &resp); err != nil {
		return nil, err
	}
	resp.SmartClassParameterID = scpID
	return &resp, nil
}

// ReadForemanOverrideValue reads an override value by ID and smart class parameter ID.
func (c *ForemanClient) ReadForemanOverrideValue(ctx context.Context, scpID int, id int) (*ForemanOverrideValue, error) {
	var resp ForemanOverrideValue
	endpoint := fmt.Sprintf(overrideValueEndpointPrefix+"/%d", scpID, id)
	if err := c.Get(ctx, endpoint, &resp); err != nil {
		return nil, err
	}
	resp.SmartClassParameterID = scpID
	return &resp, nil
}

// UpdateForemanOverrideValue updates an override value.
func (c *ForemanClient) UpdateForemanOverrideValue(ctx context.Context, scpID int, id int, req *ForemanOverrideValueRequest) (*ForemanOverrideValue, error) {
	body := map[string]interface{}{
		"override_value": map[string]interface{}{
			"match": req.Match,
			"value": req.Value,
			"omit":  req.Omit,
		},
	}

	var resp ForemanOverrideValue
	endpoint := fmt.Sprintf(overrideValueEndpointPrefix+"/%d", scpID, id)
	if err := c.Put(ctx, endpoint, "", body, &resp); err != nil {
		return nil, err
	}
	resp.SmartClassParameterID = scpID
	return &resp, nil
}

// DeleteForemanOverrideValue deletes an override value.
func (c *ForemanClient) DeleteForemanOverrideValue(ctx context.Context, scpID int, id int) error {
	endpoint := fmt.Sprintf(overrideValueEndpointPrefix+"/%d", scpID, id)
	return c.Delete(ctx, endpoint)
}

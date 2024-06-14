package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"
	"strconv"
	"strings"
)

const (
	OverrideValueEndpointPrefix = "smart_class_parameters/%d/override_values"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanPuppetClass API model represents a Puppet class
type ForemanOverrideValue struct {
	// Inherits the base object's attributes
	ForemanObject

	// The type of match to perform: fqdn, hostgroup, domain or os
	MatchType string
	// The value of requested match
	MatchValue string
	// Wether Foreman omits this parameter from the classification output
	Omit bool `json:"omit"`
	// The ID of the smarts class parameter we are overriding
	SmartClassParameterId int
	// The value of the override - hashes and array must be JSON encoded
	Value string
}

// Implement the Marshaler interface
func (ov ForemanOverrideValue) MarshalJSON() ([]byte, error) {
	utils.TraceFunctionCall()

	ovMap := map[string]interface{}{}
	ovMap["omit"] = ov.Omit
	ovMap["match"] = ov.MatchType + "=" + ov.MatchValue

	// Attempt to parse as int -> float -> bool
	// Accept the first one that succeeds, otherwise assume string
	var err error
	ovMap["value"], err = strconv.Atoi(ov.Value)
	if err != nil {
		ovMap["value"], err = strconv.ParseFloat(ov.Value, 32)
	}
	if err != nil {
		ovMap["value"], err = strconv.ParseBool(ov.Value)
	}
	if err != nil {
		ovMap["value"] = ov.Value
		utils.Debugf("override_value.go #MarshalJSON/passraw")
	}

	utils.Debugf("ovMap: [%+v]", ovMap)

	return json.Marshal(ovMap)
}

// Custom JSON unmarshal function. Unmarshal to the unexported JSON struct
// and then convert over to a ForemanHost struct.
func (ov *ForemanOverrideValue) UnmarshalJSON(b []byte) error {
	utils.TraceFunctionCall()

	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	ov.ForemanObject = fo

	var tmpMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &tmpMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	utils.Debugf("tmpMap: [%v]", tmpMap)

	var ok bool
	var match string
	if match, ok = tmpMap["match"].(string); !ok {
		match = ""
	}

	if strings.HasPrefix(match, "fqdn") {
		ov.MatchType = "fqdn"
		ov.MatchValue = strings.TrimPrefix(match, "fqdn=")
	}
	if strings.HasPrefix(match, "hostgroup") {
		ov.MatchType = "hostgroup"
		ov.MatchValue = strings.TrimPrefix(match, "hostgroup=")
	}
	if strings.HasPrefix(match, "domain") {
		ov.MatchType = "domain"
		ov.MatchValue = strings.TrimPrefix(match, "domain=")
	}
	if strings.HasPrefix(match, "os") {
		ov.MatchType = "os"
		ov.MatchValue = strings.TrimPrefix(match, "os=")
	}

	utils.Debugf("override_value.go #UnmarshalJSON/postMatch")

	if ov.Omit, ok = tmpMap["omit"].(bool); !ok {
		ov.Omit = false
	}

	if ov.Value, ok = tmpMap["value"].(string); !ok {
		vb, _ := json.Marshal(tmpMap["value"])
		ov.Value = string(vb)
	}

	utils.Debugf("override_value.go #UnmarshalJSON/postValue")

	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateOverrideValue creates a new ForemanOverrideValue with the attributes of the supplied
// ForemanOverrideValue reference and returns the created ForemanOverrideValue reference.  The
// returned reference will have its ID and other API default values set by this
// function.
func (c *Client) CreateOverrideValue(ctx context.Context, ov *ForemanOverrideValue) (*ForemanOverrideValue, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf(OverrideValueEndpointPrefix, ov.SmartClassParameterId)

	oJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("override_value", ov)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	utils.Debugf("overrideJSONBytes: [%s]", oJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(oJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdOverrideValue ForemanOverrideValue
	sendErr := c.SendAndParse(req, &createdOverrideValue)
	if sendErr != nil {
		return nil, sendErr
	}

	// Smart class param id is not returned in the respoonse so it must be manually added
	createdOverrideValue.SmartClassParameterId = ov.SmartClassParameterId

	utils.Debugf("createdOverrideValue: [%+v]", createdOverrideValue)
	return &createdOverrideValue, nil

}

// ReadOverrideValue reads the attributes of a ForemanOverrideValue identified by the
// supplied ID & SmartParameterID and returns a ForemanOverrideValue reference.
// NOTE - although override value ids appear to be unique the API requires the smart
// class parameter id as well.
func (c *Client) ReadOverrideValue(ctx context.Context, id int, scp_id int) (*ForemanOverrideValue, error) {
	utils.TraceFunctionCall()

	// Build the API endpoint
	reqEndpoint := fmt.Sprintf(OverrideValueEndpointPrefix+"/%d", scp_id, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readOverrideValue ForemanOverrideValue
	sendErr := c.SendAndParse(req, &readOverrideValue)
	if sendErr != nil {
		return nil, sendErr
	}

	readOverrideValue.SmartClassParameterId = scp_id
	utils.Debugf("readOverrideValue: [%+v]", readOverrideValue)

	return &readOverrideValue, nil
}

// UpdateOverrideValue updates a ForemanOverrideValue's attributes.
func (c *Client) UpdateOverrideValue(ctx context.Context, ov *ForemanOverrideValue) (*ForemanOverrideValue, error) {
	utils.TraceFunctionCall()

	// Build the API endpoint
	reqEndpoint := fmt.Sprintf(OverrideValueEndpointPrefix+"/%d", ov.SmartClassParameterId, ov.Id)

	ovJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("override_value", ov)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	utils.Debugf("OverrideValueJSONBytes: [%s]", ovJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(ovJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedOverrideValue ForemanOverrideValue
	sendErr := c.SendAndParse(req, &updatedOverrideValue)
	if sendErr != nil {
		return nil, sendErr
	}

	// Smart class param id is not returned in the respoonse so it must be manually added
	updatedOverrideValue.SmartClassParameterId = ov.SmartClassParameterId

	utils.Debugf("updatedOverrideValue: [%+v]", updatedOverrideValue)

	return &updatedOverrideValue, nil
}

// DeleteOverideValue deletes the ForemanOverrideValue identified by the supplied ID and smarts class param ID
func (c *Client) DeleteOverrideValue(ctx context.Context, id int, scp_id int) error {
	utils.TraceFunctionCall()

	// Build the API endpoint
	reqEndpoint := fmt.Sprintf(OverrideValueEndpointPrefix+"/%d", scp_id, id)

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

// Doesn't look like this is possible in the API
// The only field it makes sense to search on is match, but this is not supported
// So we cannot have a data object, only resource

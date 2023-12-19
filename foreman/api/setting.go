package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"
)

const (
	SettingEndpointPrefix = "settings"
)

// Represents a setting object in Foreman. Settings use strings as Id values.
type ForemanSetting struct {
	ForemanObject

	// Settings use strings as IDs
	// Overrides the ForemanObject.Id field
	Id string `json:"id"`

	// full_name field which seems to be empty
	Fullname string `json:"full_name,omitempty"`

	// The value of the setting, can be bool, string or int
	Value interface{} `json:"value"`

	// DO NOT USE: In case the value is a boolean, it could be stored in here
	// Go does not allow setting this field to nil, which would be the correct
	// value for Values that are strings. Unexpected behaviour - do not use!
	// ValueBool bool

	// Default value as bool, string or int
	Default interface{} `json:"default"`

	// Is setting read-only?
	ReadOnly bool `json:"readonly"`

	// Is setting encrypted?
	Encrypted bool `json:"encrypted"`

	// Description of the setting
	Description string `json:"description"`

	// Category of the setting in colon form, e.g. "Setting::Auth"
	Category string `json:"category"`

	// Category of the setting in human readable form, e.g. "Authentication"
	CategoryName string `json:"category_name"`

	// Type of the setting (boolean, string etc.)
	SettingsType string `json:"settings_type"`
}

// ReadSetting reads the attributes of a ForemanSetting identified by the supplied
// ID and returns a ForemanSetting reference.
func (c *Client) ReadSetting(ctx context.Context, id string) (*ForemanSetting, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%s", SettingEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readSetting ForemanSetting
	sendErr := c.SendAndParse(req, &readSetting)
	if sendErr != nil {
		return nil, sendErr
	}

	utils.Debugf("readSetting: [%+v]", readSetting)

	return &readSetting, nil
}

// QuerySetting queries for a ForemanSetting based on the attributes of the
// supplied ForemanSetting reference and returns a QueryResponse struct
// containing query/response metadata and the matching settings.
// TODO: Copied from QueryDomains.
func (c *Client) QuerySetting(ctx context.Context, d *ForemanSetting) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", SettingEndpointPrefix)
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
	name := `"` + d.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	utils.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanSetting for
	// the results
	results := []ForemanSetting{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanSetting to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

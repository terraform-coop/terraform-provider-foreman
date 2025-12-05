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
	WebhookEndpointPrefix = "webhooks"
)

// The ForemanWebhook API model represents a webhook
type ForemanWebhook struct {
	ForemanObject
	TargetURL          string `json:"target_url"`
	HTTPMethod         string `json:"http_method"`
	HTTPContentType    string `json:"http_content_type"`
	HTTPHeaders        string `json:"http_headers"`
	Event              string `json:"event"`
	Enabled            bool   `json:"enabled"`
	VerifySSL          bool   `json:"verify_ssl"`
	SSLCACerts         string `json:"ssl_ca_certs"`
	ProxyAuthorization bool   `json:"proxy_authorization"`
	User               string `json:"user"`
	Password           string `json:"password"`
	WebhookTemplateID  int    `json:"webhook_template_id"`
}

type ForemanWebhookResponse struct {
	ForemanObject
	TargetURL          string          `json:"target_url"`
	HTTPMethod         string          `json:"http_method"`
	HTTPContentType    string          `json:"http_content_type"`
	HTTPHeaders        string          `json:"http_headers"`
	Event              string          `json:"event"`
	Enabled            bool            `json:"enabled"`
	VerifySSL          bool            `json:"verify_ssl"`
	SSLCACerts         string          `json:"ssl_ca_certs"`
	ProxyAuthorization bool            `json:"proxy_authorization"`
	User               string          `json:"user"`
	PasswordSet        bool            `json:"password_set"`
	WebhookTemplate    WebhookTemplate `json:"webhook_template"`
}

type WebhookTemplate struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

// Custom JSON marshal function for webhooks. The Foreman API
// expects all parameters to be enclosed in double quotes, with the exception
// of boolean and slice values.
func (fw ForemanWebhook) MarshalJSON() ([]byte, error) {
	log.Tracef("Webhook marshal")

	// map structure representation of the passed ForemanWebhook
	// for ease of marshalling - essentially convert over to a map then call
	// json.Marshal() on the mapstructure
	fwMap := map[string]interface{}{}

	fwMap["name"] = fw.Name
	fwMap["target_url"] = fw.TargetURL
	fwMap["http_method"] = fw.HTTPMethod
	fwMap["http_content_type"] = fw.HTTPContentType
	fwMap["http_headers"] = fw.HTTPHeaders
	fwMap["event"] = fw.Event
	fwMap["enabled"] = fw.Enabled
	fwMap["verify_ssl"] = fw.VerifySSL
	fwMap["ssl_ca_certs"] = fw.SSLCACerts
	fwMap["proxy_authorization"] = fw.ProxyAuthorization
	fwMap["user"] = fw.User
	fwMap["password"] = fw.Password
	fwMap["webhook_template_id"] = intIdToJSONString(fw.WebhookTemplateID)

	log.Debugf("fwMap: [%v]", fwMap)

	return json.Marshal(fwMap)
}

// Custom JSON unmarshal function. Unmarshal to the unexported JSON struct
// and then convert over to a ForemanWebhook struct.
func (fw *ForemanWebhook) UnmarshalJSON(b []byte) error {
	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fw.ForemanObject = fo

	// Unmarshal into mapstructure and set the rest of the struct properties
	var fwMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &fwMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	var ok bool
	if fw.TargetURL, ok = fwMap["target_url"].(string); !ok {
		fw.TargetURL = ""
	}
	if fw.HTTPMethod, ok = fwMap["http_method"].(string); !ok {
		fw.HTTPMethod = ""
	}
	if fw.HTTPContentType, ok = fwMap["http_content_type"].(string); !ok {
		fw.HTTPContentType = ""
	}
	if fw.HTTPHeaders, ok = fwMap["http_headers"].(string); !ok {
		fw.HTTPHeaders = ""
	}
	if fw.Event, ok = fwMap["event"].(string); !ok {
		fw.Event = ""
	}
	if fw.Enabled, ok = fwMap["enabled"].(bool); !ok {
		fw.Enabled = false
	}
	if fw.VerifySSL, ok = fwMap["verify_ssl"].(bool); !ok {
		fw.VerifySSL = false
	}
	if fw.SSLCACerts, ok = fwMap["ssl_ca_certs"].(string); !ok {
		fw.SSLCACerts = ""
	}
	if fw.ProxyAuthorization, ok = fwMap["proxy_authorization"].(bool); !ok {
		fw.ProxyAuthorization = false
	}
	if fw.User, ok = fwMap["user"].(string); !ok {
		fw.User = ""
	}
	if fw.Password, ok = fwMap["password"].(string); !ok {
		fw.Password = ""
	}
	fw.WebhookTemplateID = unmarshalInteger(fwMap["webhook_template_id"])

	return nil
}

// CreateWebhook creates a new ForemanWebhook with the attributes
// of the supplied ForemanWebhook reference and returns the created
// ForemanWebhook reference. The returned reference will have its ID
// and other API default values set by this function.
func (c *Client) CreateWebhook(ctx context.Context, w *ForemanWebhook) (*ForemanWebhook, error) {
	log.Tracef("foreman/api/webhook.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", WebhookEndpointPrefix)

	wJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("webhook", w)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("webhookJSONBytes: [%s]", wJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(wJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdWebhook ForemanWebhook
	sendErr := c.SendAndParse(req, &createdWebhook)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdWebhook: [%+v]", createdWebhook)

	return &createdWebhook, nil
}

// ReadWebhook reads the attributes of a ForemanWebhook identified
// by the supplied ID and returns a ForemanWebhook reference.
func (c *Client) ReadWebhook(ctx context.Context, id int) (*ForemanWebhookResponse, error) {
	log.Tracef("foreman/api/webhook.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", WebhookEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readWebhook ForemanWebhookResponse
	sendErr := c.SendAndParse(req, &readWebhook)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readWebhook: [%+v]", readWebhook)

	return &readWebhook, nil
}

// UpdateWebhook updates a ForemanWebhook's attributes. The webhook with
// the ID of the supplied ForemanWebhook  will be updated. A new
// ForemanWebhookreference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateWebhook(ctx context.Context, w *ForemanWebhook) (*ForemanWebhook, error) {
	log.Tracef("foreman/api/webhook.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", WebhookEndpointPrefix, w.Id)
	wJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("webhook", w)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("webhookJSONBytes: [%s]", wJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(wJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedWebhook ForemanWebhook
	sendErr := c.SendAndParse(req, &updatedWebhook)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedWebhook: [%+v]", updatedWebhook)

	return &updatedWebhook, nil
}

// DeleteWebhook deletes the ForemanWebhook identified by the supplied ID
func (c *Client) DeleteWebhook(ctx context.Context, id int) error {
	log.Tracef("foreman/api/webhook.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", WebhookEndpointPrefix, id)

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

// QueryWebhook queries for a ForemanWebhook based on the attributes of
// the supplied ForemanWebhook reference and returns a QueryResponse struct
// containing query/response metadata and the matching templates.
func (c *Client) QueryWebhook(ctx context.Context, t *ForemanWebhook) (QueryResponse, error) {
	log.Tracef("foreman/api/webhook.go#Query")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", WebhookEndpointPrefix)
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
	name := "\"" + t.Name + "\""
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanWebhook for
	// the results
	results := []ForemanWebhook{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanWebhook to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	ProvisioningTemplateEndpointPrefix = "provisioning_templates"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanProvisioningTemplate API model represents a provisioning template.
// Provisioning templates are scripts used to describe how to boostrap and
// install the operating system on a host.
type ForemanProvisioningTemplate struct {
	// Inherits the base object's attributes
	ForemanObject

	// The markup and code of the provisioning template
	Template string
	// Whether or not the provisioning template is a snippet to be embedded
	// and used by other templates
	Snippet bool
	// Notes and comments for auditing purposes
	AuditComment string
	// Whether or not the template is locked for editing
	Locked bool
	// ID of the template kind which categorizes the provisioning template.
	// Optional for snippets, otherwise required.
	TemplateKindId int
	// IDs of operating systems associated with this provisioning template
	OperatingSystemIds []int
	// How templates are determined:
	//
	// When editing a template, you must assign a list of operating systems
	// which this template can be used with.  Optionally, you can restrict
	// a template to a list of host groups and/or environments.
	//
	// When a host requests a template, Foreman will select the best match
	// from the available templates of that type in the following order:
	//
	//   1. host group and environment
	//   2. host group only
	//   3. environment only
	//   4. operating system default
	//
	// Template combinations attributes contains an array of hostgroup IDs
	// and environment ID combinations so they can be used in the
	// provisioning template selection described above.
	TemplateCombinationsAttributes []ForemanTemplateCombinationAttribute `json:"template_combinations_attributes"`
}

// See the comment in ForemanProvisioningTemplate.TemplateCombinationsAttributes
type ForemanTemplateCombinationAttribute struct {
	// Unique identifier of the template combination
	Id int `json:"id,omitempty"`
	// Hostgroup ID associated with the template combination
	HostgroupId int `json:"hostgroup_id,omitempty"`
	// Environment ID associated with the template combination
	EnvironmentId int `json:"environment_id,omitempty"`
	// NOTE(ALL): Each of the template combinations receives a unique identifier
	//   on creation. To modify the list of template combinations, the supplied
	//   list to the API does NOT perform a replace operation. Adding new
	//   combinations to the list is rather trivial and just involves sending the
	//   new values to receive an ID.  When removing one of the combinations from
	//   the set, a secret flag "_destroy" must be supplied as part of that
	//   combination.  This is not documented as part of the Foreman API.  We
	//   omit empty here, because we only want to pass the flag when "_destroy"
	//   is "true" to signal an item removal.
	Destroy bool `json:"_destroy,omitempty"`
}

// ForemanProvisioningTemplate struct used for JSON decode.  Foreman API returns
// the operating system ids back as a list of ForemanObjects with some of the
// attributes.  For our purposes, we are only interested in the IDs.
type foremanProvisioningTemplateJSON struct {
	OperatingSystems               []ForemanObject                       `json:"operatingsystems"`
	TemplateCombinationsAttributes []ForemanTemplateCombinationAttribute `json:"template_combinations"`
}

// Custom JSON marshal function for provisioning temmplates.  The Foreman API
// expects all parameters to be enclosed in double quotes, with the exception
// of boolean and slice values.
func (ft ForemanProvisioningTemplate) MarshalJSON() ([]byte, error) {
	log.Tracef("Provisioning template marshal")

	// map structure representation of the passed ForemanProvisioningTemplate
	// for ease of marshalling - essentially convert over to a map then call
	// json.Marshal() on the mapstructure
	ftMap := map[string]interface{}{}

	ftMap["name"] = ft.Name
	ftMap["template"] = ft.Template
	ftMap["snippet"] = ft.Snippet
	ftMap["audit_comment"] = ft.AuditComment
	ftMap["locked"] = ft.Locked
	ftMap["template_kind_id"] = intIdToJSONString(ft.TemplateKindId)

	// always marshal the OS array - otherwise, when the array is updated
	// from [1,2,3] to [], we would skip the marshalling and the OS id array
	// would not get updated by the API
	//
	// Foreman API interprets the data of this field as a REPLACE operation
	ftMap["operatingsystem_ids"] = ft.OperatingSystemIds

	// only include the template combination attributes if it is set.
	// The Foreman API will return "500: Internal Server Error" with the
	// explanation "Expected Hash or Array got NilClass (nil)" if any of the
	// following are supplied: [], null, "null"
	if len(ft.TemplateCombinationsAttributes) > 0 {
		ftMap["template_combinations_attributes"] = ft.TemplateCombinationsAttributes
	}

	log.Debugf("ftMap: [%v]", ftMap)

	return json.Marshal(ftMap)
}

// Custom JSON unmarshal function. Unmarshal to the unexported JSON struct
// and then convert over to a ForemanProvisioningTemplate struct.
func (ft *ForemanProvisioningTemplate) UnmarshalJSON(b []byte) error {
	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	ft.ForemanObject = fo

	// Unmarshal to temporary JSON struct to get the properties with differently
	// named keys
	var ftJSON foremanProvisioningTemplateJSON
	jsonDecErr = json.Unmarshal(b, &ftJSON)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	ft.OperatingSystemIds = foremanObjectArrayToIdIntArray(ftJSON.OperatingSystems)
	ft.TemplateCombinationsAttributes = ftJSON.TemplateCombinationsAttributes

	// Unmarshal into mapstructure and set the rest of the struct properties
	var ftMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &ftMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	var ok bool
	if ft.Template, ok = ftMap["template"].(string); !ok {
		ft.Template = ""
	}
	if ft.Snippet, ok = ftMap["snippet"].(bool); !ok {
		ft.Snippet = false
	}
	if ft.AuditComment, ok = ftMap["audit_comment"].(string); !ok {
		ft.AuditComment = ""
	}
	if ft.Locked, ok = ftMap["locked"].(bool); !ok {
		ft.Locked = false
	}
	if ft.TemplateKindId, ok = ftMap["template_kind_id"].(int); !ok {
		ft.TemplateKindId = 0
	}

	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateProvisioningTemplate creates a new ForemanProvisioningTemplate with
// the attributes of the supplied ForemanProvisioningTemplate reference and
// returns the created ForemanProvisioningTemplate reference.  The returned
// reference will have its ID and other API default values set by this
// function.
func (c *Client) CreateProvisioningTemplate(t *ForemanProvisioningTemplate) (*ForemanProvisioningTemplate, error) {
	log.Tracef("foreman/api/provisioningtemplate.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", ProvisioningTemplateEndpointPrefix)

	tJSONBytes, jsonEncErr := json.Marshal(t)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("templateJSONBytes: [%s]", tJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(tJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdTemplate ForemanProvisioningTemplate
	sendErr := c.SendAndParse(req, &createdTemplate)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdTemplate: [%+v]", createdTemplate)

	return &createdTemplate, nil
}

// ReadProvisioningTemplate reads the attributes of a
// ForemanProvisioningTemplate identified by the supplied ID and returns a
// ForemanProvisioningTemplate reference.
func (c *Client) ReadProvisioningTemplate(id int) (*ForemanProvisioningTemplate, error) {
	log.Tracef("foreman/api/provisioningtemplate.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", ProvisioningTemplateEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readTemplate ForemanProvisioningTemplate
	sendErr := c.SendAndParse(req, &readTemplate)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readTemplate: [%+v]", readTemplate)

	return &readTemplate, nil
}

// UpdateProvisioningTemplate updates a ForemanProvisioningTemplate's
// attributes.  The template with the ID of the supplied
// ForemanProvisioningTemplate will be updated. A new
// ForemanProvisioningTemplate reference is returned with the attributes from
// the result of the update operation.
func (c *Client) UpdateProvisioningTemplate(t *ForemanProvisioningTemplate) (*ForemanProvisioningTemplate, error) {
	log.Tracef("foreman/api/provisioningtemplate.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", ProvisioningTemplateEndpointPrefix, t.Id)

	tJSONBytes, jsonEncErr := json.Marshal(t)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("templateJSONBytes: [%s]", tJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(tJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedTemplate ForemanProvisioningTemplate
	sendErr := c.SendAndParse(req, &updatedTemplate)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedTemplate: [%+v]", updatedTemplate)

	return &updatedTemplate, nil
}

// DeleteProvisioningTemplate deletes the ForemanProvisioningTemplate
// identified by the supplied ID
func (c *Client) DeleteProvisioningTemplate(id int) error {
	log.Tracef("foreman/api/provisioningtemplate.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", ProvisioningTemplateEndpointPrefix, id)

	req, reqErr := c.NewRequest(
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

// QueryProvisioningTemplate queries for a ForemanProvisioningTemplate based on
// the attributes of the supplied ForemanProvisioningTemplate reference and
// returns a QueryResponse struct containing query/response metadata and the
// matching templates.
func (c *Client) QueryProvisioningTemplate(t *ForemanProvisioningTemplate) (QueryResponse, error) {
	log.Tracef("foreman/api/provisioningtemplate.go#Query")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", ProvisioningTemplateEndpointPrefix)
	req, reqErr := c.NewRequest(
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
	// Encode back to JSON, then Unmarshal into []ForemanProvisioningTemplate for
	// the results
	results := []ForemanProvisioningTemplate{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanProvisioningTemplate to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

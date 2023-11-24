package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
)

const (
	// second parameter is the template_id
	TemplateInputEndpointPrefix string = "templates/%d/template_inputs"
)

type ForemanTemplateInput struct {
	ForemanObject

	TemplateId          int    `json:"template_id"`
	FactName            string `json:"fact_name"`
	VariableName        string `json:"variable_name"`
	PuppetParameterName string `json:"puppet_parameter_name"`
	PuppetClassName     string `json:"puppet_class_name"`
	Description         string `json:"description"`
	Required            bool   `json:"required"`
	Advanced            bool   `json:"advanced"`
	Default             string `json:"default"`

	// The value contains sensitive information and shouldn not be normally visible, useful e.g. for passwords
	HiddenValue bool `json:"hidden_value"`

	// Input type. Must be one of: user, fact, variable.
	InputType string `json:"input_type"`

	// Value type, defaults to plain. Must be one of: plain, search, date, resource.
	ValueType string `json:"value_type"`

	// For values of type search, this is the resource the value searches in Validations:
	// Must be one of: Architecture, Audit, AuthSource, Bookmark, ComputeProfile, ComputeResource, ConfigReport, Domain, ExternalUsergroup,
	// FactValue, Filter, ForemanTasks::RecurringLogic, ForemanTasks::Task, Host, Hostgroup, HttpProxy, Image, JobInvocation, JobTemplate,
	// Katello::ActivationKey, Katello::AlternateContentSource, Katello::ContentCredential, Katello::ContentView, Katello::HostCollection,
	// Katello::KTEnvironment, Katello::Product, Katello::Subscription, Katello::SyncPlan, KeyPair, Location, MailNotification, Medium, Model,
	// Operatingsystem, Organization, Parameter, PersonalAccessToken, ProvisioningTemplate, Ptable, Realm, RemoteExecutionFeature, ReportTemplate,
	// Role, Setting, SmartProxy, SshKey, Subnet, TemplateInvocation, User, Usergroup.
	ResourceType string `json:"resource_type"`
}

// Converts the struct fields to a map[string]interface as input into Terraform resource deserialization.
// Needed, because the nested "template_inputs" field in "job_template" uses JSON marshalling to read the attributes into the Terraform-internal object.
func (f *ForemanTemplateInput) ToResourceDataMap() map[string]interface{} {
	attrMap := make(map[string]interface{})

	// Setting "id" is not supported

	attrMap["name"] = f.Name
	attrMap["description"] = f.Description
	attrMap["template_id"] = f.TemplateId
	attrMap["fact_name"] = f.FactName
	attrMap["variable_name"] = f.VariableName
	attrMap["puppet_parameter_name"] = f.PuppetParameterName
	attrMap["puppet_class_name"] = f.PuppetClassName
	attrMap["required"] = f.Required
	attrMap["advanced"] = f.Advanced
	attrMap["default"] = f.Default
	attrMap["hidden_value"] = f.HiddenValue
	attrMap["input_type"] = f.InputType
	attrMap["value_type"] = f.ValueType
	attrMap["resource_type"] = f.ResourceType

	return attrMap
}

/// CRUD

func (c *Client) CreateTemplateInput(ctx context.Context, tiObj *ForemanTemplateInput) (*ForemanTemplateInput, error) {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf("/"+TemplateInputEndpointPrefix, tiObj.TemplateId)

	// No WrapJSONWithTaxonomy here, adding location and organization is not accepted by the API for POST to /api/templates/-tid-/template_inputs
	to_wrap := map[string]interface{}{
		"template_input": tiObj,
	}
	wrapped, err := json.Marshal(to_wrap)
	if err != nil {
		return nil, err
	}

	utils.Debug("template_input JSON: \n%s", wrapped)

	req, err := c.NewRequestWithContext(
		ctx, http.MethodPost, endpoint, bytes.NewBuffer(wrapped),
	)
	if err != nil {
		return nil, err
	}

	var created ForemanTemplateInput
	err = c.SendAndParse(req, &created)
	if err != nil {
		return nil, err
	}

	utils.Debug("Created TemplateInput: %+v", created)

	return &created, nil
}

func (c *Client) QueryTemplateInput(ctx context.Context, tiObj *ForemanTemplateInput) (QueryResponse, error) {
	utils.TraceFunctionCall()

	qresp := QueryResponse{}
	endpoint := fmt.Sprintf("/"+TemplateInputEndpointPrefix, tiObj.TemplateId)

	req, err := c.NewRequestWithContext(
		ctx, http.MethodGet, endpoint, nil,
	)
	if err != nil {
		return qresp, err
	}

	reqQuery := req.URL.Query()
	name := `"` + tiObj.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	err = c.SendAndParse(req, &qresp)
	if err != nil {
		return qresp, err
	}

	results := []ForemanTemplateInput{}
	resultsBytes, err := json.Marshal(qresp.Results)
	if err != nil {
		return qresp, err
	}

	err = json.Unmarshal(resultsBytes, &results)
	if err != nil {
		return qresp, err
	}

	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	qresp.Results = iArr

	return qresp, nil
}

func (c *Client) ReadTemplateInput(ctx context.Context, tiObj *ForemanTemplateInput) (*ForemanTemplateInput, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/"+TemplateInputEndpointPrefix+"/%d", tiObj.TemplateId, tiObj.Id)

	req, err := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if err != nil {
		return nil, err
	}

	var readObj ForemanTemplateInput
	err = c.SendAndParse(req, &readObj)
	if err != nil {
		return nil, err
	}

	log.Debugf("readObj: [%+v]", readObj)

	return &readObj, nil
}

func (c *Client) UpdateTemplateInput(ctx context.Context, tiObj *ForemanTemplateInput) (*ForemanTemplateInput, error) {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf("/"+TemplateInputEndpointPrefix+"/%d", tiObj.TemplateId, tiObj.Id)

	// No WrapJSONWithTaxonomy here, adding location and organization is not accepted by the API for PUT to /api/templates/-tid-/template_inputs
	to_wrap := map[string]interface{}{
		"template_input": tiObj,
	}
	wrapped, err := json.Marshal(to_wrap)
	if err != nil {
		return nil, err
	}

	utils.Debug("template_input JSON: \n%s", wrapped)

	req, err := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		endpoint,
		bytes.NewBuffer(wrapped),
	)
	if err != nil {
		return nil, err
	}

	var updatedObj ForemanTemplateInput
	err = c.SendAndParse(req, &updatedObj)
	if err != nil {
		return nil, err
	}

	return &updatedObj, nil
}

func (c *Client) DeleteTemplateInput(ctx context.Context, tiObj *ForemanTemplateInput) error {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf("/"+TemplateInputEndpointPrefix+"/%d", tiObj.TemplateId, tiObj.Id)
	req, err := c.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.SendAndParse(req, nil)
}

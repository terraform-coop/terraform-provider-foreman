package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
)

const (
	// parameter is the template_id
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
	ValueType string `json:"value_type,omitempty"`

	// For values of type search, this is the resource the value searches in Validations:
	// Must be one of: Architecture, Audit, AuthSource, Bookmark, ComputeProfile, ComputeResource, ConfigReport, Domain, ExternalUsergroup,
	// FactValue, Filter, ForemanTasks::RecurringLogic, ForemanTasks::Task, Host, Hostgroup, HttpProxy, Image, JobInvocation, JobTemplate,
	// Katello::ActivationKey, Katello::AlternateContentSource, Katello::ContentCredential, Katello::ContentView, Katello::HostCollection,
	// Katello::KTEnvironment, Katello::Product, Katello::Subscription, Katello::SyncPlan, KeyPair, Location, MailNotification, Medium, Model,
	// Operatingsystem, Organization, Parameter, PersonalAccessToken, ProvisioningTemplate, Ptable, Realm, RemoteExecutionFeature, ReportTemplate,
	// Role, Setting, SmartProxy, SshKey, Subnet, TemplateInvocation, User, Usergroup.
	ResourceType string `json:"resource_type"`
}

func (fti *ForemanTemplateInput) UnmarshalJSON(b []byte) error {
	utils.TraceFunctionCall()

	var exists bool

	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	// Special handling for Id
	if val, exists := m["id"]; exists {
		switch v := val.(type) {
		case int:
			fti.Id = v
		case float32:
			fti.Id = int(v)
		case float64:
			fti.Id = int(v)
		case string:
			utils.Debugf("ForemanTemplateInput val is string")
			if len(v) == 0 {
				// If empty string, no Id is present
				break
			}
			// Else, convert from string to int
			id, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			// And set in struct
			fti.Id = id
		default:
			utils.Fatalf("Unhandled 'id' type %T", v)
		}
	} else {
		utils.Fatalf("id not in ForemanTemplateInput JSON!")
	}

	// Same for TemplateId
	if val, exists := m["template_id"]; exists {
		switch v := val.(type) {
		case int:
			fti.TemplateId = v
		case float32:
			fti.TemplateId = int(v)
		case float64:
			fti.TemplateId = int(v)
		case string:
			if len(v) == 0 {
				break
			}
			id, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			// And set in struct
			fti.TemplateId = id
		default:
			utils.Fatalf("Unhandled 'template_id' type %T", v)
		}
	}

	// Then unmarshal the rest

	// Foreman object embedded
	if fti.Name, exists = m["name"].(string); !exists {
		fti.Name = ""
	}
	if fti.CreatedAt, exists = m["created_at"].(string); !exists {
		fti.CreatedAt = ""
	}
	if fti.UpdatedAt, exists = m["updated_at"].(string); !exists {
		fti.UpdatedAt = ""
	}

	// ForemanTemplateInput
	if fti.FactName, exists = m["fact_name"].(string); !exists {
		fti.FactName = ""
	}
	if fti.VariableName, exists = m["variable_name"].(string); !exists {
		fti.VariableName = ""
	}
	if fti.PuppetParameterName, exists = m["puppet_parameter_name"].(string); !exists {
		fti.PuppetParameterName = ""
	}
	if fti.PuppetClassName, exists = m["puppet_class_name"].(string); !exists {
		fti.PuppetClassName = ""
	}
	if fti.Description, exists = m["description"].(string); !exists {
		fti.Description = ""
	}
	if fti.Required, exists = m["required"].(bool); !exists {
		fti.Required = false
	}
	if fti.Advanced, exists = m["advanced"].(bool); !exists {
		fti.Advanced = false
	}
	if fti.Default, exists = m["default"].(string); !exists {
		fti.Default = ""
	}
	if fti.HiddenValue, exists = m["hidden_value"].(bool); !exists {
		fti.HiddenValue = false
	}
	if fti.InputType, exists = m["input_type"].(string); !exists {
		fti.InputType = ""
	}
	if fti.ValueType, exists = m["value_type"].(string); !exists {
		fti.ValueType = ""
	}
	if fti.ResourceType, exists = m["resource_type"].(string); !exists {
		fti.ResourceType = ""
	}

	return nil
}

// Converts the struct fields to a map[string]interface as input into Terraform resource deserialization.
// Needed, because the nested "template_inputs" field in "job_template" uses JSON marshalling to read the attributes into the Terraform-internal object.
func (f *ForemanTemplateInput) ToResourceDataMap(includeId bool) map[string]interface{} {
	utils.TraceFunctionCall()

	attrMap := make(map[string]interface{})

	// Differentiate cases for including ID or not. Creating the object should not include the ID parameter,
	// because it will cause "PG::UniqueViolation: ERROR:  duplicate key value violates unique constraint" if multiple
	// template inputs are defined. Reason being, that the provider will try to set id=0.
	// Foreman automatically creates unique IDs if 'id' is not passed as parameter (they will be non-zero).
	if includeId {
		attrMap["id"] = strconv.Itoa(f.Id)
	}

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
		// Use ToResourceDataMap to remove id, created_at and updated_at
		"template_input": tiObj.ToResourceDataMap(false),
	}
	wrapped, err := json.Marshal(to_wrap)
	if err != nil {
		return nil, err
	}

	utils.Debugf("template_input JSON: \n%s", wrapped)

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

	utils.Debugf("Created TemplateInput: %+v", created)

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

	return &readObj, nil
}

func (c *Client) UpdateTemplateInput(ctx context.Context, tiObj *ForemanTemplateInput) (*ForemanTemplateInput, error) {
	utils.TraceFunctionCall()

	utils.Debugf("%+v", tiObj)

	endpoint := fmt.Sprintf("/"+TemplateInputEndpointPrefix+"/%d", tiObj.TemplateId, tiObj.Id)

	// No WrapJSONWithTaxonomy here, adding location and organization is not accepted by the API for PUT to /api/templates/-tid-/template_inputs
	to_wrap := map[string]interface{}{
		"template_input": tiObj,
	}
	wrapped, err := json.Marshal(to_wrap)
	if err != nil {
		return nil, err
	}

	// TODO: handle `tiObj.ValueType == ""` here if omitempty fails

	utils.Debugf("template_input JSON: \n%s", wrapped)

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

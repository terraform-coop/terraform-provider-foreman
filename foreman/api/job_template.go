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
	JobTemplateEndpointPrefix string = "job_templates"
)

type ForemanJobTemplate struct {
	ForemanObject

	Description       string                 `json:"description"`
	DescriptionFormat string                 `json:"description_format"`
	Template          string                 `json:"template"`
	Locked            bool                   `json:"locked"`
	JobCategory       string                 `json:"job_category"`
	ProviderType      string                 `json:"provider_type"`
	Snippet           bool                   `json:"snippet"`
	TemplateInputs    []ForemanTemplateInput `json:"template_inputs"`
	EffectiveUser     interface{}            `json:"effective_user"`

	Locations     []interface{} `json:"locations"`
	Organizations []interface{} `json:"organizations"`
}

/// CRUD

func (c *Client) CreateJobTemplate(ctx context.Context, jtObj *ForemanJobTemplate) (*ForemanJobTemplate, error) {
	utils.TraceFunctionCall()

	const endpoint = "/" + JobTemplateEndpointPrefix

	wrapped, err := c.WrapJSONWithTaxonomy("job_template", jtObj)
	if err != nil {
		return nil, err
	}

	log.Debugf("jtObj: %+v", jtObj)

	utils.Debug("job_template JSON: \n%s", wrapped)

	req, err := c.NewRequestWithContext(
		ctx, http.MethodPost, endpoint, bytes.NewBuffer(wrapped),
	)
	if err != nil {
		return nil, err
	}

	var createdJT ForemanJobTemplate
	err = c.SendAndParse(req, &createdJT)
	if err != nil {
		return nil, err
	}

	// Handle TemplateInputs

	count_ti := len(jtObj.TemplateInputs)
	if count_ti > 0 {
		template_id := createdJT.Id
		created_inputs := make([]ForemanTemplateInput, count_ti)

		for idx, item := range jtObj.TemplateInputs {
			item.TemplateId = template_id

			utils.Debug("Creating TemplateInput: %+v", item)

			ti, err := c.CreateTemplateInput(ctx, &item)
			if err != nil {
				return nil, err
			}

			created_inputs[idx] = *ti
		}

		createdJT.TemplateInputs = created_inputs
	}

	utils.Debug("CreatedJT: %#v", createdJT)

	return &createdJT, nil
}

func (c *Client) QueryJobTemplate(ctx context.Context, jt *ForemanJobTemplate) (QueryResponse, error) {
	utils.TraceFunctionCall()

	qresp := QueryResponse{}
	const endpoint = "/" + JobTemplateEndpointPrefix

	req, err := c.NewRequestWithContext(
		ctx, http.MethodGet, endpoint, nil,
	)
	if err != nil {
		return qresp, err
	}

	reqQuery := req.URL.Query()
	name := `"` + jt.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	err = c.SendAndParse(req, &qresp)
	if err != nil {
		return qresp, err
	}

	results := []ForemanJobTemplate{}
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

func (c *Client) ReadJobTemplate(ctx context.Context, id int) (*ForemanJobTemplate, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf("/%s/%d", JobTemplateEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readJT ForemanJobTemplate
	sendErr := c.SendAndParse(req, &readJT)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("ReadJobTemplate jt: [%+v]", readJT)

	return &readJT, nil
}

func (c *Client) UpdateJobTemplate(ctx context.Context, jtObj *ForemanJobTemplate) (*ForemanJobTemplate, error) {
	endpoint := fmt.Sprintf("/%s/%d", JobTemplateEndpointPrefix, jtObj.Id)

	wrappedJT, err := c.WrapJSONWithTaxonomy("job_template", jtObj)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequestWithContext(
		ctx, http.MethodPut, endpoint, bytes.NewBuffer(wrappedJT),
	)
	if err != nil {
		return nil, err
	}

	var updatedJT ForemanJobTemplate
	err = c.SendAndParse(req, &updatedJT)
	if err != nil {
		return nil, err
	}

	// Handle TemplateInputs

	count_ti := len(jtObj.TemplateInputs)
	if count_ti > 0 {
		updatedTIs := make([]ForemanTemplateInput, count_ti)

		for idx, item := range jtObj.TemplateInputs {
			ti, err := c.UpdateTemplateInput(ctx, &item)
			if err != nil {
				return nil, err
			}

			updatedTIs[idx] = *ti
		}

		updatedJT.TemplateInputs = updatedTIs
	}

	return &updatedJT, nil
}

func (c *Client) DeleteJobTemplate(ctx context.Context, jt *ForemanJobTemplate) error {
	utils.TraceFunctionCall()

	// Delete all TemplateInputs
	// if len(jt.TemplateInputs) > 0 {
	// 	for _, item := range jt.TemplateInputs {
	// 		err := c.DeleteTemplateInput(ctx, &item)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	// Then handle the JobTemplate itself
	endpoint := fmt.Sprintf("/%s/%d", JobTemplateEndpointPrefix, jt.Id)
	req, err := c.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.SendAndParse(req, nil)
}

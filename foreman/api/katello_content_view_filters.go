package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/utils"
	"net/http"
)

const (
	ContentViewFilters           = "/katello/api/content_views/%d/filters"         // :content_view_id
	ContentViewFiltersUpdate     = "/katello/api/content_views/%d/filters/%d"      // :content_view_id, :id
	ContentViewFilterRules       = "/katello/api/content_view_filters/%d/rules"    // :content_view_filter_id
	ContentViewFilterRulesUpdate = "/katello/api/content_view_filters/%d/rules/%d" // :content_view_filter_id, :id
)

// ContentViewFilter is part of a ContentView and filters the presented content according to its rules.
type ContentViewFilter struct {
	ForemanObject

	Inclusion   bool   `json:"inclusion"`
	Description string `json:"description"`

	ContentView  interface{}             `json:"content_view"`
	Repositories []interface{}           `json:"repositories"`
	Type         string                  `json:"type"`
	Rules        []ContentViewFilterRule `json:"rules"`
}

func (cvf *ContentViewFilter) MarshalJSON() ([]byte, error) {
	jsonMap := map[string]interface{}{
		"id":          cvf.Id,
		"name":        cvf.Name,
		"type":        cvf.Type,
		"inclusion":   cvf.Inclusion,
		"description": cvf.Description,
		"rules":       cvf.Rules,
	}

	return json.Marshal(jsonMap)
}

type ContentViewFilterRule struct {
	ForemanObject

	ContentViewFilterId int    `json:"content_view_filter_id"`
	Architecture        string `json:"architecture"`
}

// QueryContentViewFilters returns the filters including their rules
func (c *Client) QueryContentViewFilters(ctx context.Context, cvId int) (QueryResponse, error) {
	utils.TraceFunctionCall()
	queryResponse := QueryResponse{}

	endpoint := fmt.Sprintf(ContentViewFilters, cvId)
	req, err := c.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return queryResponse, err
	}

	err = c.SendAndParse(req, &queryResponse)
	if err != nil {
		return queryResponse, err
	}

	utils.Debugf("queryResponse: %+v", queryResponse)

	var results []ContentViewFilter
	resultsBytes, err := json.Marshal(queryResponse.Results)
	if err != nil {
		return queryResponse, err
	}

	err = json.Unmarshal(resultsBytes, &results)
	if err != nil {
		return queryResponse, err
	}

	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

func (c *Client) CreateKatelloContentViewFilters(ctx context.Context, cvId int, cvfs *[]ContentViewFilter) (*[]ContentViewFilter, error) {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf(ContentViewFilters, cvId)

	var createdCvfs []ContentViewFilter

	for _, cvf := range *cvfs {
		jsonBytes, err := c.WrapJSONWithTaxonomy(nil, cvf)
		if err != nil {
			return nil, err
		}

		req, err := c.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return nil, err
		}

		var createdCvf ContentViewFilter

		err = c.SendAndParse(req, &createdCvf)
		if err != nil {
			return nil, err
		}

		utils.Debugf("createdCvf: %+v", createdCvf)

		createdRules, err := c.CreateKatelloContentViewFilterRules(ctx, createdCvf.Id, &cvf.Rules)
		if err != nil {
			utils.Fatalf("%+v", err)
		}
		createdCvf.Rules = *createdRules

		createdCvfs = append(createdCvfs, createdCvf)
	}
	return &createdCvfs, nil

}

func (c *Client) CreateKatelloContentViewFilterRules(ctx context.Context, cvfId int, cvfrs *[]ContentViewFilterRule) (*[]ContentViewFilterRule, error) {
	utils.TraceFunctionCall()
	endpoint := fmt.Sprintf(ContentViewFilterRules, cvfId)

	// https://apidocs.theforeman.org/katello/latest/apidoc/v2/content_view_filter_rules/create.html

	var createdRules []ContentViewFilterRule
	for _, rule := range *cvfrs {
		jsonBytes, err := c.WrapJSONWithTaxonomy(nil, rule)
		if err != nil {
			return nil, err
		}

		req, err := c.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return nil, err
		}

		var createdRule ContentViewFilterRule

		err = c.SendAndParse(req, &createdRule)
		if err != nil {
			return nil, err
		}

		utils.Debugf("createdRule: %+v", createdRule)

		createdRules = append(createdRules, createdRule)
	}

	return &createdRules, nil
}

func (c *Client) ReadKatelloContentViewFilters(ctx context.Context, cvId int) (*[]ContentViewFilter, error) {
	utils.TraceFunctionCall()

	qr, err := c.QueryContentViewFilters(ctx, cvId)
	if err != nil {
		return nil, err
	}

	utils.Debugf("qr: %+v", qr)
	var cvfs []ContentViewFilter

	// TODO: this is redundant if queryResponse.Results already did the conversion
	for _, item := range qr.Results {
		cvfs = append(cvfs, item.(ContentViewFilter))
	}

	utils.Debugf("read content_view filters: %+v", cvfs)

	return &cvfs, nil
}

func (c *Client) UpdateKatelloContentViewFilters(ctx context.Context, cvId int, cvfs *[]ContentViewFilter) (*[]ContentViewFilter, error) {
	utils.TraceFunctionCall()

	var updatedCvfs []ContentViewFilter

	for _, item := range *cvfs {
		endpoint := fmt.Sprintf(ContentViewFiltersUpdate, cvId, item.Id)

		jsonBytes, err := c.WrapJSONWithTaxonomy(nil, item)
		if err != nil {
			return nil, err
		}

		utils.Debugf("jsonBytes: %s", jsonBytes)

		req, err := c.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return nil, err
		}

		var updatedCvf ContentViewFilter
		err = c.SendAndParse(req, &updatedCvf)
		if err != nil {
			return nil, err
		}

		cvfrs, err := c.UpdateKatelloContentViewFilterRules(ctx, updatedCvf.Id, &item.Rules)
		if err != nil {
			return nil, err
		}
		updatedCvf.Rules = *cvfrs

		utils.Debugf("updatedCvf: %+v", updatedCvf)

		updatedCvfs = append(updatedCvfs, updatedCvf)
	}

	return &updatedCvfs, nil
}

func (c *Client) UpdateKatelloContentViewFilterRules(ctx context.Context, cvId int, cvfrs *[]ContentViewFilterRule) (*[]ContentViewFilterRule, error) {
	utils.TraceFunctionCall()

	var updatedRules []ContentViewFilterRule
	for _, item := range *cvfrs {
		endpoint := fmt.Sprintf(ContentViewFilterRulesUpdate, cvId, item.Id)
		jsonBytes, err := c.WrapJSONWithTaxonomy(nil, item)
		if err != nil {
			return nil, err
		}

		utils.Debugf("jsonBytes: %s", jsonBytes)
		req, err := c.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBuffer(jsonBytes))
		if err != nil {
			return nil, err
		}

		var updatedCvfr ContentViewFilterRule
		err = c.SendAndParse(req, &updatedCvfr)
		if err != nil {
			return nil, err
		}

		utils.Debugf("updatedCvfr: %+v", updatedCvfr)
		updatedRules = append(updatedRules, updatedCvfr)
	}
	return &updatedRules, nil
}

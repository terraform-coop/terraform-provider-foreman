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
	ContentViewEndpointPrefix = "/katello/api/content_views"
	ContentViewById           = ContentViewEndpointPrefix + "/%d"             // :id
	ContentViewsByOrg         = "/katello/api/organizations/%d/content_views" // :organization_id
	ContentViewFilters        = "/katello/api/content_views/%d/filters"       // :content_view_id
	ContentViewFilterRules    = "/katello/api/content_view_filters/%d/rules"  // :content_view_filter_id
)

// A ContentView contains repositories, filters etc. to manage specific views on the Katello contents.
type ContentView struct {
	ForemanObject

	ContentHostCount    int           `json:"content_host_count"`
	Composite           bool          `json:"composite"`
	ComponentIds        []int         `json:"component_ids"`
	Default             bool          `json:"default"`
	VersionCount        int           `json:"version_count"`
	LatestVersion       string        `json:"latest_version"`
	LatestVersionId     int           `json:"latest_version_id"`
	AutoPublish         bool          `json:"auto_publish"`
	SolveDependencies   bool          `json:"solve_dependencies"`
	ImportOnly          bool          `json:"import_only"`
	GeneratedFor        string        `json:"generated_for"`
	RelatedCvCount      int           `json:"related_cv_count"`
	RelatedCompositeCvs []interface{} `json:"related_composite_cvs"`
	NeedsPublish        bool          `json:"needs_publish"`
	Filtered            bool          `json:"filtered"`

	Label       string `json:"label"`
	Description string `json:"description"`

	OrganizationId int `json:"organization_id"`
	Organization   struct {
		Name  string `json:"name"`
		Label string `json:"label"`
		Id    int    `json:"id"`
	} `json:"organization"`

	LastTask struct {
		Id            string `json:"id"`
		StartedAt     string `json:"started_at"`
		Result        string `json:"result"`
		LastSyncWords string `json:"last_sync_words"`
	} `json:"last_task"`

	LatestVersionEnvironments []struct {
		Id    int    `json:"id"`
		Name  string `json:"name"`
		Label string `json:"label"`
	} `json:"latest_version_environments"`

	RepositoryIds []int `json:"repository_ids"`
	Repositories  []struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Label       string `json:"label"`
		ContentType string `json:"content_type"`
	} `json:"repositories"`

	Versions []struct {
		Id             int    `json:"id"`
		Version        string `json:"version"`
		Published      string `json:"published"`
		EnvironmentIds []int  `json:"environment_ids"`
		FiltersApplied bool   `json:"filters_applied"`
	} `json:"versions"`

	Components            []interface{} `json:"components"`
	ContentViewComponents []interface{} `json:"content_view_components"`
	ActivationKeys        []interface{} `json:"activation_keys"`
	Hosts                 []interface{} `json:"hosts"`
	NextVersion           string        `json:"next_version"`
	LastPublished         string        `json:"last_published"`

	Environments []struct {
		Id             int           `json:"id"`
		Label          string        `json:"label"`
		Name           string        `json:"name"`
		ActivationKeys []interface{} `json:"activation_keys"`
		Hosts          []interface{} `json:"hosts"`
		Permissions    struct {
			Readable bool `json:"readable"`
		} `json:"permissions"`
	} `json:"environments"`

	DuplicateRepositoriesToPublish []interface{} `json:"duplicate_repositories_to_publish"`
	Errors                         interface{}   `json:"errors"`

	// Filters are not part of this struct in upstream, but we couple the objects in the provider
	Filters []ContentViewFilter
}

func (cv *ContentView) MarshalJSON() ([]byte, error) {
	jsonMap := map[string]interface{}{
		"id":                 cv.Id,
		"name":               cv.Name,
		"description":        cv.Description,
		"organization_id":    cv.OrganizationId,
		"label":              cv.Label,
		"composite":          cv.Composite,
		"auto_publish":       cv.AutoPublish,       // for CCV
		"solve_dependencies": cv.SolveDependencies, // for CV
		"filtered":           cv.Filtered,
		"repository_ids":     cv.RepositoryIds,
		"component_ids":      cv.ComponentIds,
	}

	return json.Marshal(jsonMap)
}

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

type ContentViewFilterRule struct {
	ForemanObject

	ContentViewFilterId int    `json:"content_view_filter_id"`
	Architecture        string `json:"architecture"`
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

func (c *Client) QueryContentView(ctx context.Context, d *ContentView) (QueryResponse, error) {
	utils.TraceFunctionCall()

	queryResponse := QueryResponse{}

	endpoint := ContentViewEndpointPrefix
	req, err := c.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return queryResponse, err
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	name := `"` + d.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	err = c.SendAndParse(req, &queryResponse)
	if err != nil {
		return queryResponse, err
	}

	utils.Debugf("queryResponse: %+v", queryResponse)

	var results []ContentView
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

func (c *Client) CreateKatelloContentView(ctx context.Context, cv *ContentView) (*ContentView, error) {
	utils.TraceFunctionCall()

	endpoint := ContentViewEndpointPrefix

	jsonBytes, err := c.WrapJSONWithTaxonomy(nil, cv)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	var createdCv ContentView
	err = c.SendAndParse(req, &createdCv)
	if err != nil {
		return nil, err
	}

	cvfs, err := c.CreateKatelloContentViewFilters(ctx, createdCv.Id, &cv.Filters)
	if err != nil {
		return nil, err
	}
	createdCv.Filters = *cvfs

	utils.Debugf("createdCv: %+v", createdCv)

	return &createdCv, nil
}

func (c *Client) CreateKatelloContentViewFilters(ctx context.Context, cvId int, cvf *[]ContentViewFilter) (*[]ContentViewFilter, error) {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf(ContentViewFilters, cvId)

	jsonBytes, err := c.WrapJSONWithTaxonomy(nil, cvf)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	var createdCvf []ContentViewFilter
	err = c.SendAndParse(req, &createdCvf)
	if err != nil {
		return nil, err
	}

	utils.Debugf("createdCvf: %+v", createdCvf)

	return &createdCvf, nil
}

func (c *Client) ReadKatelloContentView(ctx context.Context, d *ContentView) (*ContentView, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf(ContentViewById, d.Id)
	var cv ContentView

	req, err := c.NewRequestWithContext(ctx, http.MethodGet, reqEndpoint, nil)
	if err != nil {
		return nil, err
	}

	err = c.SendAndParse(req, &cv)
	if err != nil {
		return nil, err
	}

	cvfs, err := c.ReadContentViewFilters(ctx, cv.Id)
	if err != nil {
		return nil, err
	}
	cv.Filters = *cvfs

	utils.Debugf("read content_view: %+v", cv)

	return &cv, nil
}

func (c *Client) ReadContentViewFilters(ctx context.Context, cvId int) (*[]ContentViewFilter, error) {
	utils.TraceFunctionCall()

	reqEndpoint := fmt.Sprintf(ContentViewFilters, cvId)
	var cvf []ContentViewFilter

	req, err := c.NewRequestWithContext(ctx, http.MethodGet, reqEndpoint, nil)
	if err != nil {
		return nil, err
	}

	err = c.SendAndParse(req, &cvf)
	if err != nil {
		return nil, err
	}

	utils.Debugf("read content_view filter: %+v", cvf)

	return &cvf, nil
}

func (c *Client) UpdateKatelloContentView(ctx context.Context, cv *ContentView) (*ContentView, error) {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf(ContentViewById, cv.Id)

	jsonBytes, err := c.WrapJSONWithTaxonomy(nil, cv)
	if err != nil {
		return nil, err
	}

	utils.Debugf("jsonBytes: %s", jsonBytes)

	req, err := c.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	var updatedCv ContentView
	err = c.SendAndParse(req, &updatedCv)
	if err != nil {
		return nil, err
	}

	cvfs, err := c.UpdateKatelloContentViewFilters(ctx, updatedCv.Id, &cv.Filters)
	if err != nil {
		return nil, err
	}
	updatedCv.Filters = *cvfs

	utils.Debugf("updatedCv: %+v", updatedCv)

	return &updatedCv, nil
}

func (c *Client) UpdateKatelloContentViewFilters(ctx context.Context, cvId int, cvf *[]ContentViewFilter) (*[]ContentViewFilter, error) {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf(ContentViewFilters, cvId)

	jsonBytes, err := c.WrapJSONWithTaxonomy(nil, cvf)
	if err != nil {
		return nil, err
	}

	utils.Debugf("jsonBytes: %s", jsonBytes)

	req, err := c.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	var updatedCvf []ContentViewFilter
	err = c.SendAndParse(req, &updatedCvf)
	if err != nil {
		return nil, err
	}

	utils.Debugf("updatedCvf: %+v", updatedCvf)

	return &updatedCvf, nil
}

// DeleteKatelloContentView also deletes all Filters and Rules
func (c *Client) DeleteKatelloContentView(ctx context.Context, id int) error {
	utils.TraceFunctionCall()

	endpoint := fmt.Sprintf(ContentViewById, id)

	req, err := c.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	return c.SendAndParse(req, nil)
}

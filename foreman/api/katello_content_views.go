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
	ContentViewById           = ContentViewEndpointPrefix + "/%d"       // :id
	ContentViewPublish        = "/katello/api/content_views/%d/publish" // :id
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
		"id":              cv.Id,
		"name":            cv.Name,
		"description":     cv.Description,
		"organization_id": cv.OrganizationId,
		"label":           cv.Label,
		"composite":       cv.Composite,

		"auto_publish":       cv.AutoPublish,       // for CCV
		"solve_dependencies": cv.SolveDependencies, // for CV
		"filtered":           cv.Filtered,
		"repository_ids":     cv.RepositoryIds,
		"component_ids":      cv.ComponentIds,

		"versions":          cv.Versions,
		"latest_version":    cv.LatestVersion,
		"latest_version_id": cv.LatestVersionId,
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

	// Create Filters if given
	if cv.Filters != nil {
		cvfs, err := c.CreateKatelloContentViewFilters(ctx, createdCv.Id, &cv.Filters)
		if err != nil {
			return nil, err
		}
		createdCv.Filters = *cvfs
	}

	utils.Debugf("createdCv: %+v", createdCv)

	// Publish an initial version
	publishedCv, err := publishNewContentView(c, ctx, createdCv)
	if err != nil {
		return nil, err
	}

	return publishedCv, nil
}

func publishNewContentView(c *Client, ctx context.Context, createdCv ContentView) (*ContentView, error) {
	publishEndpoint := fmt.Sprintf(ContentViewPublish, createdCv.Id)
	req, err := c.NewRequestWithContext(ctx, http.MethodPost, publishEndpoint, nil)
	if err != nil {
		return nil, err
	}

	var publishedCv ContentView
	err = c.SendAndParse(req, &publishedCv)
	if err != nil {
		return nil, err
	}

	utils.Debugf("publishdCv: %+v", publishedCv)
	return &publishedCv, nil
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

	cvfs, err := c.ReadKatelloContentViewFilters(ctx, cv.Id)
	if err != nil {
		return nil, err
	}
	cv.Filters = *cvfs

	utils.Debugf("read content_view: %+v", cv)

	return &cv, nil
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

// DeleteKatelloContentView also deletes all Filters and Rules
func (c *Client) DeleteKatelloContentView(ctx context.Context, id int) error {
	utils.TraceFunctionCall()

	// Using the PUT /remove endpoint also deletes associated filters, filter rules and
	// resolves any associated environments. This allows using only one call to Katello.
	// Otherwise, we would need to fetch all environments for this (composite) content view
	// and delete them first, then finally DELETEing the (C)CV.
	endpoint := fmt.Sprintf(ContentViewById+"/remove", id)

	type contentViewDestroyBody struct {
		DestroyContentView bool `json:"destroy_content_view"`
	}

	body := contentViewDestroyBody{DestroyContentView: true}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := c.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBuffer(bodyJson))
	if err != nil {
		return err
	}

	return c.SendAndParse(req, nil)
}

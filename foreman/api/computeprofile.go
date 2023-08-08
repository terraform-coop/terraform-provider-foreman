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
	ComputeProfileEndpointPrefix = "compute_profiles"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

type ForemanComputeProfile struct {
	ForemanObject
	ComputeAttributes []*ForemanComputeAttribute `json:"compute_attributes,omitempty"`
}

type ForemanComputeAttribute struct {
	ForemanObject
	ComputeResourceId int                    `json:"compute_resource_id"`
	VMAttrs           map[string]interface{} `json:"vm_attrs,omitempty"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// ReadComputeProfile reads the attributes of a ForemanComputeProfile identified by
// the supplied ID and returns a ForemanComputeProfile reference.
func (c *Client) ReadComputeProfile(ctx context.Context, id int) (*ForemanComputeProfile, error) {
	log.Tracef("foreman/api/templatekind.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", ComputeProfileEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readComputeProfile ForemanComputeProfile
	sendErr := c.SendAndParse(req, &readComputeProfile)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readComputeProfile: [%+v]", readComputeProfile)

	return &readComputeProfile, nil
}

// -----------------------------------------------------------------------------
// Query Implementation
// -----------------------------------------------------------------------------

// QueryComputeProfile queries for a ForemanComputeProfile based on the attributes
// of the supplied ForemanComputeProfile reference and returns a QueryResponse
// struct containing query/response metadata and the matching template kinds
func (c *Client) QueryComputeProfile(ctx context.Context, t *ForemanComputeProfile) (QueryResponse, error) {
	log.Tracef("foreman/api/templatekind.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", ComputeProfileEndpointPrefix)
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
	name := `"` + t.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanComputeProfile for
	// the results
	results := []ForemanComputeProfile{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanComputeProfile to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

func (c *Client) CreateComputeprofile(ctx context.Context, d *ForemanComputeProfile) (*ForemanComputeProfile, error) {
	log.Tracef("foreman/api/computeprofile.go#Create")

	reqEndpoint := ComputeProfileEndpointPrefix //fmt.Sprintf("%s", ComputeProfileEndpointPrefix)

	// Copy the original obj and then remove ComputeAttributes
	compProfileClean := new(ForemanComputeProfile)
	compProfileClean.ForemanObject = d.ForemanObject
	compProfileClean.ComputeAttributes = nil

	cprofJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("compute_profile", compProfileClean)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("cprofJSONBytes: [%s]", cprofJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(cprofJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdComputeprofile ForemanComputeProfile
	sendErr := c.SendAndParse(req, &createdComputeprofile)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("d: %+v", d)

	// Add the compute attributes as well
	for i := 0; i < len(d.ComputeAttributes); i++ {
		compattrsEndpoint := fmt.Sprintf("%s/%d/compute_resources/%d/compute_attributes",
			ComputeProfileEndpointPrefix,
			createdComputeprofile.Id,
			d.ComputeAttributes[i].ComputeResourceId)

		by, err := c.WrapJSONWithTaxonomy("compute_attribute", d.ComputeAttributes[i])
		if err != nil {
			return nil, err
		}
		log.Debugf("%s", by)
		req, reqErr = c.NewRequestWithContext(
			ctx, http.MethodPost, compattrsEndpoint, bytes.NewBuffer(by),
		)
		if reqErr != nil {
			return nil, reqErr
		}
		var createdComputeAttribute ForemanComputeAttribute
		sendErr = c.SendAndParse(req, &createdComputeAttribute)
		if sendErr != nil {
			return nil, sendErr
		}
		createdComputeprofile.ComputeAttributes = append(createdComputeprofile.ComputeAttributes, &createdComputeAttribute)
	}

	log.Debugf("createdComputeprofile: [%+v]", createdComputeprofile)

	return &createdComputeprofile, nil
}

func (c *Client) DeleteComputeProfile(ctx context.Context, id int) error {
	log.Tracef("foreman/api/computeprofile.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", ComputeProfileEndpointPrefix, id)
	req, reqErr := c.NewRequestWithContext(ctx, http.MethodDelete, reqEndpoint, nil)
	if reqErr != nil {
		return reqErr
	}

	return c.SendAndParse(req, nil)
}

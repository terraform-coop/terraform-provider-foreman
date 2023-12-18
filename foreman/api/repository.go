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
	// KatelloRepositoryEndpointPrefix api endpoint prefix for katello repositories
	// 'katello/ will be removed, it's a marker to detect talking with katello api
	KatelloRepositoryEndpointPrefix = "katello/repositories"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

type Product struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// ForemanKatelloRepository API model representing a repository.
type ForemanKatelloRepository struct {
	// Inherits the base object's attributes
	ForemanObject

	Description         string  `json:"description"`
	Label               string  `json:"label"`
	ProductId           int     `json:"product_id"`
	Product             Product `json:"product"`
	ContentType         string  `json:"content_type"`
	Url                 string  `json:"url"`
	GpgKeyId            int     `json:"gpg_key_id"`
	Unprotected         bool    `json:"unprotected"`
	ChecksumType        string  `json:"checksum_type"`
	IgnoreGlobalProxy   bool    `json:"ignore_global_proxy"`
	IgnorableContent    string  `json:"ignorable_content"`
	DownloadPolicy      string  `json:"download_policy"`
	DownloadConcurrency int     `json:"download_concurrency"`

	// MirrorOnSync is deprecated
	MirrorOnSync bool `json:"mirror_on_sync"`
	// MirroringPolicy replaces MirrorOnSync
	// Values: "mirror_content_only" or "additive"
	MirroringPolicy string `json:"mirroring_policy"`

	VerifySslOnSync  bool   `json:"verify_ssl_on_sync"`
	UpstreamUsername string `json:"upstream_username"`
	UpstreamPassword string `json:"upstream_password"`

	HttpProxyPolicy string `json:"http_proxy_policy"`
	HttpProxyId     int    `json:"http_proxy_id"`

	DebReleases      string `json:"deb_releases"`
	DebComponents    string `json:"deb_components"`
	DebArchitectures string `json:"deb_architectures"`

	DockerUpstreamName  string `json:"docker_upstream_name"`
	DockerTagsWhitelist string `json:"docker_tags_whitelist"`

	AnsibleCollectionRequirements string `json:"ansible_collection_requirements"`
}

func (r *ForemanKatelloRepository) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"id":                   r.Id,
		"name":                 r.Name,
		"description":          r.Description,
		"label":                r.Label,
		"product_id":           r.ProductId,
		"url":                  r.Url,
		"unprotected":          r.Unprotected,
		"checksum_type":        r.ChecksumType,
		"ignore_global_proxy":  r.IgnoreGlobalProxy,
		"ignorable_content":    r.IgnorableContent,
		"download_policy":      r.DownloadPolicy,
		"download_concurrency": r.DownloadConcurrency,
		"mirroring_policy":     r.MirroringPolicy,
		"mirror_on_sync":       r.MirrorOnSync, // deprecated
		"verify_ssl_on_sync":   r.VerifySslOnSync,
		"upstream_username":    r.UpstreamUsername,
		"upstream_password":    r.UpstreamPassword,
	}

	m["content_type"] = r.ContentType
	switch r.ContentType {
	case "deb":
		m["deb_releases"] = r.DebReleases
		m["deb_components"] = r.DebComponents
		m["deb_architectures"] = r.DebArchitectures
		break
	case "docker":
		m["docker_upstream_name"] = r.DockerUpstreamName
		m["docker_tags_whitelist"] = r.DockerTagsWhitelist
	case "ansible_collection":
		m["ansible_collection_requirements"] = r.AnsibleCollectionRequirements
	}

	if r.GpgKeyId != 0 {
		m["gpg_key_id"] = r.GpgKeyId
	}

	m["http_proxy_policy"] = r.HttpProxyPolicy
	if r.HttpProxyPolicy != "global_default_http_proxy" {
		m["http_proxy_id"] = r.HttpProxyId
	}

	return json.Marshal(m)

}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateKatelloRepository creates a new ForemanKatelloRepository with the attributes of the
// supplied ForemanKatelloRepository reference and returns the created
// ForemanKatelloRepository reference. The returned reference will have its ID and
// other API default values set by this function.
func (c *Client) CreateKatelloRepository(ctx context.Context, p *ForemanKatelloRepository) (*ForemanKatelloRepository, error) {
	log.Tracef("foreman/api/repository.go#Create")

	sJSONBytes, jsonEncErr := c.WrapJSON(nil, p)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("KatelloRepositoryJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		KatelloRepositoryEndpointPrefix,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdKatelloRepository ForemanKatelloRepository
	sendErr := c.SendAndParse(req, &createdKatelloRepository)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdKatelloRepository: [%+v]", createdKatelloRepository)

	return &createdKatelloRepository, nil
}

// ReadKatelloRepository reads the attributes of a ForemanKatelloRepository identified by the
// supplied ID and returns a ForemanKatelloRepository reference.
func (c *Client) ReadKatelloRepository(ctx context.Context, id int) (*ForemanKatelloRepository, error) {
	log.Tracef("foreman/api/repository.go#Read")

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloRepositoryEndpointPrefix, id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readKatelloRepository ForemanKatelloRepository
	sendErr := c.SendAndParse(req, &readKatelloRepository)
	if sendErr != nil {
		return nil, sendErr
	}

	readKatelloRepository.ProductId = readKatelloRepository.Product.Id

	log.Debugf("readKatelloRepository: [%+v]", readKatelloRepository)

	return &readKatelloRepository, nil
}

// UpdateKatelloRepository updates a ForemanKatelloRepository's attributes.  The sync plan
// with the ID of the supplied ForemanKatelloRepository will be updated. A new
// ForemanKatelloRepository reference is returned with the attributes from the result
// of the update operation.
func (c *Client) UpdateKatelloRepository(ctx context.Context, p *ForemanKatelloRepository) (*ForemanKatelloRepository, error) {
	log.Tracef("foreman/api/repository.go#Update")

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloRepositoryEndpointPrefix, p.Id)

	sJSONBytes, jsonEncErr := c.WrapJSON(nil, p)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("KatelloRepositoryJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedKatelloRepository ForemanKatelloRepository
	sendErr := c.SendAndParse(req, &updatedKatelloRepository)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedKatelloRepository: [%+v]", updatedKatelloRepository)

	return &updatedKatelloRepository, nil
}

// DeleteKatelloRepository deletes the ForemanKatelloRepository identified by the supplied ID
func (c *Client) DeleteKatelloRepository(ctx context.Context, id int) error {
	log.Tracef("foreman/api/repository.go#Delete")

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloRepositoryEndpointPrefix, id)

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

// -----------------------------------------------------------------------------
// Query Implementation
// -----------------------------------------------------------------------------

// QueryKatelloRepository queries for a ForemanKatelloRepository based on the attributes of
// the supplied ForemanKatelloRepository reference and returns a QueryResponse struct
// containing query/response metadata and the matching sync plan.
func (c *Client) QueryKatelloRepository(ctx context.Context, p *ForemanKatelloRepository) (QueryResponse, error) {
	log.Tracef("foreman/api/repository.go#Search")

	queryResponse := QueryResponse{}

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		KatelloRepositoryEndpointPrefix,
		nil,
	)
	if reqErr != nil {
		return queryResponse, reqErr
	}

	// dynamically build the query based on the attributes
	reqQuery := req.URL.Query()
	name := `"` + p.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanKatelloRepository for
	// the results
	results := []ForemanKatelloRepository{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanKatelloRepository to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

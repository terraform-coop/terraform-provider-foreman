package api

import (
	"bytes"
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

	Description                   string  `json:"description"`
	Label                         string  `json:"label"`
	ProductId                     int     `json:"product_id"`
	Product                       Product `json:"product"`
	ContentType                   string  `json:"content_type"`
	Url                           string  `json:"url"`
	GpgKeyId                      int     `json:"gpg_key_id"`
	Unprotected                   bool    `json:"unprotected"`
	ChecksumType                  string  `json:"checksum_type"`
	DockerUpstreamName            string  `json:"docker_upstream_name"`
	DockerTagsWhitelist           string  `json:"docker_tags_whitelist"`
	DownloadPolicy                string  `json:"download_policy"`
	DownloadConcurrency           int     `json:"download_concurrency"`
	MirrorOnSync                  bool    `json:"mirror_on_sync"`
	VerifySslOnSync               bool    `json:"verify_ssl_on_sync"`
	UpstreamUsername              string  `json:"upstream_username"`
	UpstreamPassword              string  `json:"upstream_password"`
	DebReleases                   string  `json:"deb_releases"`
	DebComponents                 string  `json:"deb_components"`
	DebArchitectures              string  `json:"deb_architectures"`
	IgnoreGlobalProxy             bool    `json:"ignore_global_proxy"`
	IgnorableContent              string  `json:"ignorable_content"`
	AnsibleCollectionRequirements string  `json:"ansible_collection_requirements"`
	HttpProxyPolicy               string  `json:"http_proxy_policy"`
	HttpProxyId                   int     `json:"http_proxy_id"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateKatelloRepository creates a new ForemanKatelloRepository with the attributes of the
// supplied ForemanKatelloRepository reference and returns the created
// ForemanKatelloRepository reference. The returned reference will have its ID and
// other API default values set by this function.
func (c *Client) CreateKatelloRepository(p *ForemanKatelloRepository) (*ForemanKatelloRepository, error) {
	log.Tracef("foreman/api/repository.go#Create")

	sJSONBytes, jsonEncErr := c.WrapJSON(nil, p)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("KatelloRepositoryJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequest(
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
func (c *Client) ReadKatelloRepository(id int) (*ForemanKatelloRepository, error) {
	log.Tracef("foreman/api/repository.go#Read")

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloRepositoryEndpointPrefix, id)

	req, reqErr := c.NewRequest(
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
func (c *Client) UpdateKatelloRepository(p *ForemanKatelloRepository) (*ForemanKatelloRepository, error) {
	log.Tracef("foreman/api/repository.go#Update")

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloRepositoryEndpointPrefix, p.Id)

	sJSONBytes, jsonEncErr := c.WrapJSON(nil, p)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("KatelloRepositoryJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequest(
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
func (c *Client) DeleteKatelloRepository(id int) error {
	log.Tracef("foreman/api/repository.go#Delete")

	reqEndpoint := fmt.Sprintf("%s/%d", KatelloRepositoryEndpointPrefix, id)

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

// QueryKatelloRepository queries for a ForemanKatelloRepository based on the attributes of
// the supplied ForemanKatelloRepository reference and returns a QueryResponse struct
// containing query/response metadata and the matching sync plan.
func (c *Client) QueryKatelloRepository(p *ForemanKatelloRepository) (QueryResponse, error) {
	log.Tracef("foreman/api/repository.go#Search")

	queryResponse := QueryResponse{}

	req, reqErr := c.NewRequest(
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

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
	ComputeResourceEndpoint = "compute_resources"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanImage API model represents the image name. Images serve as an
// identification string that defines autonomy, authority, or control for
// a portion of a network.

type ForemanImage struct {
	ForemanObject

	// UUID of the image. Can be the path to the image on the compute resource e.g.
	UUID string `json:"uuid"`
	// Username used for login on the image
	Username string `json:"username"`
	// Password for the initial user
	Password string `json:"password"`
	// OperatingSystemId of the operating system associated with the image
	OperatingSystemID int `json:"operatingsystem_id"`
	// ComputeResourceId of the resource this image can be cloned on
	ComputeResourceID int `json:"compute_resource_id"`
	// ArchitectureId of the architecture this image works on
	ArchitectureID int `json:"architecture_id"`
	// Does the image support providing user data (e.g. cloud-init)?
	UserData bool `json:"user_data"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateImage creates a new ForemanImage with the attributes of the supplied
// ForemanImage reference and returns the created ForemanImage reference.
// The returned reference will have its ID and other API default values set by
// this function.
func (c *Client) CreateImage(ctx context.Context, d *ForemanImage, compute_resource int) (*ForemanImage, error) {
	log.Tracef("foreman/api/image.go#Create")

	reqEndpoint := fmt.Sprintf("%s/%d/images", ComputeResourceEndpoint, compute_resource)

	// Custom marshalling content to match the Foreman API.
	// The WrapJSONWithTaxonomy func created problems by adding organization_id/location_id and
	// not handling the types as expected.
	marsh := json.RawMessage(fmt.Sprintf(`{"image":{
		"uuid": "%s",
		"username": "%s",
		"name": "%s",
		"operatingsystem_id": "%d",
		"architecture_id": "%d",
		"compute_resource_id": "%d"
	}}`, d.UUID, d.Username, d.Name, d.OperatingSystemID, d.ArchitectureID, d.ComputeResourceID))

	log.Debugf("marsh: %s", marsh)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(marsh),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdImage ForemanImage
	sendErr := c.SendAndParse(req, &createdImage)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdImage: [%+v]", createdImage)

	return &createdImage, nil
}

// ReadImage reads the attributes of a ForemanImage identified by the
// supplied ID and returns a ForemanImage reference.
func (c *Client) ReadImage(ctx context.Context, d *ForemanImage) (*ForemanImage, error) {
	log.Tracef("foreman/api/image.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d/images/%d", ComputeResourceEndpoint, d.ComputeResourceID, d.Id)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readImage ForemanImage
	sendErr := c.SendAndParse(req, &readImage)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readImage: [%+v]", readImage)

	return &readImage, nil
}

// UpdateImage updates a ForemanImage's attributes.  The image with the ID
// of the supplied ForemanImage will be updated. A new ForemanImage reference
// is returned with the attributes from the result of the update operation.
func (c *Client) UpdateImage(ctx context.Context, d *ForemanImage) (*ForemanImage, error) {
	log.Tracef("foreman/api/image.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d/images/%d", ComputeResourceEndpoint, d.ComputeResourceID, d.Id)

	imageJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("image", d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("imageJSONBytes: [%s]", imageJSONBytes)

	req, reqErr := c.NewRequestWithContext(
		ctx,
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(imageJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedImage ForemanImage
	sendErr := c.SendAndParse(req, &updatedImage)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedImage: [%+v]", updatedImage)

	return &updatedImage, nil
}

// DeleteImage deletes the ForemanImage identified by the supplied ID
func (c *Client) DeleteImage(ctx context.Context, compute_resource, id int) error {
	log.Tracef("foreman/api/image.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d/images/%d", ComputeResourceEndpoint, compute_resource, id)

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

// QueryImage queries for a ForemanImage based on the attributes of the
// supplied ForemanImage reference and returns a QueryResponse struct
// containing query/response metadata and the matching images.
func (c *Client) QueryImage(ctx context.Context, d *ForemanImage) (QueryResponse, error) {
	log.Tracef("foreman/api/image.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("%s/%d/images", ComputeResourceEndpoint, d.ComputeResourceID)
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
	name := `"` + d.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanImage for
	// the results
	results := []ForemanImage{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanImage to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

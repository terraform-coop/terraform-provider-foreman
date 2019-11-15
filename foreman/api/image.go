package api

import (
	"bytes"
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
	// Inherits the base object's attributes
	ForemanObject

	// UUID of the image. Can be the path to the image on the compute resource e.g.
	UUID string `json:"uuid"`
	// Username used for login on the image
	Username string `json:"username"`
	// Name of the image on the compute resource
	Name string `json:"name"`

	// OperatingSystemId of the operating system associated with the image
	OperatingSystemID int `json:"operating_system_id"`
	// ComputeResourceId of the resource this image can be cloned on
	ComputeResourceID int `json:"compute_resource_id"`
	// ArchitectureId of the architecture this image works on
	ArchitectureID int `json:"architecture_id"`
}

// Custom JSON unmarshal function. Unmarshal to the unexported JSON struct
// and then convert over to a ForemanImage struct.
func (fi *ForemanImage) UnmarshalJSON(b []byte) error {
	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fi.ForemanObject = fo

	// Unmarshal into mapstructure and set the rest of the struct properties
	// NOTE(ALL): Properties unmarshalled are of type float64 as opposed to int, hence the below testing
	// Without this, properties will define as default values in state file.
	var fiMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &fiMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	log.Debugf("fiMap: [%v]", fiMap)
	var ok bool

	if fi.Name, ok = fiMap["name"].(string); !ok {
		fi.Name = ""
	}
	if fi.Username, ok = fiMap["username"].(string); !ok {
		fi.Username = ""
	}
	if fi.UUID, ok = fiMap["uuid"].(string); !ok {
		fi.UUID = ""
	}
	if fi.OperatingSystemID, ok = fiMap["operating_system_id"].(int); !ok {
		fi.OperatingSystemID = 0
	}
	if fi.ComputeResourceID, ok = fiMap["compute_resource_id"].(int); !ok {
		fi.ComputeResourceID = 0
	}
	if fi.ArchitectureID, ok = fiMap["architecture_id"].(int); !ok {
		fi.ArchitectureID = 0
	}

	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateImage creates a new ForemanImage with the attributes of the supplied
// ForemanImage reference and returns the created ForemanImage reference.
// The returned reference will have its ID and other API default values set by
// this function.
func (c *Client) CreateImage(d *ForemanImage, compute_resource int) (*ForemanImage, error) {
	log.Tracef("foreman/api/image.go#Create")

	reqEndpoint := fmt.Sprintf("%s/%d/images", ComputeResourceEndpoint, compute_resource)

	imageJSONBytes, jsonEncErr := WrapJson("image", d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("imageJSONBytes: [%s]", imageJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(imageJSONBytes),
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
func (c *Client) ReadImage(d *ForemanImage) (*ForemanImage, error) {
	log.Tracef("foreman/api/image.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d/images/%d", ComputeResourceEndpoint, d.ComputeResourceID, d.Id)

	req, reqErr := c.NewRequest(
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
func (c *Client) UpdateImage(d *ForemanImage) (*ForemanImage, error) {
	log.Tracef("foreman/api/image.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d/images/%d", ComputeResourceEndpoint, d.ComputeResourceID, d.Id)

	imageJSONBytes, jsonEncErr := WrapJson("image", d)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("imageJSONBytes: [%s]", imageJSONBytes)

	req, reqErr := c.NewRequest(
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
func (c *Client) DeleteImage(compute_resource, id int) error {
	log.Tracef("foreman/api/image.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d/images/%d", ComputeResourceEndpoint, compute_resource, id)

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

// QueryImage queries for a ForemanImage based on the attributes of the
// supplied ForemanImage reference and returns a QueryResponse struct
// containing query/response metadata and the matching images.
func (c *Client) QueryImage(d *ForemanImage) (QueryResponse, error) {
	log.Tracef("foreman/api/image.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("%s/%d/images", ComputeResourceEndpoint, d.ComputeResourceID)
	req, reqErr := c.NewRequest(
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

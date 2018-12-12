package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	MediaEndpointPrefix = "media"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The ForemanMedia API model represents a remote installation media.
type ForemanMedia struct {
	// Inherits the base object's attributes
	ForemanObject

	// The path to the medium, can be a URL or a valid NFS server (exclusive
	// of the architecture).  For example:
	//
	// http://mirror.centos.org/centos/$version/os/$arch
	//
	// Where $arch will be substituted for the host's actual OS architecture
	// and $version, $major, $minor will be substituted for the version of the
	// operating system.
	//
	// Solaris and Debian media may also use $release.
	Path string `json:"path"`
	// Operating sysem family. Available values: AIX, Altlinux, Archlinux,
	// Coreos, Debian, Freebsd, Gentoo, Junos, NXOS, Redhat, Solaris, Suse,
	// Windows.
	OSFamily string `json:"os_family"`
	// IDs of operating systems associated with this media
	OperatingSystemIds []int `json:"operatingsystem_ids"`
}

// ForemanMedia struct used for JSON decode.  Foreman API returns the
// operating system ids back as a list of ForemanObjects with some of
// the attributes of an operating system.  However, we are only interested
// in the IDs returned.
type foremanMediaJSON struct {
	OperatingSystems []ForemanObject `json:"operatingsystems"`
}

// Implement the Unmarshaler interface
func (fm *ForemanMedia) UnmarshalJSON(b []byte) error {
	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fm.ForemanObject = fo

	// Unmarshal to temporary JSON struct to get the properties with
	// differently named keys
	var fmJSON foremanMediaJSON
	jsonDecErr = json.Unmarshal(b, &fmJSON)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	fm.OperatingSystemIds = foremanObjectArrayToIdIntArray(fmJSON.OperatingSystems)

	// Unmarshal into mapstructure and set the rest of the struct properties
	var fmMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &fmMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	var ok bool
	if fm.Path, ok = fmMap["path"].(string); !ok {
		fm.Path = ""
	}
	if fm.OSFamily, ok = fmMap["os_family"].(string); !ok {
		fm.OSFamily = ""
	}

	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateMedia creates a new ForemanMedia with the attributes of the supplied
// ForemanMedia reference and returns the created ForemanMedia reference.  The
// returned reference will have its ID and other API default values set by this
// function.
func (c *Client) CreateMedia(m *ForemanMedia) (*ForemanMedia, error) {
	log.Tracef("foreman/api/media.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", MediaEndpointPrefix)

	mJSONBytes, jsonEncErr := json.Marshal(m)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("mediaJSONBytes: [%s]", mJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(mJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdMedia ForemanMedia
	sendErr := c.SendAndParse(req, &createdMedia)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdMedia: [%+v]", createdMedia)

	return &createdMedia, nil
}

// ReadMedia reads the attributes of a ForemanMedia identified by the supplied
// ID and returns a ForemanMedia reference.
func (c *Client) ReadMedia(id int) (*ForemanMedia, error) {
	log.Tracef("foreman/api/media.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", MediaEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readMedia ForemanMedia
	sendErr := c.SendAndParse(req, &readMedia)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readMedia: [%+v]", readMedia)

	return &readMedia, nil
}

// UpdateMedia updates a ForemanMedia's attributes.  The media with the ID of
// the supplied ForemanMedia will be updated. A new ForemanMedia reference is
// returned with the attributes from the result of the update operation.
func (c *Client) UpdateMedia(m *ForemanMedia) (*ForemanMedia, error) {
	log.Tracef("foreman/api/media.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", MediaEndpointPrefix, m.Id)

	mJSONBytes, jsonEncErr := json.Marshal(m)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("mediaJSONBytes: [%s]", mJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(mJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedMedia ForemanMedia
	sendErr := c.SendAndParse(req, &updatedMedia)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedMedia: [%+v]", updatedMedia)

	return &updatedMedia, nil
}

// DeleteMedia deletes the ForemanMedia identified by the supplied ID
func (c *Client) DeleteMedia(id int) error {
	log.Tracef("foreman/api/media.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", MediaEndpointPrefix, id)

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

// QueryMedia queries for a ForemanMedia based on the attributes of the
// supplied ForemanMedia reference and returns a QueryResponse struct
// containing query/response metadata and the matching media.
func (c *Client) QueryMedia(m *ForemanMedia) (QueryResponse, error) {
	log.Tracef("foreman/api/media.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", MediaEndpointPrefix)
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
	name := `"` + m.Name + `"`
	reqQuery.Set("search", "name="+name)

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanMedia for
	// the results
	results := []ForemanMedia{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanMedia to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

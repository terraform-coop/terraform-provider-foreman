package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wayfair/terraform-provider-utils/log"
)

const (
	PartitionTableEndpointPrefix = "ptables"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// The FormanPartitionTable API model represents the disk partition layout
// of the host.  The actual meta-script is stored inside of the Layout
// attribute.
type ForemanPartitionTable struct {
	// Inherits the base object's attributes
	ForemanObject

	// The script that defines the partition table layout
	Layout string `json:"layout"`
	// Whether or not this partition table is a snippet to be embedded in
	// other partition tables
	Snippet bool `json:"snippet"`
	// Any audit comments to associate with the partition table.
	//
	// The Audit Comment field is saved with the template auditing to document
	// the template changes.
	AuditComment string `json:"audit_comment"`
	// Whether or not this partition table is locked for editing
	Locked bool `json:"locked"`
	// Operating sysem family. Available values: AIX, Altlinux, Archlinux,
	// Coreos, Debian, Freebsd, Gentoo, Junos, NXOS, Redhat, Solaris, Suse,
	// Windows.
	OSFamily string `json:"os_family"`

	// IDs of the operating system this partition table applies
	OperatingSystemIds []int `json:"operatingsystem_ids"`
	// IDs of the hostgroups this partition table applies
	HostgroupIds []int `json:"hostgroup_ids"`
	// IDs of the hosts this partition table applies
	HostIds []int `json:"host_ids"`
}

// Intermediary JSON struct - used for unmarshalling JSON data from the
// Foreman API that change key names between create/update and read calls.
type foremanPartitionTableJSON struct {
	OperatingSystems []ForemanObject `json:"operatingsystems"`
}

// Implement the Unmarshaler interface
func (ft *ForemanPartitionTable) UnmarshalJSON(b []byte) error {
	var jsonDecErr error

	// Unmarshal the common Foreman object properties
	var fo ForemanObject
	jsonDecErr = json.Unmarshal(b, &fo)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	ft.ForemanObject = fo

	// Unmarshal to temporary JSON struct to get the properties with differently
	// named keys
	var ftJSON foremanPartitionTableJSON
	jsonDecErr = json.Unmarshal(b, &ftJSON)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	ft.OperatingSystemIds = foremanObjectArrayToIdIntArray(ftJSON.OperatingSystems)

	// Unmarshal into mapstructure and set the rest of the struct properties
	var ftMap map[string]interface{}
	jsonDecErr = json.Unmarshal(b, &ftMap)
	if jsonDecErr != nil {
		return jsonDecErr
	}
	var ok bool
	if ft.Layout, ok = ftMap["layout"].(string); !ok {
		ft.Layout = ""
	}
	if ft.Snippet, ok = ftMap["snippet"].(bool); !ok {
		ft.Snippet = false
	}
	if ft.AuditComment, ok = ftMap["audit_comment"].(string); !ok {
		ft.AuditComment = ""
	}
	if ft.Locked, ok = ftMap["locked"].(bool); !ok {
		ft.Locked = false
	}
	if ft.OSFamily, ok = ftMap["os_family"].(string); !ok {
		ft.OSFamily = ""
	}

	return nil
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreatePartitionTable creates a new ForemanPartitionTable with the attributes
// of the supplied ForemanPartitionTable reference and returns the created
// ForemanPartitionTable reference.  The returned reference will have its ID
// and other API default values set by this function.
func (c *Client) CreatePartitionTable(t *ForemanPartitionTable) (*ForemanPartitionTable, error) {
	log.Tracef("foreman/api/partitiontable.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", PartitionTableEndpointPrefix)

	tJSONBytes, jsonEncErr := json.Marshal(t)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("partitiontableJSONBytes: [%s]", tJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(tJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdPartitionTable ForemanPartitionTable
	sendErr := c.SendAndParse(req, &createdPartitionTable)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdPartitionTable: [%+v]", createdPartitionTable)

	return &createdPartitionTable, nil
}

// ReadPartitionTable reads the attributes of a ForemanPartitionTable
// identified by the supplied ID and returns a ForemanPartitionTable reference.
func (c *Client) ReadPartitionTable(id int) (*ForemanPartitionTable, error) {
	log.Tracef("foreman/api/partitiontable.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", PartitionTableEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readPartitionTable ForemanPartitionTable
	sendErr := c.SendAndParse(req, &readPartitionTable)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("readPartitionTable: [%+v]", readPartitionTable)

	return &readPartitionTable, nil
}

// UpdatePartitionTable updates a ForemanPartitionTable's attributes.  The
// partition table with the ID of the supplied ForemanPartitionTable will be
// updated. A new ForemanPartitionTable reference is returned with the
// attributes from the result of the update operation.
func (c *Client) UpdatePartitionTable(t *ForemanPartitionTable) (*ForemanPartitionTable, error) {
	log.Tracef("foreman/api/partitiontable.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", PartitionTableEndpointPrefix, t.Id)

	tJSONBytes, jsonEncErr := json.Marshal(t)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("partitiontableJSONBytes: [%s]", tJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(tJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedPartitionTable ForemanPartitionTable
	sendErr := c.SendAndParse(req, &updatedPartitionTable)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedPartitionTable: [%+v]", updatedPartitionTable)

	return &updatedPartitionTable, nil
}

// DeletePartitionTable deletes the ForemanPartitionTable identified by the
// supplied ID
func (c *Client) DeletePartitionTable(id int) error {
	log.Tracef("foreman/api/partitiontable.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", PartitionTableEndpointPrefix, id)

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

// QueryPartitionTable queries for a ForemanPartitionTable based on the
// attributes of the supplied ForemanPartitionTable reference and returns a
// QueryResponse struct containing query/response metadata and the matching
// partition tables.
func (c *Client) QueryPartitionTable(t *ForemanPartitionTable) (QueryResponse, error) {
	log.Tracef("foreman/api/partitiontable.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", PartitionTableEndpointPrefix)
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
	// Encode back to JSON, then Unmarshal into []ForemanPartitionTable for
	// the results
	results := []ForemanPartitionTable{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanPartitionTable to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

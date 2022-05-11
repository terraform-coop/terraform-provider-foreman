package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HanseMerkur/terraform-provider-utils/log"
)

const (
	SubnetEndpointPrefix = "subnets"
)

// -----------------------------------------------------------------------------
// Struct Definition and Helpers
// -----------------------------------------------------------------------------

// Domain represents a domain structure inside a ForemanSubnet
type Domain struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// The ForemanSubnet API model represents a subnet
type ForemanSubnet struct {
	// Inherits the base object's attributes
	ForemanObject

	// Subnet network (ie: 192.168.100.0)
	Network string `json:"network"`
	// Netmask for this subnet (ie: 255.255.255.0)
	Mask string `json:"mask"`
	// Gateway server to use when connecting/communicating to anything not
	// on the same network
	Gateway string `json:"gateway"`
	// Primary DNS server for this subnet
	DnsPrimary string `json:"dns_primary"`
	// Secondary DNS server for this subnet
	DnsSecondary string `json:"dns_secondary"`
	// IP address auto-suggestion mode for this subnet.  If set, valid values
	// are "DHCP", "Internal DB", and "None".
	Ipam string `json:"ipam"`
	// Starting IP address for IP auto suggestion
	From string `json:"from"`
	// Ending IP address for IP auto suggestion
	To string `json:"to"`
	// Default boot mode for instances assigned to this subnet.  If set, valid
	// values are "Static" and "DHCP".
	BootMode string `json:"boot_mode"`
	// Network CIDR
	NetworkAddress string `json:"network_address"`
	// VLAN id that is in use in the subnet
	VlanID int `json:"vlanid"`
	// Description for the subnet
	Description string `json:"description"`
	// MTU Default for the subnet
	Mtu int `json:"mtu"`
	// Template ID
	TemplateID int `json:"template_id,omitempty"`
	// DHCP ID
	DhcpID int `json:"dhcp_id,omitempty"`
	// BMC ID
	BmcID *int `json:"bmc_id,omitempty"`
	// TFTP ID
	TftpID int `json:"tftp_id,omitempty"`
	// HTTP Boot ID
	HTTPBootID int `json:"httpboot_id,omitempty"`
	// Domain IDs
	DomainIDs []int `json:"domain_ids"`
	// Domains (for internal use)
	Domains []Domain `json:"domains"`
	// Network Type
	NetworkType string `json:"network_type"`
}

// -----------------------------------------------------------------------------
// CRUD Implementation
// -----------------------------------------------------------------------------

// CreateSubnet creates a new ForemanSubnet with the attributes of the supplied
// ForemanSubnet reference and returns the created ForemanSubnet reference.
// The returned reference will have its ID and other API default values set by
// this function.
func (c *Client) CreateSubnet(s *ForemanSubnet) (*ForemanSubnet, error) {
	log.Tracef("foreman/api/subnet.go#Create")

	reqEndpoint := fmt.Sprintf("/%s", SubnetEndpointPrefix)

	sJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("subnet", s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("sJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPost,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var createdSubnet ForemanSubnet
	sendErr := c.SendAndParse(req, &createdSubnet)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("createdSubnet: [%+v]", createdSubnet)

	return &createdSubnet, nil
}

// ReadSubnet reads the attributes of a ForemanSubnet identified by the
// supplied ID and returns a ForemanSubnet reference.
func (c *Client) ReadSubnet(id int) (*ForemanSubnet, error) {
	log.Tracef("foreman/api/subnet.go#Read")

	reqEndpoint := fmt.Sprintf("/%s/%d", SubnetEndpointPrefix, id)

	req, reqErr := c.NewRequest(
		http.MethodGet,
		reqEndpoint,
		nil,
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var readSubnet ForemanSubnet
	sendErr := c.SendAndParse(req, &readSubnet)
	if sendErr != nil {
		return nil, sendErr
	}

	// copy domain ids from readSubnet.Domains to readSubnet.DomainIDs
	for _, m := range readSubnet.Domains {
		readSubnet.DomainIDs = append(readSubnet.DomainIDs, m.ID)
	}

	log.Debugf("readSubnet: [%+v]", readSubnet)

	return &readSubnet, nil
}

// UpdateSubnet updates a ForemanSubnet's attributes.  The subnet with the ID
// of the supplied ForemanSubnet will be updated. A new ForemanSubnet reference
// is returned with the attributes from the result of the update operation.
func (c *Client) UpdateSubnet(s *ForemanSubnet) (*ForemanSubnet, error) {
	log.Tracef("foreman/api/subnet.go#Update")

	reqEndpoint := fmt.Sprintf("/%s/%d", SubnetEndpointPrefix, s.Id)

	sJSONBytes, jsonEncErr := c.WrapJSONWithTaxonomy("subnet", s)
	if jsonEncErr != nil {
		return nil, jsonEncErr
	}

	log.Debugf("sJSONBytes: [%s]", sJSONBytes)

	req, reqErr := c.NewRequest(
		http.MethodPut,
		reqEndpoint,
		bytes.NewBuffer(sJSONBytes),
	)
	if reqErr != nil {
		return nil, reqErr
	}

	var updatedSubnet ForemanSubnet
	sendErr := c.SendAndParse(req, &updatedSubnet)
	if sendErr != nil {
		return nil, sendErr
	}

	log.Debugf("updatedSubnet: [%+v]", updatedSubnet)

	return &updatedSubnet, nil
}

// DeleteSubnet deletes the ForemanSubnet identified by the supplied ID
func (c *Client) DeleteSubnet(id int) error {
	log.Tracef("foreman/api/subnet.go#Delete")

	reqEndpoint := fmt.Sprintf("/%s/%d", SubnetEndpointPrefix, id)

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

// QuerySubnet queries for a ForemanSubnet based on the attributes of the
// supplied ForemanSubnet reference and returns a QueryResponse struct
// containing query/response metadata and the matching subnets
func (c *Client) QuerySubnet(s *ForemanSubnet) (QueryResponse, error) {
	log.Tracef("foreman/api/subnet.go#Search")

	queryResponse := QueryResponse{}

	reqEndpoint := fmt.Sprintf("/%s", SubnetEndpointPrefix)
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
	if s.Name != "" {
		name := `"` + s.Name + `"`
		reqQuery.Set("search", "name="+name)
	} else if s.Network != "" {
		network := `"` + s.Network + `"`
		reqQuery.Set("search", "network="+network)
	}

	req.URL.RawQuery = reqQuery.Encode()
	sendErr := c.SendAndParse(req, &queryResponse)
	if sendErr != nil {
		return queryResponse, sendErr
	}

	log.Debugf("queryResponse: [%+v]", queryResponse)

	// Results will be Unmarshaled into a []map[string]interface{}
	//
	// Encode back to JSON, then Unmarshal into []ForemanSubnet for
	// the results
	results := []ForemanSubnet{}
	resultsBytes, jsonEncErr := json.Marshal(queryResponse.Results)
	if jsonEncErr != nil {
		return queryResponse, jsonEncErr
	}
	jsonDecErr := json.Unmarshal(resultsBytes, &results)
	if jsonDecErr != nil {
		return queryResponse, jsonDecErr
	}
	// convert the search results from []ForemanSubnet to []interface
	// and set the search results on the query
	iArr := make([]interface{}, len(results))
	for idx, val := range results {
		iArr[idx] = val
	}
	queryResponse.Results = iArr

	return queryResponse, nil
}

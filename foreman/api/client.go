package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/HanseMerkur/terraform-provider-utils/log"
	"github.com/dpotapov/go-spnego"

	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// Every Foreman API call has the following prefix to the path component
	// of the URL.  The client hepler functions utilize this to automatically
	// create endpoint URLs.
	FOREMAN_API_URL_PREFIX = "/api"
	// FOREMAN_KATELLO_API_URL_PREFIX is the Foreman Katello API endpoint
	FOREMAN_KATELLO_API_URL_PREFIX = "/katello/api"
	// API Prefix for Puppet plugin
	FOREMAN_PUPPET_API_URL_PREFIX = "/foreman_puppet/api"
	// The Foreman API allows you to request a specific API version in the
	// Accept header of the HTTP request.  The two supported versions (at
	// the time of writing) are 1 and 2, which version 1 planning on being
	// deprecated after version 1.17.
	FOREMAN_API_VERSION = "2"
)

// ----------------------------------------------------------------------------
// Client / Server Configuration
// ----------------------------------------------------------------------------

// Credentials used to authenticate the client against the remote server - in
// this case, the Foreman API
type ClientCredentials struct {
	Username string
	Password string
}

// Configurable features to apply the REST client
type ClientConfig struct {
	// Whether or not to verify the server's certificate/hostname.  This flag
	// is passed to the TLS config when initializing the REST client for API
	// communication.
	//
	// See 'pkg/crypto/tls/#Config.InsecureSkipVerify' for more information
	TLSInsecureEnabled bool

	// Whether or not the client should try to authenticate to foreman
	// through the HTTP negotiate mechanism.
	NegotiateAuthEnabled bool

	// Information as required by all API calls
	LocationID     int
	OrganizationID int
}

type Client struct {
	// Foreman URL used to communicate and interact with the API.
	server Server
	// Set of credentials to authenticate the client
	credentials ClientCredentials
	// Instance of the HTTP client used to communicate with the webservice.  After
	// the intial setup, the client should never modify or interact directly with
	// the underlying HTTP client and should instead use the helper functions.
	httpClient *http.Client

	// Keep a copy of the client configuration for use in API calls
	clientConfig ClientConfig
}

type HTTPError struct {
	Endpoint   string
	StatusCode int
	RespBody   string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf(
		"HTTP Error:{\n"+
			"  endpoint:   [%s]\n"+
			"  statusCode: [%d]\n"+
			"  respBody:   [%s]\n"+
			"}",
		e.Endpoint,
		e.StatusCode,
		e.RespBody,
	)
}

// KVParameters are used in all inline Parameter Maps. i.e. Host, HostGroup
type ForemanKVParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// JSON obect for creating and updating puppetattributes on hosts and hostgroups
type PuppetAttribute struct {
	Puppetclass_ids []int `json:"puppetclass_ids"`
	ConfigGroup_ids []int `json:"config_group_ids"`
}

func FromKV(kv []ForemanKVParameter) (ret map[string]string) {
	ret = make(map[string]string)
	for _, pair := range kv {
		ret[pair.Name] = pair.Value
	}
	return ret
}

func ToKV(m map[string]interface{}) (ret []ForemanKVParameter) {
	for key, value := range m {
		ret = append(ret, ForemanKVParameter{
			Name:  key,
			Value: value.(string),
		})
	}
	return ret
}

// NewClient creates a new instance of the REST client for communication with
// the API gateway.
func NewClient(s Server, c ClientCredentials, cfg ClientConfig) *Client {
	log.Tracef("foreman/api/client.go#NewClient")
	log.Debugf(
		"Server: [%+v], "+
			"ClientConfig: [%+v]",
		s,
		cfg,
	)

	// Initialize the HTTP client for use by the provider.  The insecure flag
	// from the provider config is used when configuring the TLS settings of
	// the HTTP client.
	cleanClient := cleanhttp.DefaultClient()
	tlsClientConfig := &tls.Config{
		InsecureSkipVerify: cfg.TLSInsecureEnabled,
	}
	if cfg.NegotiateAuthEnabled {
		transCfg := &spnego.Transport{}
		transCfg.TLSClientConfig = tlsClientConfig
		cleanClient.Transport = transCfg
	} else {
		transCfg := &http.Transport{}
		transCfg.TLSClientConfig = tlsClientConfig
		cleanClient.Transport = transCfg
	}

	// Initialize and return the unauthenticated client.
	client := Client{
		httpClient:   cleanClient,
		server:       s,
		credentials:  c,
		clientConfig: cfg,
	}
	return &client
}

// ----------------------------------------------------------------------------
// Client Helper Functions
// ----------------------------------------------------------------------------

// NewRequestWithContext constructs an HTTP request using the client configuration.
// Common request functionality is abstracted and wrapped into this function
// (ie: headers, cookies, MIME-info, etc).  The client should never interact
// with the underlying HTTP client or request object directly.
//
// If the user provides an invalid HTTP method, the function returns 'nil'
// for the request and will return an Error.
//
// The following headers are added and set automatically:
//
//	User-Agent
//	ACCEPT
//	Content-Type
//	Authorization
//
// method
//
//	The HTTP Verb to use.  This should correspond to a 'Method*' constant
//	from 'net/http'.
//
// endpoint
//
//	The server's endpoint to send the request.  The endpoint value is
//	appended to the client's server URL to construct the full URL for the
//	request.  NewRequestWithContext() will automatically prepend the Foreman API URL
//	prefix to the endpoint.
//
// body
//
//	Functions exactly like net/http/NewRequestWithContext()
func (client *Client) NewRequestWithContext(ctx context.Context, method string, endpoint string, body io.Reader) (*http.Request, error) {
	log.Tracef("foreman/api/client.go#NewRequestWithContext")
	log.Debugf(
		"method: [%s], endpoint: [%s]",
		method,
		endpoint,
	)

	if !isValidRequestMethod(method) {
		log.Errorf("Invalid HTTP request method: [%s]\n", method)
		return nil, fmt.Errorf("Invalid HTTP request method: [%s]", method)
	}

	var version_append string = ""

	// Build the URL for the request
	reqURL := client.server.URL
	// Check for katello endpoint
	if strings.HasPrefix(endpoint, "katello") {
		reqURL.Path = FOREMAN_KATELLO_API_URL_PREFIX + strings.TrimPrefix(endpoint, "katello")
	} else if strings.HasPrefix(endpoint, "puppet") {
		reqURL.Path = FOREMAN_PUPPET_API_URL_PREFIX + strings.TrimPrefix(endpoint, "puppet")
	} else {
		if strings.HasPrefix(endpoint, "/") {
			reqURL.Path = FOREMAN_API_URL_PREFIX + endpoint
		} else {
			reqURL.Path = FOREMAN_API_URL_PREFIX + "/" + endpoint
		}
		version_append = "version=" + FOREMAN_API_VERSION
	}

	log.Debugf(
		"reqURL: [%s]\n",
		reqURL.String(),
	)

	// Create the request object, bubble up errors if any were encountered
	req, reqErr := http.NewRequestWithContext(
		ctx,
		strings.ToUpper(method),
		reqURL.String(),
		body,
	)
	if reqErr != nil {
		log.Errorf(
			"Failed to construct a new HTTP request\n"+
				"  Error: %s",
			reqErr.Error(),
		)
		return req, reqErr
	}
	// Add common meta-data and header information for the request
	req.Header.Add("User-Agent", "terraform-provider-foreman")
	req.Header.Add("Accept", "application/json,"+version_append)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(client.credentials.Username, client.credentials.Password)
	return req, nil
}

// isValidRequestMethod is a helper function used to determine if an HTTP
// request method is valid.
//
// NOTE(ALL): Go's HTTP client does not support sending a request with
//
//	the 'CONNECT' method and therefore is not counted as a valid request
//	method. See http.Transport, http.Client for more information.
func isValidRequestMethod(method string) bool {
	// Slice of valid HTTP methods for sending and creating requests
	validHTTPMethods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodTrace,
	}
	// list isn't large - use linear search to validate the method.  Use
	// strings.EqualFold to perform case-insensitive comparisons
	for _, value := range validHTTPMethods {
		if strings.EqualFold(value, method) {
			return true
		}
	}
	return false
}

// Send sends an HTTP request generated by Client.NewRequestWithContext() and returns
// the StatusCode, response. Serves as a facade to the Client's underlying
// HTTP client.
//
// If an error is encountered when reading the server's response, the returned
// StatusCode will be -1.  If an error is encountered during any step of the
// the send and response parsing, an empty slice will be returned as the
// request body.
//
// request
//
//	An HTTP request generated by Client.NewRequestWithContext()
func (client *Client) Send(request *http.Request) (int, []byte, error) {
	log.Tracef("foreman/api/client.go#Send")

	emptySlice := []byte{}

	if request == nil {
		log.Errorf("Client trying to send a nil request")
		return -1, emptySlice, fmt.Errorf("Client trying to send a nil request")
	}

	// Send the request to the server
	resp, respErr := client.httpClient.Do(request)
	if respErr != nil {
		log.Errorf(
			"Error encountered when sending HTTP request to server\n"+
				"  Error: %s",
			respErr.Error(),
		)
		return -1, emptySlice, respErr
	}
	// NOTE(ALL): Golang stdlib dictates that it is the caller's resposibility
	//   to close the response body.  See net/http Response type for more
	//   information.
	defer resp.Body.Close()

	// Read the server's response
	respBody, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Errorf(
			"Error encountered when reading HTTP response from server\n"+
				"  Error: %s",
			readErr.Error(),
		)
		return resp.StatusCode, emptySlice, readErr
	}

	return resp.StatusCode, respBody, nil
}

// SendAndParse sends an HTTP request generated by Client.NewRequestWithContext() and
// parses the server's response for errors.  If an error is encountered during
// the sending or response parsing, the function returns an error.  Otherwise,
// the server's response is unmarshalled into the supplied interface (if the
// interface is not nil).
func (client *Client) SendAndParse(req *http.Request, obj interface{}) error {
	log.Tracef("foreman/api/client.go#SendAndParse")

	statusCode, respBody, sendErr := client.Send(req)
	if sendErr != nil {
		return sendErr
	}

	log.Debugf(
		"server response:{\n"+
			"  endpoint:   [%s]\n"+
			"  method:     [%s]\n"+
			"  statusCode: [%d]\n"+
			"  respBody:   [%s]\n"+
			"}",
		req.URL,
		req.Method,
		statusCode,
		respBody,
	)

	if statusCode < 200 || statusCode > 299 {
		return HTTPError{req.URL.String(), statusCode, string(respBody[:])}
	}

	if obj != nil {
		return json.Unmarshal(respBody, &obj)
	}
	return nil
}

// Taken from terraform-openstack-provider
// CheckDeleted checks the error to see if it's a 404 (Not Found) and, if so,
// sets the resource ID to the empty string instead of throwing an error.
func CheckDeleted(d *schema.ResourceData, err error) error {
	if httpError, ok := err.(HTTPError); ok && httpError.StatusCode == 404 {
		d.SetId("")
		return nil
	}

	return err
}

// wrapParameter wraps the given parameters as an object of its own name
func (client *Client) wrapParameters(name interface{}, item interface{}) (map[string]interface{}, error) {

	var wrapped map[string]interface{}

	if name != nil {
		wrapped = map[string]interface{}{
			fmt.Sprintf("%v", name): item,
		}
	} else {
		data, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &wrapped); err != nil {
			return nil, err
		}
	}

	return wrapped, nil
}

// WrapJSON wraps the given parameters as an object of its own name and marshals it to JSON
func (client *Client) WrapJSON(name interface{}, item interface{}) ([]byte, error) {

	wrapped, _ := client.wrapParameters(name, item)

	return json.Marshal(wrapped)
}

// WrapJSONWithTaxonomy wraps the given parameters as an object of its own name,
// includes additional information for the api call and marshals it to JSON
func (client *Client) WrapJSONWithTaxonomy(name interface{}, item interface{}) ([]byte, error) {

	wrapped, _ := client.wrapParameters(name, item)

	// Workaround for Foreman versions < 1.21 in case no default location/organization was defined for resources
	if client.clientConfig.LocationID >= 0 && client.clientConfig.OrganizationID >= 0 {
		wrapped["location_id"] = client.clientConfig.LocationID
		wrapped["organization_id"] = client.clientConfig.OrganizationID
		log.Debugf("client.go#WrapJSONWithTaxonomy: item %+v", wrapped)
	}

	return json.Marshal(wrapped)
}

package api

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Helper Functions
// ----------------------------------------------------------------------------

// Creates a mock Foreman API server for testing with the client.
//
// NOTE(ALL): It is the caller's responsibility to call Close() on the server
// when finished to prevent a resource leak
func NewForemanAPI() (*http.ServeMux, *httptest.Server) {
	urlMux := http.NewServeMux()
	server := httptest.NewServer(urlMux)
	return urlMux, server
}

// In addition to creating and setting up the mock Foreman API server,
// initialize and set up a Client to communicate with the server.
//
// NOTE(ALL): It is the caller's responsibility to call Close() on the server
// when finished to prevent a resource leak
//
// cred
//   A set of credentials used to authenticate the client
func NewForemanAPIAndClient(cred ClientCredentials, conf ClientConfig) (*http.ServeMux, *httptest.Server, *Client) {
	urlMux, server := NewForemanAPI()
	// Server's URL is stored as a string, parse into a url.URL and point the
	// client at it.  url.Parse() will return *url.URL so dereference before
	// passing to the client.  Safely ignore the error, the server's URL should
	// be valid or there is a problem in Golang stdlib.
	serverURL, _ := url.Parse(server.URL)
	s := Server{
		URL: *serverURL,
	}
	// use unsafe TLS when talking to the mock server
	client := NewClient(s, cred, conf)
	return urlMux, server, client
}

// ----------------------------------------------------------------------------
// NewClient
// ----------------------------------------------------------------------------

// Ensures the client is not modifying the server's URL when creating new
// client struct.
func TestNewClient_ServerURL(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	_, server := NewForemanAPI()
	defer server.Close()
	// create an instance of the client and point it to the server
	serverURL, _ := url.Parse(server.URL)
	client := NewClient(
		Server{
			URL: *serverURL,
		},
		cred,
		conf,
	)

	// Client should have its Server URL set to the mock Foreman API
	if client.server.URL != *serverURL {
		t.Fatalf(
			"Server URL does not match the Client's server URL. "+
				"Expected [%s], got [%s].\n",
			serverURL.String(),
			client.server.URL.String(),
		)
	}

}

// Ensures the client is not modifying the passed credentials when creating
// new client struct
func TestNewClient_Credentials(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{
		Username: "Admin",
		Password: "ChangeMe",
	}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	// Client should have its Server URL set to the mock Foreman API
	if !reflect.DeepEqual(cred, client.credentials) {
		t.Fatalf(
			"Client credentials do not match the expected values. "+
				"Expected [%+v], got [%+v].\n",
			cred,
			client.credentials,
		)
	}
}

// Ensures if the client has enabled TLS insecure, then the client's
// underlying HTTP transport has disabled TLS verification. Otherwise,
// TLS verification should be enabled.
func TestNewClient_ConfigTLSInsecureEnabled(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{}

	testCases := []struct {
		Insecure      bool
		ExpectedValue bool
	}{
		{
			Insecure:      true,
			ExpectedValue: true,
		},
		{
			Insecure:      false,
			ExpectedValue: false,
		},
	}

	for _, testCase := range testCases {

		conf := ClientConfig{
			TLSInsecureEnabled: testCase.Insecure,
		}

		client := NewClient(serv, cred, conf)

		// http.Client.Transport is *http.RoundTripper (interface). Type assert
		// the underlying *http.Transport (struct) to read the transport
		// configuration
		transCfg, _ := client.httpClient.Transport.(*http.Transport)
		tlsCfg := transCfg.TLSClientConfig

		if tlsCfg.InsecureSkipVerify != testCase.ExpectedValue {
			t.Fatalf(
				"Client did not properly set TLS config from configuration. "+
					"Expected TLSClientConfig.InsecureSkipVerify to be [%t], got "+
					"[%t] for insecure [%t]",
				testCase.ExpectedValue,
				tlsCfg.InsecureSkipVerify,
				testCase.Insecure,
			)
		}
	}
}

// ----------------------------------------------------------------------------
// Client.NewRequest
// ----------------------------------------------------------------------------

// Ensures Client.NewRequest() returns an error for a bad HTTP method.
func TestNewRequest_BadHTTPMethodError(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	badHTTPMethods := []string{"",
		"FOO",
		"ZZ",
		"fo0",
		"connect",
		"CONNECT",
		"10",
		" GET",
		"\tGET\n",
		"get\n",
	}
	for _, value := range badHTTPMethods {
		_, badReqErr := client.NewRequest(value, "/foo", nil)
		if badReqErr == nil {
			t.Fatalf(
				"Client.NewRequest did not return error when given invalid HTTP method [%s]. "+
					"Expected [error], got [nil].",
				value,
			)
		}
	}

}

// Ensures Client.NewRequest() does not raise an error when given a valid
// HTTP method.
func TestNewRequest_GoodHTTPMethodNoError(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	goodHTTPMethods := []string{
		"GET",
		"gEt",
		"get",
		"Get",
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodTrace,
		http.MethodPatch,
	}
	for _, value := range goodHTTPMethods {
		_, reqErr := client.NewRequest(value, "/foo", nil)
		if reqErr != nil {
			t.Fatalf(
				"Client.NewRequest returned an error when given valid HTTP method [%s]. "+
					"Expected [nil], got [%s].",
				value,
				reqErr.Error(),
			)
		}
	}

}

// Ensures Client.NewRequest() sets the HTTP request's method to upper case.
func TestNewRequest_RequestMethodToUpper(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	testMethods := []string{
		"get",
		"Get",
		"gEt",
		"geT",
		"GEt",
		"gET",
		"GeT",
		"GET",
		http.MethodGet,
	}
	expectedMethod := "GET"

	for _, value := range testMethods {
		req, _ := client.NewRequest(value, "/foo", nil)
		if req.Method != expectedMethod {
			t.Fatalf(
				"http.Request returned by Client.NewRequest() has incorrect Method. "+
					"Expected [%s], got [%s].\n",
				expectedMethod,
				req.Method,
			)
		}
	}

}

// Ensures Client.NewRequest() sets the correct meta-data on the HTTP
// request.
func TestNewRequest_Header(t *testing.T) {
	serv := Server{}
	cred := ClientCredentials{
		Username: "Admin",
		Password: "ChangeMe",
	}
	conf := ClientConfig{}
	client := NewClient(serv, cred, conf)

	// perform HTTP basic access authorization for the credentials
	// SEE: RFC 7617
	credentialsEncoded := "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(cred.Username+":"+cred.Password),
	)

	req, _ := client.NewRequest(http.MethodGet, "/foo", nil)

	expectedHeader := http.Header{}
	expectedHeader.Add("User-Agent", "terraform-provider-foreman")
	expectedHeader.Add("Content-Type", "application/json")
	expectedHeader.Add("ACCEPT", "application/json,version="+FOREMAN_API_VERSION)
	expectedHeader.Add("Authorization", credentialsEncoded)

	for key := range expectedHeader {
		if req.Header.Get(key) != expectedHeader.Get(key) {
			t.Fatalf(
				"http.Request returned by Client.NewRequest() has incorrect HTTP header. "+
					"Expected [%s], got [%s] for Header key [%s].\n",
				expectedHeader.Get(key),
				req.Header.Get(key),
				key,
			)
		}
	}

}

// Ensures Client.NewRequest() is properly concatenating the server's URL
// and the endpoint when constructing the request's URL.
func TestNewRequest_URL(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	_, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	// map with the endpoint as key and the expected constructed URL path as
	// the value
	testEndpoints := map[string]string{
		"/foo":     FOREMAN_API_URL_PREFIX + "/foo",
		"/":        FOREMAN_API_URL_PREFIX + "/",
		"":         FOREMAN_API_URL_PREFIX + "/",
		"/foo/bar": FOREMAN_API_URL_PREFIX + "/foo/bar",
		"foo/bar":  FOREMAN_API_URL_PREFIX + "/foo/bar",
	}

	for key, value := range testEndpoints {
		req, _ := client.NewRequest(http.MethodGet, key, nil)
		expectedURL := client.server.URL
		expectedURL.Path = value
		if *(req.URL) != expectedURL {
			t.Fatalf(
				"http.Request returned by Client.NewRequest() has incorrect URL. "+
					"Expected [%s], got [%s].\n",
				expectedURL.String(),
				req.URL.String(),
			)
		}
	}

}

// ----------------------------------------------------------------------------
// Client.Send
// ----------------------------------------------------------------------------

// Ensure Client.Send() returns an error when attempting to send a nil
// http.Request reference.
func TestSend_NilRequestError(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	_, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	_, _, sendErr := client.Send(nil)
	if sendErr == nil {
		t.Fatalf(
			"Client.Send() did not return error when given nil. " +
				"Expected [error], got [nil].",
		)
	}

}

// Ensure Client.Send() returns the server's HTTP response status code.
func TestSend_StatusCode(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	// dummy '[GET] /foo' endpoint - just returns 200
	mux.HandleFunc(FOREMAN_API_URL_PREFIX+"/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, _ := client.NewRequest(http.MethodGet, "/foo", nil)
	statusCode, _, _ := client.Send(req)

	if statusCode != http.StatusOK {
		t.Fatalf(
			"Client.Send() did not return correct status code from server. "+
				"Expected [%d], got [%d].",
			http.StatusOK,
			statusCode,
		)
	}

}

// Ensure Client.Send() returns the server's response body.
func TestSend_ResponseBody(t *testing.T) {
	cred := ClientCredentials{}
	conf := ClientConfig{}
	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	expectedRespStr := "Hello, World!"
	expectedRespBody := []byte(expectedRespStr)

	// dummy '[GET] /foo' endpoint - returns "Hello, World!"
	mux.HandleFunc(FOREMAN_API_URL_PREFIX+"/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(expectedRespBody)
	})

	req, _ := client.NewRequest(http.MethodGet, "/foo", nil)
	_, respBody, _ := client.Send(req)

	if string(respBody) != expectedRespStr {
		t.Fatalf(
			"Client.Send() did not return the correct response body from server. "+
				"Expected [%s], got [%s].",
			expectedRespStr,
			respBody,
		)
	}

}

// ----------------------------------------------------------------------------
// Client.SendAndParse
// ----------------------------------------------------------------------------

// Ensure SendAndParse() returns an error when the server responds with a
// status code not in the 2xx range
func TestSendAndParseStatusCodeError(t *testing.T) {
	cred := ClientCredentials{
		Username: "Admin",
		Password: "ChangeMe",
	}
	conf := ClientConfig{}
	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	// dummy '[GET] /foo' endpoint - returns 500 Internal server error
	mux.HandleFunc(FOREMAN_API_URL_PREFIX+"/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req, _ := client.NewRequest(http.MethodGet, "/foo", nil)
	sendErr := client.SendAndParse(req, nil)
	if sendErr == nil {
		t.Errorf(
			"Client.ParseAndSend() did not return an error when the server responded " +
				"with a non-2xx status code. Expected [error] got [nil]",
		)
	}
}

// Ensure SendAndParse() returns no errors when the server responds with a
// status code in the 2xx range
func TestSendAndParseStatusCodeNoError(t *testing.T) {
	cred := ClientCredentials{
		Username: "Admin",
		Password: "ChangeMe",
	}
	conf := ClientConfig{}
	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	// dummy '[GET] /foo' endpoint - returns 500 Internal server error
	mux.HandleFunc(FOREMAN_API_URL_PREFIX+"/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, _ := client.NewRequest(http.MethodGet, "/foo", nil)
	sendErr := client.SendAndParse(req, nil)
	if sendErr != nil {
		t.Errorf(
			"Client.ParseAndSend() did not return an error when the server responded " +
				"with a 2xx status code. Expected [nil] got [error]",
		)
	}
}

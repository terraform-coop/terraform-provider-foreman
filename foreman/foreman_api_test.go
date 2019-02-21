package foreman

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	tfrand "github.com/wayfair/terraform-provider-utils/rand"

	"github.com/hashicorp/terraform/helper/schema"
)

// ----------------------------------------------------------------------------
// Test Helper Functions
// ----------------------------------------------------------------------------

// NewForemanAPI creates a mock Foreman API server for testing with the client.
//
// NOTE(ALL): It is the caller's responsibility to call Close() on the server
// when finished to prevent a resource leak
func NewForemanAPI() (*http.ServeMux, *httptest.Server) {
	urlMux := http.NewServeMux()
	server := httptest.NewServer(urlMux)
	return urlMux, server
}

// NewForemanApiAndClient - in addition to creating and setting up the mock
// Foreman API server, initialize and set up a Client to communicate with the
// server.
//
// NOTE(ALL): It is the caller's responsibility to call Close() on the server
// when finished to prevent a resource leak
func NewForemanAPIAndClient(cred api.ClientCredentials, conf api.ClientConfig) (*http.ServeMux, *httptest.Server, *api.Client) {
	urlMux, server := NewForemanAPI()
	// Server's URL is stored as a string, parse into a url.URL and point the
	// client at it.  url.Parse() will return *url.URL so dereference before
	// passing to the client.  Safely ignore the error, the server's URL should
	// be valid or there is a problem in Golang stdlib.
	serverURL, _ := url.Parse(server.URL)
	s := api.Server{
		URL: *serverURL,
	}
	// use unsafe TLS when talking to the mock server
	client := api.NewClient(s, cred, conf)
	return urlMux, server, client
}

// ParseJSONFile reads the JSON file at the given path and unmarshals the
// file's contents into the supplied obj.  If there was an error when reading
// the file or during the JSON unmarshal, the test reference will error.
func ParseJSONFile(t *testing.T, path string, obj interface{}) {
	bytes, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		t.Errorf(
			"Could not read the file [%s]. Error: [%s]",
			path,
			readErr.Error(),
		)
	}
	jsonDecErr := json.Unmarshal(bytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"Could not decode the contents of the file [%s] into supplied [%T] "+
				"object. Error: [%s]",
			path,
			obj,
			jsonDecErr.Error(),
		)
	}
}

// CompareResourceDataAttributes tests and compares two ResourceData references
// by comparing the attributes of each ResourceData reference and the
// attribute's value.  The attributes are expected to be provided in the map,
// m, which maps the attribute's name to its type.  If the attribute's value
// differ between the ResourceData references, then the test will raise a
// fatal.
func CompareResourceDataAttributes(t *testing.T, m map[string]schema.ValueType, r1 *schema.ResourceData, r2 *schema.ResourceData) {
	for key, value := range m {
		ok1, ok2 := false, false
		attr1, attr2 := r1.Get(key), r2.Get(key)

		if value == schema.TypeBool {
			attr1, ok1 = attr1.(bool)
			attr2, ok2 = attr2.(bool)
		} else if value == schema.TypeInt {
			attr1, ok1 = attr1.(int)
			attr2, ok2 = attr2.(int)
		} else if value == schema.TypeFloat {
			attr1, ok1 = attr1.(float64)
			attr2, ok2 = attr2.(float64)
		} else if value == schema.TypeString {
			attr1, ok1 = attr1.(string)
			attr2, ok2 = attr2.(string)
		} else {
			// Unknown type - ignore test
			return
		}

		// type assertion failed for both - just exit function
		if !ok1 && !ok2 {
			return
		}

		// type assertion failed for one of the attributes
		if (ok1 && !ok2) || (!ok1 && ok2) {
			t.Fatalf(
				"ResourceData references differ in attribute [%s]. The "+
					"values differ in type: [%T], [%T]",
				key,
				attr1,
				attr2,
			)
		}

		// type assertion succeeded, but they differ in value
		if attr1 != attr2 {
			t.Fatalf(
				"ResourceData references differ in attribute [%s]. The "+
					"attributes value differ: [%v], [%v]",
				key,
				attr1,
				attr2,
			)
		}

	} //end for
}

// ----------------------------------------------------------------------------
// Common Tests and Unit Test Framework
// ----------------------------------------------------------------------------

// The location to the test data folder containing the mock responses for
// each API endpoint
const TestDataPath = "testdata/1.11"

// RandForemanObject creates a random ForemanObject
func RandForemanObject() api.ForemanObject {
	return api.ForemanObject{
		Id:        rand.Int(),
		Name:      tfrand.String(10, tfrand.Lower),
		CreatedAt: tfrand.Time().Format("2006-01-02 03:04:05 UTC"),
		UpdatedAt: tfrand.Time().Format("2006-01-02 03:04:05 UTC"),
	}
}

// Type definition describing the signature of a CRUD operation for a
// terraform resource/data source
type CRUDFunc func(*schema.ResourceData, interface{}) error

// Base struct definition for all test cases
type TestCase struct {
	// The name of the CRUD function being tested - used when generating
	// errors/fatals during unit tests
	funcName string
	// The function being tested
	crudFunc CRUDFunc
	// A mock resource data for the resouce/data source - this will be supplied
	// to the CRUDFunc along with the mocked client to test behavior. This
	// resource data is the initial state for the tests at the time the
	// CRUD function is invoked
	resourceData *schema.ResourceData
}

// Test case struct definition for checking if the expected URL is called
// with the correct HTTP method
type TestCaseCorrectURLAndMethod struct {
	TestCase
	// The expected API endpoint to be hit - a mock Foreman API server will
	// be instantiated and the server's mux will be initialized to handle
	// requests to this endpoint
	expectedURI string
	// The expected HTTP method to be used for that URI
	expectedMethod string
}

// TestCRUDFunction_CorrectURLAndMethod ensures each of the CRUD functions
// calls the correct API endpoint with the correct HTTP method.  The test will
// fail if the incorrect URI is invoked, or the wrong HTTP method is sent with
// the request.
func TestCRUDFunction_CorrectURLAndMethod(t *testing.T) {
	testCases := []TestCaseCorrectURLAndMethod{}

	testCases = append(testCases, ResourceForemanArchitectureCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanArchitectureCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanDomainCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanDomainCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanEnvironmentCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanEnvironmentCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanHostCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanHostgroupCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanHostgroupCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanMediaCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanMediaCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanModelCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanModelCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanOperatingSystemCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanOperatingSystemCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanPartitionTableCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanPartitionTableCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanProvisioningTemplateCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanProvisioningTemplateCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanSmartProxyCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSmartProxyCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanSubnetCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSubnetCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, ResourceForemanComputeResourceCorrectURLAndMethodTestCases(t)...)
	testCases = append(testCases, DataSourceForemanComputeResourceCorrectURLAndMethodTestCases(t)...)

	testCases = append(testCases, DataSourceForemanTemplateKindCorrectURLAndMethodTestCases(t)...)

	cred := api.ClientCredentials{}
	conf := api.ClientConfig{}

	for _, testCase := range testCases {
		t.Logf("test case: [%+v]", testCase)

		mux, server, client := NewForemanAPIAndClient(cred, conf)
		defer server.Close()

		// expected handler to be called
		mux.HandleFunc(testCase.expectedURI, func(w http.ResponseWriter, r *http.Request) {
			// assert expected HTTP method
			if !strings.EqualFold(testCase.expectedMethod, r.Method) {
				t.Fatalf(
					"[%s] did not use the correct HTTP method. Expected [%s], "+
						"got [%s] for URI [%s].",
					testCase.funcName,
					testCase.expectedMethod,
					r.Method,
					testCase.expectedURI,
				)
			}
			w.WriteHeader(http.StatusOK)
		})
		// match all other patterns - this should not be invoked
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf(
				"[%s] did not call the correct Foreman API URL.  Expected [%s], got "+
					"[%s]",
				testCase.funcName,
				testCase.expectedURI,
				r.URL.String(),
			)
			w.WriteHeader(http.StatusOK)
		})

		testCase.crudFunc(testCase.resourceData, client)
		server.Close()

	} //end for
}

// TestCRUDFunction_RequestDataEmpty ensures each of the CRUD functions does
// not send any request data with the HTTP request.  The test will fail if the
// server receives any data.
func TestCRUDFunction_RequestDataEmpty(t *testing.T) {
	testCases := []TestCase{}

	testCases = append(testCases, ResourceForemanArchitectureRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanArchitectureRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanDomainRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanDomainRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanEnvironmentRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanEnvironmentRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanHostRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanHostgroupRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanHostgroupRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanMediaRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanMediaRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanModelRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanModelRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanOperatingSystemRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanOperatingSystemRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanPartitionTableRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanPartitionTableRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanProvisioningTemplateRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanProvisioningTemplateRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanSmartProxyRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSmartProxyRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, ResourceForemanSubnetRequestDataEmptyTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSubnetRequestDataEmptyTestCases(t)...)

	testCases = append(testCases, DataSourceForemanTemplateKindRequestDataEmptyTestCases(t)...)

	cred := api.ClientCredentials{}
	conf := api.ClientConfig{}

	for _, testCase := range testCases {
		t.Logf("test case: [%+v]", testCase)

		mux, server, client := NewForemanAPIAndClient(cred, conf)
		defer server.Close()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			reqBytes, _ := ioutil.ReadAll(r.Body)
			bodyLen := len(reqBytes)
			if bodyLen > 0 {
				t.Fatalf(
					"[%s] sent data as part of the request body to [%s].  Expected "+
						"length [0], got [%d].",
					testCase.funcName,
					r.URL.String(),
					bodyLen,
				)
			}
			w.WriteHeader(http.StatusOK)
		})

		testCase.crudFunc(testCase.resourceData, client)
		server.Close()

	} //end for
}

// Test case struct definition for checking the request data payload sent
// to the server
type TestCaseRequestData struct {
	TestCase
	// The expected data - the mock server should receive this along with the
	// request.
	expectedData []byte
}

// TestCRUDFunction_RequestData ensures each of the CRUD functions sends the
// correct request data with the HTTP request.  The test fails if the server
// did not receive the expected data payload
func TestCRUDFunction_RequestData(t *testing.T) {
	testCases := []TestCaseRequestData{}
	testCases = append(testCases, ResourceForemanArchitectureRequestDataTestCases(t)...)
	testCases = append(testCases, ResourceForemanHostRequestDataTestCases(t)...)
	testCases = append(testCases, ResourceForemanHostgroupRequestDataTestCases(t)...)
	testCases = append(testCases, ResourceForemanMediaRequestDataTestCases(t)...)
	testCases = append(testCases, ResourceForemanModelRequestDataTestCases(t)...)
	testCases = append(testCases, ResourceForemanPartitionTableRequestDataTestCases(t)...)
	testCases = append(testCases, ResourceForemanProvisioningTemplateRequestDataTestCases(t)...)
	testCases = append(testCases, ResourceForemanSmartProxyRequestDataTestCases(t)...)
	cred := api.ClientCredentials{}
	conf := api.ClientConfig{}

	for _, testCase := range testCases {
		t.Logf("test case: [%+v]", testCase)

		mux, server, client := NewForemanAPIAndClient(cred, conf)
		defer server.Close()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			reqBytes, _ := ioutil.ReadAll(r.Body)
			bodyLen := len(reqBytes)
			if bodyLen == 0 {
				t.Fatalf(
					"[%s] did not send data as part of the request body to [%s].  "+
						"Expected length [0], got [%d].",
					testCase.funcName,
					r.URL.String(),
					bodyLen,
				)
			}
			if !reflect.DeepEqual(testCase.expectedData, reqBytes) {
				t.Fatalf(
					"[%s] did not send the correct request data to [%s]. "+
						"Expected [%s], got [%s]",
					testCase.funcName,
					r.URL.String(),
					testCase.expectedData,
					reqBytes,
				)
			}
			w.WriteHeader(http.StatusOK)
		})

		testCase.crudFunc(testCase.resourceData, client)
		server.Close()

	} //end for
}

// TestCRUDFunction_StatusCodeError ensures each of the CRUD functions returns
// an error if the mock server responds with a non-2xx status code.  The test
// fails if the CRUD function returns nil.
func TestCRUDFunction_StatusCodeError(t *testing.T) {
	testCases := []TestCase{}

	testCases = append(testCases, ResourceForemanArchitectureStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanArchitectureStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanDomainStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanDomainStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanEnvironmentStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanEnvironmentStatusCodeTestCases(t)...)

	//testCases = append(testCases, ResourceForemanHostStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanHostgroupStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanHostgroupStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanMediaStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanMediaStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanModelStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanModelStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanOperatingSystemStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanOperatingSystemStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanPartitionTableStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanPartitionTableStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanProvisioningTemplateStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanProvisioningTemplateStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanSmartProxyStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSmartProxyStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanSubnetStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSubnetStatusCodeTestCases(t)...)

	testCases = append(testCases, ResourceForemanComputeResourceStatusCodeTestCases(t)...)
	testCases = append(testCases, DataSourceForemanComputeResourceStatusCodeTestCases(t)...)

	testCases = append(testCases, DataSourceForemanTemplateKindStatusCodeTestCases(t)...)

	cred := api.ClientCredentials{}
	conf := api.ClientConfig{}

	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	for _, testCase := range testCases {
		t.Logf("test case: [%+v]", testCase)

		err := testCase.crudFunc(testCase.resourceData, client)
		if err == nil {
			t.Fatalf(
				"[%s] did not return an error when the server responded with a non "+
					"2xx status code.  Expected [error] got [nil]",
				testCase.funcName,
			)
		}

	} //end for
}

// TestCRUDFunction_EmptyResponseError ensures each of the CRUD functions
// returns an error when the server's response does not include any data.  The
// mock server will respond by just setting the response header to a
// http.StatusOK.  The test will fail if the CRUD function returns nil.
func TestCRUDFunction_EmptyResponseError(t *testing.T) {
	testCases := []TestCase{}

	testCases = append(testCases, ResourceForemanArchitectureEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanArchitectureEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanDomainEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanDomainEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanEnvironmentEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanEnvironmentEmptyResponseTestCases(t)...)

	//testCases = append(testCases, ResourceForemanHostEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanHostgroupEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanHostgroupEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanMediaEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanMediaEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanModelEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanModelEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanOperatingSystemEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanOperatingSystemEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanPartitionTableEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanPartitionTableEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanProvisioningTemplateEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanProvisioningTemplateEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanSmartProxyEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSmartProxyEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanSubnetEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSubnetEmptyResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanComputeResourceEmptyResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanComputeResourceEmptyResponseTestCases(t)...)

	testCases = append(testCases, DataSourceForemanTemplateKindEmptyResponseTestCases(t)...)

	cred := api.ClientCredentials{}
	conf := api.ClientConfig{}

	mux, server, client := NewForemanAPIAndClient(cred, conf)
	defer server.Close()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, testCase := range testCases {
		t.Logf("test case: [%+v]", testCase)

		err := testCase.crudFunc(testCase.resourceData, client)
		if err == nil {
			t.Fatalf(
				"[%s] did not return an error when the server sent an empty response. "+
					"Expected [error] got [nil].",
				testCase.funcName,
			)
		}

	} //end for
}

// Compares two ResourceData references by checking their attributes.  Each
// resource/data source should implement its own version of this function
// and use it as an argument to the TestCaseMockResponse struct.  If the
// ResourceData references are not equal, the function is expected to call
// the error or fail methods on the supplied T reference.
type ResourceDataCompareFunc func(*testing.T, *schema.ResourceData, *schema.ResourceData)

// Test case struct definition for simulating a server's response and whether
// or not we expect Terraform to return an error or succeed.
type TestCaseMockResponse struct {
	TestCase
	// Path to the testdata file.  The server will return the contents of this
	// file as its response.
	responseFile string
	// Whether or not the CRUD function is expected to return an error. If the
	// CRUD function is expected to return an error, then the test framework
	// will not attempt to validate the end state of ResourceData reference.
	returnError bool
	// The expected state after the CRUD function is applied. The CRUD function
	// will change the state of the ResourceData reference.  If the supplied
	// reference is nil, then the test does not compare the end state with
	// the expected end state.
	expectedResourceData *schema.ResourceData
	// Comparison function - this should be defined and set for each of the
	// resource/data sources.  This is used to verify the end state of the
	// ResourceData reference
	compareFunc ResourceDataCompareFunc
}

// TestCRUDFunction_MockResponse ensures the correct return and the end state
// of the ResourceData for each CRUD function.  Each test case signals if the
// CRUD function should return an error or not.  If the function is expected to
// error and an error is not returned, the test fails.  Likewise, if the test
// was expected to return nil but returns an error, the test will also fail.
//
// These tests also validate the state of the ResourceData reference at the
// end of each test by calling the compareFunc with the actual and expected
// ResourceData references.  Each resource/data source should implement its
// own compareFunc and signal when to fail/error out the test.
//
// The server's responses are mocked - returning the contents of the file
// at responseFile.
func TestCRUDFunction_MockResponse(t *testing.T) {
	testCases := []TestCaseMockResponse{}

	testCases = append(testCases, ResourceForemanArchitectureMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanArchitectureMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanDomainMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanDomainMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanEnvironmentMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanEnvironmentMockResponseTestCases(t)...)

	//testCases = append(testCases, ResourceForemanHostMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanHostgroupMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanHostgroupMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanMediaMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanMediaMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanModelMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanModelMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanOperatingSystemMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanOperatingSystemMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanPartitionTableMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanPartitionTableMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanProvisioningTemplateMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanProvisioningTemplateMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanSmartProxyMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSmartProxyMockResponseTestCases(t)...)

	testCases = append(testCases, ResourceForemanSubnetMockResponseTestCases(t)...)
	testCases = append(testCases, DataSourceForemanSubnetMockResponseTestCases(t)...)

	testCases = append(testCases, DataSourceForemanTemplateKindMockResponseTestCases(t)...)

	cred := api.ClientCredentials{}
	conf := api.ClientConfig{}

	for _, testCase := range testCases {
		t.Logf("test case: [%+v]", testCase)

		mux, server, client := NewForemanAPIAndClient(cred, conf)
		defer server.Close()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			bytes, readErr := ioutil.ReadFile(testCase.responseFile)
			if readErr != nil {
				t.Fatalf(
					"Error reading file [%s] to send as server response. Failing Test. Error: [%s]",
					testCase.responseFile,
					readErr.Error(),
				)
			}
			w.Write(bytes)
		})

		err := testCase.crudFunc(testCase.resourceData, client)
		server.Close()

		if testCase.returnError {
			// expecting an error, but did not get one
			if err == nil {
				t.Fatalf(
					"[%s] did not return an error when the server responded with "+
						"the contents of [%s] as the response body. The "+
						"operation was expected to fail.",
					testCase.funcName,
					testCase.responseFile,
				)
			}
		} else {
			// expecting success, but got an error
			if err != nil {
				t.Fatalf(
					"[%s] returned an error when the server responded with "+
						"the contents of [%s] as the response body. The "+
						"operation was expected to succeed.",
					testCase.funcName,
					testCase.responseFile,
				)
			}
			// operation succeeded, validate the end state (if we are expecting
			// the end state to have changed)
			if testCase.expectedResourceData != nil {
				testCase.compareFunc(t, testCase.resourceData, testCase.expectedResourceData)
			}
		}

	} //end for
}

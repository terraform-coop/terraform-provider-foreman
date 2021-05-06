package foreman

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	tfrand "github.com/HanseMerkur/terraform-provider-utils/rand"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// -----------------------------------------------------------------------------
// Test Helper Functions
// -----------------------------------------------------------------------------

const HTTPProxiesURI = api.FOREMAN_API_URL_PREFIX + "/http_proxies"
const HTTPProxiesTestDataPath = "testdata/1.11/http_proxies"

// Given a ForemanHTTPProxy, create a mock instance state reference
func ForemanHTTPProxyToInstanceState(obj api.ForemanHTTPProxy) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanHTTPProxy
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["url"] = obj.URL
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanHTTPProxy resource, create a
// mock ResourceData reference.
func MockForemanHTTPProxyResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanHTTPProxy()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a model
// ResourceData reference
func MockForemanHTTPProxyResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanHTTPProxy
	ParseJSONFile(t, path, &obj)
	s := ForemanHTTPProxyToInstanceState(obj)
	return MockForemanHTTPProxyResourceData(s)
}

// Creates a random ForemanHTTPProxy struct
func RandForemanHTTPProxy() api.ForemanHTTPProxy {
	obj := api.ForemanHTTPProxy{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.URL = tfrand.String(30, tfrand.Lower+"/:.")

	return obj
}

// Compares two ResourceData references for a ForemanHTTPProxy resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanHTTPProxyResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

	// compare IDs
	if r1.Id() != r2.Id() {
		t.Fatalf(
			"ResourceData references differ in Id. [%s], [%s]",
			r1.Id(),
			r2.Id(),
		)
	}

	// build the attribute map
	m := map[string]schema.ValueType{}
	r := resourceForemanHTTPProxy()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

}

// -----------------------------------------------------------------------------
// UnmarshalJSON
// -----------------------------------------------------------------------------

// Ensures the JSON unmarshal correctly sets the base attributes from
// ForemanObject
func TestHTTPProxyUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanHTTPProxy
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanHTTPProxy UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanHTTPProxy UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanHTTPProxy
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanHTTPProxy
func TestBuildForemanHTTPProxy(t *testing.T) {

	expectedObj := RandForemanHTTPProxy()
	expectedState := ForemanHTTPProxyToInstanceState(expectedObj)
	expectedResourceData := MockForemanHTTPProxyResourceData(expectedState)

	actualObj := *buildForemanHTTPProxy(expectedResourceData)

	actualState := ForemanHTTPProxyToInstanceState(actualObj)
	actualResourceData := MockForemanHTTPProxyResourceData(actualState)

	ForemanHTTPProxyResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanHTTPProxy
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanHTTPProxy_Value(t *testing.T) {

	expectedObj := RandForemanHTTPProxy()
	expectedState := ForemanHTTPProxyToInstanceState(expectedObj)
	expectedResourceData := MockForemanHTTPProxyResourceData(expectedState)

	actualObj := api.ForemanHTTPProxy{}
	actualState := ForemanHTTPProxyToInstanceState(actualObj)
	actualResourceData := MockForemanHTTPProxyResourceData(actualState)

	setResourceDataFromForemanHTTPProxy(actualResourceData, &expectedObj)

	ForemanHTTPProxyResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanHTTPProxyCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanHTTPProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanHTTPProxyToInstanceState(obj)
	httpProxiesURIByID := HTTPProxiesURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHTTPProxyCreate",
				crudFunc:     resourceForemanHTTPProxyCreate,
				resourceData: MockForemanHTTPProxyResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    HTTPProxiesURI,
					expectedMethod: http.MethodPost,
				},
			},
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHTTPProxyRead",
				crudFunc:     resourceForemanHTTPProxyRead,
				resourceData: MockForemanHTTPProxyResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    httpProxiesURIByID,
					expectedMethod: http.MethodGet,
				},
			},
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHTTPProxyUpdate",
				crudFunc:     resourceForemanHTTPProxyUpdate,
				resourceData: MockForemanHTTPProxyResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    httpProxiesURIByID,
					expectedMethod: http.MethodPut,
				},
			},
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHTTPProxyDelete",
				crudFunc:     resourceForemanHTTPProxyDelete,
				resourceData: MockForemanHTTPProxyResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    httpProxiesURIByID,
					expectedMethod: http.MethodDelete,
				},
			},
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanHTTPProxyRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanHTTPProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanHTTPProxyToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanHTTPProxyRead",
			crudFunc:     resourceForemanHTTPProxyRead,
			resourceData: MockForemanHTTPProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHTTPProxyDelete",
			crudFunc:     resourceForemanHTTPProxyDelete,
			resourceData: MockForemanHTTPProxyResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanHTTPProxyRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanHTTPProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanHTTPProxyToInstanceState(obj)

	rd := MockForemanHTTPProxyResourceData(s)
	obj = *buildForemanHTTPProxy(rd)
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanHTTPProxyCreate",
				crudFunc:     resourceForemanHTTPProxyCreate,
				resourceData: MockForemanHTTPProxyResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanHTTPProxyUpdate",
				crudFunc:     resourceForemanHTTPProxyUpdate,
				resourceData: MockForemanHTTPProxyResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanHTTPProxyStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanHTTPProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanHTTPProxyToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanHTTPProxyCreate",
			crudFunc:     resourceForemanHTTPProxyCreate,
			resourceData: MockForemanHTTPProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHTTPProxyRead",
			crudFunc:     resourceForemanHTTPProxyRead,
			resourceData: MockForemanHTTPProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHTTPProxyUpdate",
			crudFunc:     resourceForemanHTTPProxyUpdate,
			resourceData: MockForemanHTTPProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHTTPProxyDelete",
			crudFunc:     resourceForemanHTTPProxyDelete,
			resourceData: MockForemanHTTPProxyResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanHTTPProxyEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanHTTPProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanHTTPProxyToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanHTTPProxyCreate",
			crudFunc:     resourceForemanHTTPProxyCreate,
			resourceData: MockForemanHTTPProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHTTPProxyRead",
			crudFunc:     resourceForemanHTTPProxyRead,
			resourceData: MockForemanHTTPProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHTTPProxyUpdate",
			crudFunc:     resourceForemanHTTPProxyUpdate,
			resourceData: MockForemanHTTPProxyResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanHTTPProxyMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanHTTPProxy()
	s := ForemanHTTPProxyToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanHTTPProxyCreate",
				crudFunc:     resourceForemanHTTPProxyCreate,
				resourceData: MockForemanHTTPProxyResourceData(s),
			},
			responseFile: HTTPProxiesTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanHTTPProxyResourceDataFromFile(
				t,
				HTTPProxiesTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanHTTPProxyResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanHTTPProxyRead",
				crudFunc:     resourceForemanHTTPProxyRead,
				resourceData: MockForemanHTTPProxyResourceData(s),
			},
			responseFile: HTTPProxiesTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanHTTPProxyResourceDataFromFile(
				t,
				HTTPProxiesTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanHTTPProxyResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanHTTPProxyUpdate",
				crudFunc:     resourceForemanHTTPProxyUpdate,
				resourceData: MockForemanHTTPProxyResourceData(s),
			},
			responseFile: HTTPProxiesTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanHTTPProxyResourceDataFromFile(
				t,
				HTTPProxiesTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanHTTPProxyResourceDataCompare,
		},
	}

}

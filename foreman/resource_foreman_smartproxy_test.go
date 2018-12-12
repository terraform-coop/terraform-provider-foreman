package foreman

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	tfrand "github.com/wayfair/terraform-provider-utils/rand"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// -----------------------------------------------------------------------------
// Test Helper Functions
// -----------------------------------------------------------------------------

const SmartProxiesURI = api.FOREMAN_API_URL_PREFIX + "/smart_proxies"
const SmartProxiesTestDataPath = "testdata/1.11/smart_proxies"

// Given a ForemanSmartProxy, create a mock instance state reference
func ForemanSmartProxyToInstanceState(obj api.ForemanSmartProxy) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanSmartProxy
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["url"] = obj.URL
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanSmartProxy resource, create a
// mock ResourceData reference.
func MockForemanSmartProxyResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanSmartProxy()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a model
// ResourceData reference
func MockForemanSmartProxyResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanSmartProxy
	ParseJSONFile(t, path, &obj)
	s := ForemanSmartProxyToInstanceState(obj)
	return MockForemanSmartProxyResourceData(s)
}

// Creates a random ForemanSmartProxy struct
func RandForemanSmartProxy() api.ForemanSmartProxy {
	obj := api.ForemanSmartProxy{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.URL = tfrand.String(30, tfrand.Lower+"/:.")

	return obj
}

// Compares two ResourceData references for a ForemanSmartProxy resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanSmartProxyResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanSmartProxy()
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
func TestSmartProxyUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanSmartProxy
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanSmartProxy UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanSmartProxy UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanSmartProxy
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanSmartProxy
func TestBuildForemanSmartProxy(t *testing.T) {

	expectedObj := RandForemanSmartProxy()
	expectedState := ForemanSmartProxyToInstanceState(expectedObj)
	expectedResourceData := MockForemanSmartProxyResourceData(expectedState)

	actualObj := *buildForemanSmartProxy(expectedResourceData)

	actualState := ForemanSmartProxyToInstanceState(actualObj)
	actualResourceData := MockForemanSmartProxyResourceData(actualState)

	ForemanSmartProxyResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanSmartProxy
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanSmartProxy_Value(t *testing.T) {

	expectedObj := RandForemanSmartProxy()
	expectedState := ForemanSmartProxyToInstanceState(expectedObj)
	expectedResourceData := MockForemanSmartProxyResourceData(expectedState)

	actualObj := api.ForemanSmartProxy{}
	actualState := ForemanSmartProxyToInstanceState(actualObj)
	actualResourceData := MockForemanSmartProxyResourceData(actualState)

	setResourceDataFromForemanSmartProxy(actualResourceData, &expectedObj)

	ForemanSmartProxyResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanSmartProxyCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanSmartProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanSmartProxyToInstanceState(obj)
	smartProxiesURIById := SmartProxiesURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanSmartProxyCreate",
				crudFunc:     resourceForemanSmartProxyCreate,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			expectedURI:    SmartProxiesURI,
			expectedMethod: http.MethodPost,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanSmartProxyRead",
				crudFunc:     resourceForemanSmartProxyRead,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			expectedURI:    smartProxiesURIById,
			expectedMethod: http.MethodGet,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanSmartProxyUpdate",
				crudFunc:     resourceForemanSmartProxyUpdate,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			expectedURI:    smartProxiesURIById,
			expectedMethod: http.MethodPut,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanSmartProxyDelete",
				crudFunc:     resourceForemanSmartProxyDelete,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			expectedURI:    smartProxiesURIById,
			expectedMethod: http.MethodDelete,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanSmartProxyRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanSmartProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanSmartProxyToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanSmartProxyRead",
			crudFunc:     resourceForemanSmartProxyRead,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanSmartProxyDelete",
			crudFunc:     resourceForemanSmartProxyDelete,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanSmartProxyRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanSmartProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanSmartProxyToInstanceState(obj)

	rd := MockForemanSmartProxyResourceData(s)
	obj = *buildForemanSmartProxy(rd)
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanSmartProxyCreate",
				crudFunc:     resourceForemanSmartProxyCreate,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanSmartProxyUpdate",
				crudFunc:     resourceForemanSmartProxyUpdate,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanSmartProxyStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanSmartProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanSmartProxyToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanSmartProxyCreate",
			crudFunc:     resourceForemanSmartProxyCreate,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanSmartProxyRead",
			crudFunc:     resourceForemanSmartProxyRead,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanSmartProxyUpdate",
			crudFunc:     resourceForemanSmartProxyUpdate,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanSmartProxyDelete",
			crudFunc:     resourceForemanSmartProxyDelete,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanSmartProxyEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanSmartProxy{}
	obj.Id = rand.Intn(100)
	s := ForemanSmartProxyToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanSmartProxyCreate",
			crudFunc:     resourceForemanSmartProxyCreate,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanSmartProxyRead",
			crudFunc:     resourceForemanSmartProxyRead,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanSmartProxyUpdate",
			crudFunc:     resourceForemanSmartProxyUpdate,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanSmartProxyMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanSmartProxy()
	s := ForemanSmartProxyToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanSmartProxyCreate",
				crudFunc:     resourceForemanSmartProxyCreate,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			responseFile: SmartProxiesTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanSmartProxyResourceDataFromFile(
				t,
				SmartProxiesTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanSmartProxyResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanSmartProxyRead",
				crudFunc:     resourceForemanSmartProxyRead,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			responseFile: SmartProxiesTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanSmartProxyResourceDataFromFile(
				t,
				SmartProxiesTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanSmartProxyResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanSmartProxyUpdate",
				crudFunc:     resourceForemanSmartProxyUpdate,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			responseFile: SmartProxiesTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanSmartProxyResourceDataFromFile(
				t,
				SmartProxiesTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanSmartProxyResourceDataCompare,
		},
	}

}

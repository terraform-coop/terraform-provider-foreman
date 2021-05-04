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

const UsergroupsURI = api.FOREMAN_API_URL_PREFIX + "/usergroups"
const UsergroupsTestDataPath = "testdata/1.11/usergroups"

// Given a ForemanUsergroup, create a mock instance state reference
func ForemanUsergroupToInstanceState(obj api.ForemanUsergroup) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanUsergroup
	attr := map[string]string{}
	attr["name"] = obj.Name
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanUsergroup resource, create a
// mock ResourceData reference.
func MockForemanUsergroupResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanUsergroup()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a usergroup
// ResourceData reference
func MockForemanUsergroupResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanUsergroup
	ParseJSONFile(t, path, &obj)
	s := ForemanUsergroupToInstanceState(obj)
	return MockForemanUsergroupResourceData(s)
}

// Creates a random ForemanUsergroup struct
func RandForemanUsergroup() api.ForemanUsergroup {
	obj := api.ForemanUsergroup{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Name = tfrand.String(15)

	return obj
}

// Compares two ResourceData references for a ForemanUsergroup resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanUsergroupResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanUsergroup()
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
func TestUsergroupUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanUsergroup
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanUsergroup UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanUsergroup UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanUsergroup
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanUsergroup
func TestBuildForemanUsergroup(t *testing.T) {

	expectedObj := RandForemanUsergroup()
	expectedState := ForemanUsergroupToInstanceState(expectedObj)
	expectedResourceData := MockForemanUsergroupResourceData(expectedState)

	actualObj := *buildForemanUsergroup(expectedResourceData)

	actualState := ForemanUsergroupToInstanceState(actualObj)
	actualResourceData := MockForemanUsergroupResourceData(actualState)

	ForemanUsergroupResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanUsergroup
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanUsergroup_Value(t *testing.T) {

	expectedObj := RandForemanUsergroup()
	expectedState := ForemanUsergroupToInstanceState(expectedObj)
	expectedResourceData := MockForemanUsergroupResourceData(expectedState)

	actualObj := api.ForemanUsergroup{}
	actualState := ForemanUsergroupToInstanceState(actualObj)
	actualResourceData := MockForemanUsergroupResourceData(actualState)

	setResourceDataFromForemanUsergroup(actualResourceData, &expectedObj)

	ForemanUsergroupResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanUsergroupCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanUsergroup{}
	obj.Id = rand.Intn(100)
	s := ForemanUsergroupToInstanceState(obj)
	usergroupsURIById := UsergroupsURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanUsergroupCreate",
				crudFunc:     resourceForemanUsergroupCreate,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			expectedURI:    UsergroupsURI,
			expectedMethod: http.MethodPost,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanUsergroupRead",
				crudFunc:     resourceForemanUsergroupRead,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			expectedURI:    usergroupsURIById,
			expectedMethod: http.MethodGet,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanUsergroupUpdate",
				crudFunc:     resourceForemanUsergroupUpdate,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			expectedURI:    usergroupsURIById,
			expectedMethod: http.MethodPut,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanUsergroupDelete",
				crudFunc:     resourceForemanUsergroupDelete,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			expectedURI:    usergroupsURIById,
			expectedMethod: http.MethodDelete,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanUsergroupRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanUsergroup{}
	obj.Id = rand.Intn(100)
	s := ForemanUsergroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanUsergroupRead",
			crudFunc:     resourceForemanUsergroupRead,
			resourceData: MockForemanUsergroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanUsergroupDelete",
			crudFunc:     resourceForemanUsergroupDelete,
			resourceData: MockForemanUsergroupResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanUsergroupRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanUsergroup{}
	obj.Id = rand.Intn(100)
	s := ForemanUsergroupToInstanceState(obj)

	rd := MockForemanUsergroupResourceData(s)
	obj = *buildForemanUsergroup(rd)
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanUsergroupCreate",
				crudFunc:     resourceForemanUsergroupCreate,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanUsergroupUpdate",
				crudFunc:     resourceForemanUsergroupUpdate,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanUsergroupStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanUsergroup{}
	obj.Id = rand.Intn(100)
	s := ForemanUsergroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanUsergroupCreate",
			crudFunc:     resourceForemanUsergroupCreate,
			resourceData: MockForemanUsergroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanUsergroupRead",
			crudFunc:     resourceForemanUsergroupRead,
			resourceData: MockForemanUsergroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanUsergroupUpdate",
			crudFunc:     resourceForemanUsergroupUpdate,
			resourceData: MockForemanUsergroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanUsergroupDelete",
			crudFunc:     resourceForemanUsergroupDelete,
			resourceData: MockForemanUsergroupResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanUsergroupEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanUsergroup{}
	obj.Id = rand.Intn(100)
	s := ForemanUsergroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanUsergroupCreate",
			crudFunc:     resourceForemanUsergroupCreate,
			resourceData: MockForemanUsergroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanUsergroupRead",
			crudFunc:     resourceForemanUsergroupRead,
			resourceData: MockForemanUsergroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanUsergroupUpdate",
			crudFunc:     resourceForemanUsergroupUpdate,
			resourceData: MockForemanUsergroupResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanUsergroupMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanUsergroup()
	s := ForemanUsergroupToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanUsergroupCreate",
				crudFunc:     resourceForemanUsergroupCreate,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			responseFile: UsergroupsTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanUsergroupResourceDataFromFile(
				t,
				UsergroupsTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanUsergroupResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanUsergroupRead",
				crudFunc:     resourceForemanUsergroupRead,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			responseFile: UsergroupsTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanUsergroupResourceDataFromFile(
				t,
				UsergroupsTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanUsergroupResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanUsergroupUpdate",
				crudFunc:     resourceForemanUsergroupUpdate,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			responseFile: UsergroupsTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanUsergroupResourceDataFromFile(
				t,
				UsergroupsTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanUsergroupResourceDataCompare,
		},
	}

}

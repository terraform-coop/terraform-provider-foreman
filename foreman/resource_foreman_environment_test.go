package foreman

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// -----------------------------------------------------------------------------
// Test Helper Functions
// -----------------------------------------------------------------------------

const EnvironmentsURI = api.FOREMAN_API_URL_PREFIX + "/environments"
const EnvironmentsTestDataPath = "testdata/1.11/environments"

// Given a ForemanEnvironment, create a mock instance state reference
func ForemanEnvironmentToInstanceState(obj api.ForemanEnvironment) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanEnvironment
	attr := map[string]string{}
	attr["name"] = obj.Name
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanEnvironment resource, create a
// mock ResourceData reference.
func MockForemanEnvironmentResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanEnvironment()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a  domain
// ResourceData reference
func MockForemanEnvironmentResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanEnvironment
	ParseJSONFile(t, path, &obj)
	s := ForemanEnvironmentToInstanceState(obj)
	return MockForemanEnvironmentResourceData(s)
}

// Creates a random ForemanEnvironment struct
func RandForemanEnvironment() api.ForemanEnvironment {
	obj := api.ForemanEnvironment{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	return obj
}

// Compares two ResourceData references for a ForemanEnvironment resource.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanEnvironmentResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanEnvironment()
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
func TestEnvironmentUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanEnvironment
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanEnvironment UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanEnvironment UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanEnvironment
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanEnvironment_Value(t *testing.T) {

	expectedObj := RandForemanEnvironment()
	expectedState := ForemanEnvironmentToInstanceState(expectedObj)
	expectedResourceData := MockForemanEnvironmentResourceData(expectedState)

	actualObj := api.ForemanEnvironment{}
	actualState := ForemanEnvironmentToInstanceState(actualObj)
	actualResourceData := MockForemanEnvironmentResourceData(actualState)

	setResourceDataFromForemanEnvironment(actualResourceData, &expectedObj)

	ForemanEnvironmentResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanEnvironmentCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanEnvironment{}
	obj.Id = rand.Intn(100)
	s := ForemanEnvironmentToInstanceState(obj)
	environmentsURIById := EnvironmentsURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanEnvironmentRead",
				crudFunc:     resourceForemanEnvironmentRead,
				resourceData: MockForemanEnvironmentResourceData(s),
			},
			expectedURI:    environmentsURIById,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanEnvironmentRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanEnvironment{}
	obj.Id = rand.Intn(100)
	s := ForemanEnvironmentToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanEnvironmentRead",
			crudFunc:     resourceForemanEnvironmentRead,
			resourceData: MockForemanEnvironmentResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanEnvironmentStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanEnvironment{}
	obj.Id = rand.Intn(100)
	s := ForemanEnvironmentToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanEnvironmentRead",
			crudFunc:     resourceForemanEnvironmentRead,
			resourceData: MockForemanEnvironmentResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanEnvironmentEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanEnvironment{}
	obj.Id = rand.Intn(100)
	s := ForemanEnvironmentToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanEnvironmentRead",
			crudFunc:     resourceForemanEnvironmentRead,
			resourceData: MockForemanEnvironmentResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanEnvironmentMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanEnvironment()
	s := ForemanEnvironmentToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanEnvironmentRead",
				crudFunc:     resourceForemanEnvironmentRead,
				resourceData: MockForemanEnvironmentResourceData(s),
			},
			responseFile: EnvironmentsTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanEnvironmentResourceDataFromFile(
				t,
				EnvironmentsTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanEnvironmentResourceDataCompare,
		},
	}

}

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

const OperatingSystemsURI = api.FOREMAN_API_URL_PREFIX + "/operatingsystems"
const OperatingSystemsTestDataPath = "testdata/1.11/operatingsystems"

// Given a ForemanOperatingSystem, create a mock instance state reference
func ForemanOperatingSystemToInstanceState(obj api.ForemanOperatingSystem) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanOperatingSystem
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["major"] = obj.Major
	attr["minor"] = obj.Minor
	attr["title"] = obj.Title
	attr["description"] = obj.Description
	attr["family"] = obj.Family
	attr["release_name"] = obj.ReleaseName
	attr["password_hash"] = obj.PasswordHash
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanOperatingSystem resource, create a
// mock ResourceData reference.
func MockForemanOperatingSystemResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanOperatingSystem()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a  domain
// ResourceData reference
func MockForemanOperatingSystemResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanOperatingSystem
	ParseJSONFile(t, path, &obj)
	s := ForemanOperatingSystemToInstanceState(obj)
	return MockForemanOperatingSystemResourceData(s)
}

// Creates a random ForemanOperatingSystem struct
func RandForemanOperatingSystem() api.ForemanOperatingSystem {
	obj := api.ForemanOperatingSystem{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Major = strconv.Itoa(rand.Intn(20))
	obj.Minor = strconv.Itoa(rand.Intn(20))
	obj.Title = tfrand.String(30, tfrand.Lower+tfrand.Digit+".")
	obj.Description = tfrand.String(30, tfrand.Lower+" ")
	obj.Family = tfrand.String(10, tfrand.Lower)
	obj.ReleaseName = tfrand.String(15, tfrand.Lower+" ")
	obj.PasswordHash = tfrand.String(5, tfrand.Lower+tfrand.Digit)

	return obj
}

// Compares two ResourceData references for a ForemanOperatingSystem resource.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanOperatingSystemResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanOperatingSystem()
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
func TestOperatingSystemUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanOperatingSystem
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanOperatingSystem UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanOperatingSystem UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanOperatingSystem
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanOperatingSystem_Value(t *testing.T) {

	expectedObj := RandForemanOperatingSystem()
	expectedState := ForemanOperatingSystemToInstanceState(expectedObj)
	expectedResourceData := MockForemanOperatingSystemResourceData(expectedState)

	actualObj := api.ForemanOperatingSystem{}
	actualState := ForemanOperatingSystemToInstanceState(actualObj)
	actualResourceData := MockForemanOperatingSystemResourceData(actualState)

	setResourceDataFromForemanOperatingSystem(actualResourceData, &expectedObj)

	ForemanOperatingSystemResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanOperatingSystemCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanOperatingSystem{}
	obj.Id = rand.Intn(100)
	s := ForemanOperatingSystemToInstanceState(obj)
	operatingSystemsURIById := OperatingSystemsURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanOperatingSystemRead",
				crudFunc:     resourceForemanOperatingSystemRead,
				resourceData: MockForemanOperatingSystemResourceData(s),
			},
			expectedURI:    operatingSystemsURIById,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanOperatingSystemRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanOperatingSystem{}
	obj.Id = rand.Intn(100)
	s := ForemanOperatingSystemToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanOperatingSystemRead",
			crudFunc:     resourceForemanOperatingSystemRead,
			resourceData: MockForemanOperatingSystemResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanOperatingSystemStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanOperatingSystem{}
	obj.Id = rand.Intn(100)
	s := ForemanOperatingSystemToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanOperatingSystemRead",
			crudFunc:     resourceForemanOperatingSystemRead,
			resourceData: MockForemanOperatingSystemResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanOperatingSystemEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanOperatingSystem{}
	obj.Id = rand.Intn(100)
	s := ForemanOperatingSystemToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanOperatingSystemRead",
			crudFunc:     resourceForemanOperatingSystemRead,
			resourceData: MockForemanOperatingSystemResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanOperatingSystemMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanOperatingSystem()
	s := ForemanOperatingSystemToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanOperatingSystemRead",
				crudFunc:     resourceForemanOperatingSystemRead,
				resourceData: MockForemanOperatingSystemResourceData(s),
			},
			responseFile: OperatingSystemsTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanOperatingSystemResourceDataFromFile(
				t,
				OperatingSystemsTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanOperatingSystemResourceDataCompare,
		},
	}

}

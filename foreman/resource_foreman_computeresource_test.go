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

const ComputeResourcesURI = api.FOREMAN_API_URL_PREFIX + "/compute_resources"
const ComputeResourcesTestDataPath = "testdata/1.11/compute_resources"

// Given a ForemanComputeResource, create a mock instance state reference
func ForemanComputeResourceToInstanceState(obj api.ForemanComputeResource) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanComputeResource
	state.Attributes = map[string]string{
		"name":               obj.Name,
		"url":                obj.URL,
		"hypervisor":         obj.Provider,
		"displaytype":        obj.DisplayType,
		"user":               obj.User,
		"password":           obj.Password,
		"datacenter":         obj.Datacenter,
		"server":             obj.Server,
		"setconsolepassword": strconv.FormatBool(obj.SetConsolePassword),
		"cachingenabled":     strconv.FormatBool(obj.CachingEnabled),
	}
	return &state
}

// Given a mock instance state for a ForemanComputeResource resource, create a
// mock ResourceData reference.
func MockForemanComputeResourceResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanComputeResource()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a  compute_resource
// ResourceData reference
func MockForemanComputeResourceResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanComputeResource
	ParseJSONFile(t, path, &obj)
	s := ForemanComputeResourceToInstanceState(obj)
	return MockForemanComputeResourceResourceData(s)
}

// Creates a random ForemanComputeResource struct
func RandForemanComputeResource() api.ForemanComputeResource {
	fo := RandForemanObject()
	return api.ForemanComputeResource{
		ForemanObject:      fo,
		Name:               tfrand.String(10, tfrand.Lower),
		URL:                "https://" + tfrand.String(10, tfrand.Lower) + ".localdomain",
		Provider:           tfrand.String(10, tfrand.Lower),
		DisplayType:        tfrand.String(10, tfrand.Lower),
		User:               tfrand.String(10, tfrand.Lower),
		Password:           tfrand.String(10, tfrand.Lower),
		Datacenter:         tfrand.String(10, tfrand.Lower),
		Server:             tfrand.String(10, tfrand.Lower),
		SetConsolePassword: rand.Intn(2) > 0,
		CachingEnabled:     rand.Intn(2) > 0,
	}
}

// Compares two ResourceData references for a ForemanComputeResource resource.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanComputeResourceResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanComputeResource()
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
func TestComputeResourceUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanComputeResource
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanComputeResource UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanComputeResource UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanComputeResource
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanComputeResource_Value(t *testing.T) {

	expectedObj := RandForemanComputeResource()
	expectedState := ForemanComputeResourceToInstanceState(expectedObj)
	expectedResourceData := MockForemanComputeResourceResourceData(expectedState)

	actualObj := api.ForemanComputeResource{}
	actualState := ForemanComputeResourceToInstanceState(actualObj)
	actualResourceData := MockForemanComputeResourceResourceData(actualState)

	setResourceDataFromForemanComputeResource(actualResourceData, &expectedObj)

	ForemanComputeResourceResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanComputeResourceCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanComputeResource{}
	obj.Id = rand.Intn(100)
	s := ForemanComputeResourceToInstanceState(obj)
	compute_resourcesURIById := ComputeResourcesURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanComputeResourceRead",
				crudFunc:     resourceForemanComputeResourceRead,
				resourceData: MockForemanComputeResourceResourceData(s),
			},
			expectedURI:    compute_resourcesURIById,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanComputeResourceRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanComputeResource{}
	obj.Id = rand.Intn(100)
	s := ForemanComputeResourceToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanComputeResourceRead",
			crudFunc:     resourceForemanComputeResourceRead,
			resourceData: MockForemanComputeResourceResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanComputeResourceStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanComputeResource{}
	obj.Id = rand.Intn(100)
	s := ForemanComputeResourceToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanComputeResourceRead",
			crudFunc:     resourceForemanComputeResourceRead,
			resourceData: MockForemanComputeResourceResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanComputeResourceEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanComputeResource{}
	obj.Id = rand.Intn(100)
	s := ForemanComputeResourceToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanComputeResourceRead",
			crudFunc:     resourceForemanComputeResourceRead,
			resourceData: MockForemanComputeResourceResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanComputeResourceMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanComputeResource()
	s := ForemanComputeResourceToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanComputeResourceRead",
				crudFunc:     resourceForemanComputeResourceRead,
				resourceData: MockForemanComputeResourceResourceData(s),
			},
			responseFile: ComputeResourcesTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanComputeResourceResourceDataFromFile(
				t,
				ComputeResourcesTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanComputeResourceResourceDataCompare,
		},
	}

}

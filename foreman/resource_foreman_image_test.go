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

const ImagesURI = api.FOREMAN_API_URL_PREFIX + "/compute_resources"
const ImagesTestDataPath = "testdata/1.11/image"

// Given a ForemanImage, create a mock instance state reference
func ForemanImageToInstanceState(obj api.ForemanImage) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanImage
	state.Attributes = map[string]string{
		"name":                obj.Name,
		"username":            obj.Username,
		"uuid":                obj.UUID,
		"architecture_id":     strconv.Itoa(obj.ArchitectureID),
		"compute_resource_id": strconv.Itoa(obj.ComputeResourceID),
		"operating_system_id": strconv.Itoa(obj.OperatingSystemID),
	}
	return &state
}

// Given a mock instance state for a ForemanImage resource, create a
// mock ResourceData reference.
func MockForemanImageResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanImage()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a  compute_resource
// ResourceData reference
func MockForemanImageResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanImage
	ParseJSONFile(t, path, &obj)
	s := ForemanImageToInstanceState(obj)
	return MockForemanImageResourceData(s)
}

// Creates a random ForemanImage struct
func RandForemanImage() api.ForemanImage {
	fo := RandForemanObject()
	return api.ForemanImage{
		ForemanObject:     fo,
		Name:              tfrand.String(10, tfrand.Lower),
		Username:          tfrand.String(10, tfrand.Lower),
		UUID:              tfrand.String(10, tfrand.Lower),
		ArchitectureID:    rand.Intn(100),
		OperatingSystemID: rand.Intn(100),
		ComputeResourceID: rand.Intn(100),
	}
}

// Compares two ResourceData references for a ForemanImage resource.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanImageResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanImage()
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
func TestImageUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanImage
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanImage UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanImage UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanImage
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanImage_Value(t *testing.T) {

	expectedObj := RandForemanImage()
	expectedState := ForemanImageToInstanceState(expectedObj)
	expectedResourceData := MockForemanImageResourceData(expectedState)

	actualObj := api.ForemanImage{}
	actualState := ForemanImageToInstanceState(actualObj)
	actualResourceData := MockForemanImageResourceData(actualState)

	setResourceDataFromForemanImage(actualResourceData, &expectedObj)

	ForemanImageResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanImageCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanImage{}
	obj.ComputeResourceID = rand.Intn(100)
	obj.Id = rand.Intn(100)
	s := ForemanImageToInstanceState(obj)
	imageURIById := ImagesURI + "/" + strconv.Itoa(obj.ComputeResourceID) + "/images/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanImageRead",
				crudFunc:     resourceForemanImageRead,
				resourceData: MockForemanImageResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    imageURIById,
					expectedMethod: http.MethodGet,
				},
			},
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanImageRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanImage{}
	obj.Id = rand.Intn(100)
	s := ForemanImageToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanImageRead",
			crudFunc:     resourceForemanImageRead,
			resourceData: MockForemanImageResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanImageStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanImage{}
	obj.Id = rand.Intn(100)
	s := ForemanImageToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanImageRead",
			crudFunc:     resourceForemanImageRead,
			resourceData: MockForemanImageResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanImageEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanImage{}
	obj.Id = rand.Intn(100)
	s := ForemanImageToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanImageRead",
			crudFunc:     resourceForemanImageRead,
			resourceData: MockForemanImageResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanImageMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanImage()
	s := ForemanImageToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanImageRead",
				crudFunc:     resourceForemanImageRead,
				resourceData: MockForemanImageResourceData(s),
			},
			responseFile: ImagesTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanImageResourceDataFromFile(
				t,
				ImagesTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanImageResourceDataCompare,
		},
	}

}

package foreman

import (
	"encoding/json"
	"fmt"
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

const ArchitecturesURI = api.FOREMAN_API_URL_PREFIX + "/architectures"
const ArchitecturesTestDataPath = "testdata/1.11/architectures"

// Given a ForemanArchitecture, create a mock instance state reference
func ForemanArchitectureToInstanceState(obj api.ForemanArchitecture) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanArchitecture
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["operatingsystem_ids.#"] = strconv.Itoa(len(obj.OperatingSystemIds))
	for idx, val := range obj.OperatingSystemIds {
		key := fmt.Sprintf("operatingsystem_ids.%d", idx)
		attr[key] = strconv.Itoa(val)
	}
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanArchitecture resource, create a
// mock ResourceData reference.
func MockForemanArchitectureResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanArchitecture()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates an architecture
// ResourceData reference
func MockForemanArchitectureResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanArchitecture
	ParseJSONFile(t, path, &obj)
	s := ForemanArchitectureToInstanceState(obj)
	return MockForemanArchitectureResourceData(s)
}

// Creates a random ForemanArchitecture struct
func RandForemanArchitecture() api.ForemanArchitecture {
	obj := api.ForemanArchitecture{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.OperatingSystemIds = tfrand.IntArrayUnique(5)

	return obj
}

// Compares two ResourceData references for a ForemanArchitecture resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanArchitectureResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanArchitecture()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

	var ok1, ok2 bool
	var attr1, attr2 interface{}

	attr1, ok1 = r1.Get("operatingsystem_ids").(*schema.Set)
	attr2, ok2 = r2.Get("operatingsystem_ids").(*schema.Set)
	if ok1 && ok2 {
		attr1Set := attr1.(*schema.Set)
		attr2Set := attr1.(*schema.Set)
		if !attr1Set.Equal(attr2Set) {
			t.Fatalf(
				"ResourceData reference differ in operatingsystem_ids. "+
					"[%v], [%v]",
				attr1Set.List(),
				attr2Set.List(),
			)
		}
	} else if (ok1 && !ok2) || (!ok1 && ok2) {
		t.Fatalf(
			"ResourceData references differ in operatingsystem_ids. "+
				"[%T], [%T]",
			attr1,
			attr2,
		)
	}

}

// -----------------------------------------------------------------------------
// UnmarshalJSON
// -----------------------------------------------------------------------------

// Ensures the JSON unmarshal correctly sets the base attributes from
// ForemanObject
func TestArchitectureUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanArchitecture
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanArchitecture UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanArchitecture UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanArchitecture
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanArchitecture
func TestBuildForemanArchitecture(t *testing.T) {

	expectedObj := RandForemanArchitecture()
	expectedState := ForemanArchitectureToInstanceState(expectedObj)
	expectedResourceData := MockForemanArchitectureResourceData(expectedState)

	actualObj := *buildForemanArchitecture(expectedResourceData)

	actualState := ForemanArchitectureToInstanceState(actualObj)
	actualResourceData := MockForemanArchitectureResourceData(actualState)

	ForemanArchitectureResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanArchitecture
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanArchitecture_Value(t *testing.T) {

	expectedObj := RandForemanArchitecture()
	expectedState := ForemanArchitectureToInstanceState(expectedObj)
	expectedResourceData := MockForemanArchitectureResourceData(expectedState)

	actualObj := api.ForemanArchitecture{}
	actualState := ForemanArchitectureToInstanceState(actualObj)
	actualResourceData := MockForemanArchitectureResourceData(actualState)

	setResourceDataFromForemanArchitecture(actualResourceData, &expectedObj)

	ForemanArchitectureResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanArchitectureCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanArchitecture{}
	obj.Id = rand.Intn(100)
	s := ForemanArchitectureToInstanceState(obj)
	architecturesURIById := ArchitecturesURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanArchitectureCreate",
				crudFunc:     resourceForemanArchitectureCreate,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			expectedURI:    ArchitecturesURI,
			expectedMethod: http.MethodPost,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanArchitectureRead",
				crudFunc:     resourceForemanArchitectureRead,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			expectedURI:    architecturesURIById,
			expectedMethod: http.MethodGet,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanArchitectureUpdate",
				crudFunc:     resourceForemanArchitectureUpdate,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			expectedURI:    architecturesURIById,
			expectedMethod: http.MethodPut,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanArchitectureDelete",
				crudFunc:     resourceForemanArchitectureDelete,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			expectedURI:    architecturesURIById,
			expectedMethod: http.MethodDelete,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanArchitectureRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanArchitecture{}
	obj.Id = rand.Intn(100)
	s := ForemanArchitectureToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanArchitectureRead",
			crudFunc:     resourceForemanArchitectureRead,
			resourceData: MockForemanArchitectureResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanArchitectureDelete",
			crudFunc:     resourceForemanArchitectureDelete,
			resourceData: MockForemanArchitectureResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanArchitectureRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanArchitecture{}
	obj.Id = rand.Intn(100)
	s := ForemanArchitectureToInstanceState(obj)

	rd := MockForemanArchitectureResourceData(s)
	obj = *buildForemanArchitecture(rd)
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanArchitectureCreate",
				crudFunc:     resourceForemanArchitectureCreate,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanArchitectureUpdate",
				crudFunc:     resourceForemanArchitectureUpdate,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanArchitectureStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanArchitecture{}
	obj.Id = rand.Intn(100)
	s := ForemanArchitectureToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanArchitectureCreate",
			crudFunc:     resourceForemanArchitectureCreate,
			resourceData: MockForemanArchitectureResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanArchitectureRead",
			crudFunc:     resourceForemanArchitectureRead,
			resourceData: MockForemanArchitectureResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanArchitectureUpdate",
			crudFunc:     resourceForemanArchitectureUpdate,
			resourceData: MockForemanArchitectureResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanArchitectureDelete",
			crudFunc:     resourceForemanArchitectureDelete,
			resourceData: MockForemanArchitectureResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanArchitectureEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanArchitecture{}
	obj.Id = rand.Intn(100)
	s := ForemanArchitectureToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanArchitectureCreate",
			crudFunc:     resourceForemanArchitectureCreate,
			resourceData: MockForemanArchitectureResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanArchitectureRead",
			crudFunc:     resourceForemanArchitectureRead,
			resourceData: MockForemanArchitectureResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanArchitectureUpdate",
			crudFunc:     resourceForemanArchitectureUpdate,
			resourceData: MockForemanArchitectureResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanArchitectureMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanArchitecture()
	s := ForemanArchitectureToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanArchitectureCreate",
				crudFunc:     resourceForemanArchitectureCreate,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			responseFile: ArchitecturesTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanArchitectureResourceDataFromFile(
				t,
				ArchitecturesTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanArchitectureResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanArchitectureRead",
				crudFunc:     resourceForemanArchitectureRead,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			responseFile: ArchitecturesTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanArchitectureResourceDataFromFile(
				t,
				ArchitecturesTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanArchitectureResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanArchitectureUpdate",
				crudFunc:     resourceForemanArchitectureUpdate,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			responseFile: ArchitecturesTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanArchitectureResourceDataFromFile(
				t,
				ArchitecturesTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanArchitectureResourceDataCompare,
		},
	}

}

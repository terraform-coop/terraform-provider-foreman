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

const MediasURI = api.FOREMAN_API_URL_PREFIX + "/media"
const MediasTestDataPath = "testdata/1.11/media"

// Given a ForemanMedia, create a mock instance state reference
func ForemanMediaToInstanceState(obj api.ForemanMedia) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanMedia
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["path"] = obj.Path
	attr["os_family"] = obj.OSFamily
	attr["operatingsystem_ids.#"] = strconv.Itoa(len(obj.OperatingSystemIds))
	for idx, val := range obj.OperatingSystemIds {
		key := fmt.Sprintf("operatingsystem_ids.%d", idx)
		attr[key] = strconv.Itoa(val)
	}
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanMedia resource, create a
// mock ResourceData reference.
func MockForemanMediaResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanMedia()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a media
// ResourceData reference
func MockForemanMediaResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanMedia
	ParseJSONFile(t, path, &obj)
	s := ForemanMediaToInstanceState(obj)
	return MockForemanMediaResourceData(s)
}

// Creates a random ForemanMedia struct
func RandForemanMedia() api.ForemanMedia {
	obj := api.ForemanMedia{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Path = tfrand.String(50, tfrand.Lower+"/:.")
	obj.OSFamily = tfrand.String(10, tfrand.Lower)
	obj.OperatingSystemIds = tfrand.IntArrayUnique(rand.Intn(5))

	return obj
}

// Compares two ResourceData references for a ForemanMedia resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanMediaResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanMedia()
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
func TestMediaUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanMedia
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanMedia UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanMedia UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanMedia
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanMedia
func TestBuildForemanMedia(t *testing.T) {

	expectedObj := RandForemanMedia()
	expectedState := ForemanMediaToInstanceState(expectedObj)
	expectedResourceData := MockForemanMediaResourceData(expectedState)

	actualObj := *buildForemanMedia(expectedResourceData)

	actualState := ForemanMediaToInstanceState(actualObj)
	actualResourceData := MockForemanMediaResourceData(actualState)

	ForemanMediaResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanMedia
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanMedia_Value(t *testing.T) {

	expectedObj := RandForemanMedia()
	expectedState := ForemanMediaToInstanceState(expectedObj)
	expectedResourceData := MockForemanMediaResourceData(expectedState)

	actualObj := api.ForemanMedia{}
	actualState := ForemanMediaToInstanceState(actualObj)
	actualResourceData := MockForemanMediaResourceData(actualState)

	setResourceDataFromForemanMedia(actualResourceData, &expectedObj)

	ForemanMediaResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanMediaCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanMedia{}
	obj.Id = rand.Intn(100)
	s := ForemanMediaToInstanceState(obj)
	mediasURIById := MediasURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanMediaCreate",
				crudFunc:     resourceForemanMediaCreate,
				resourceData: MockForemanMediaResourceData(s),
			},
			expectedURI:    MediasURI,
			expectedMethod: http.MethodPost,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanMediaRead",
				crudFunc:     resourceForemanMediaRead,
				resourceData: MockForemanMediaResourceData(s),
			},
			expectedURI:    mediasURIById,
			expectedMethod: http.MethodGet,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanMediaUpdate",
				crudFunc:     resourceForemanMediaUpdate,
				resourceData: MockForemanMediaResourceData(s),
			},
			expectedURI:    mediasURIById,
			expectedMethod: http.MethodPut,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanMediaDelete",
				crudFunc:     resourceForemanMediaDelete,
				resourceData: MockForemanMediaResourceData(s),
			},
			expectedURI:    mediasURIById,
			expectedMethod: http.MethodDelete,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanMediaRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanMedia{}
	obj.Id = rand.Intn(100)
	s := ForemanMediaToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanMediaRead",
			crudFunc:     resourceForemanMediaRead,
			resourceData: MockForemanMediaResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanMediaDelete",
			crudFunc:     resourceForemanMediaDelete,
			resourceData: MockForemanMediaResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanMediaRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanMedia{}
	obj.Id = rand.Intn(100)
	s := ForemanMediaToInstanceState(obj)

	rd := MockForemanMediaResourceData(s)
	obj = *buildForemanMedia(rd)
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanMediaCreate",
				crudFunc:     resourceForemanMediaCreate,
				resourceData: MockForemanMediaResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanMediaUpdate",
				crudFunc:     resourceForemanMediaUpdate,
				resourceData: MockForemanMediaResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanMediaStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanMedia{}
	obj.Id = rand.Intn(100)
	s := ForemanMediaToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanMediaCreate",
			crudFunc:     resourceForemanMediaCreate,
			resourceData: MockForemanMediaResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanMediaRead",
			crudFunc:     resourceForemanMediaRead,
			resourceData: MockForemanMediaResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanMediaUpdate",
			crudFunc:     resourceForemanMediaUpdate,
			resourceData: MockForemanMediaResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanMediaDelete",
			crudFunc:     resourceForemanMediaDelete,
			resourceData: MockForemanMediaResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanMediaEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanMedia{}
	obj.Id = rand.Intn(100)
	s := ForemanMediaToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanMediaCreate",
			crudFunc:     resourceForemanMediaCreate,
			resourceData: MockForemanMediaResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanMediaRead",
			crudFunc:     resourceForemanMediaRead,
			resourceData: MockForemanMediaResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanMediaUpdate",
			crudFunc:     resourceForemanMediaUpdate,
			resourceData: MockForemanMediaResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanMediaMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanMedia()
	s := ForemanMediaToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanMediaCreate",
				crudFunc:     resourceForemanMediaCreate,
				resourceData: MockForemanMediaResourceData(s),
			},
			responseFile: MediasTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanMediaResourceDataFromFile(
				t,
				MediasTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanMediaResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanMediaRead",
				crudFunc:     resourceForemanMediaRead,
				resourceData: MockForemanMediaResourceData(s),
			},
			responseFile: MediasTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanMediaResourceDataFromFile(
				t,
				MediasTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanMediaResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanMediaUpdate",
				crudFunc:     resourceForemanMediaUpdate,
				resourceData: MockForemanMediaResourceData(s),
			},
			responseFile: MediasTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanMediaResourceDataFromFile(
				t,
				MediasTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanMediaResourceDataCompare,
		},
	}

}

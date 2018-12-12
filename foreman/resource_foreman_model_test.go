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

const ModelsURI = api.FOREMAN_API_URL_PREFIX + "/models"
const ModelsTestDataPath = "testdata/1.11/models"

// Given a ForemanModel, create a mock instance state reference
func ForemanModelToInstanceState(obj api.ForemanModel) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanModel
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["info"] = obj.Info
	attr["vendor_class"] = obj.VendorClass
	attr["hardware_model"] = obj.HardwareModel
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanModel resource, create a
// mock ResourceData reference.
func MockForemanModelResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanModel()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a model
// ResourceData reference
func MockForemanModelResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanModel
	ParseJSONFile(t, path, &obj)
	s := ForemanModelToInstanceState(obj)
	return MockForemanModelResourceData(s)
}

// Creates a random ForemanModel struct
func RandForemanModel() api.ForemanModel {
	obj := api.ForemanModel{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Info = tfrand.String(50, tfrand.Lower+" .")
	obj.VendorClass = tfrand.String(10, tfrand.Lower+tfrand.Digit)
	obj.HardwareModel = tfrand.String(15, tfrand.Lower+tfrand.Digit)

	return obj
}

// Compares two ResourceData references for a ForemanModel resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanModelResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanModel()
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
func TestModelUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanModel
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanModel UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanModel UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanModel
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanModel
func TestBuildForemanModel(t *testing.T) {

	expectedObj := RandForemanModel()
	expectedState := ForemanModelToInstanceState(expectedObj)
	expectedResourceData := MockForemanModelResourceData(expectedState)

	actualObj := *buildForemanModel(expectedResourceData)

	actualState := ForemanModelToInstanceState(actualObj)
	actualResourceData := MockForemanModelResourceData(actualState)

	ForemanModelResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanModel
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanModel_Value(t *testing.T) {

	expectedObj := RandForemanModel()
	expectedState := ForemanModelToInstanceState(expectedObj)
	expectedResourceData := MockForemanModelResourceData(expectedState)

	actualObj := api.ForemanModel{}
	actualState := ForemanModelToInstanceState(actualObj)
	actualResourceData := MockForemanModelResourceData(actualState)

	setResourceDataFromForemanModel(actualResourceData, &expectedObj)

	ForemanModelResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanModelCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanModel{}
	obj.Id = rand.Intn(100)
	s := ForemanModelToInstanceState(obj)
	modelsURIById := ModelsURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanModelCreate",
				crudFunc:     resourceForemanModelCreate,
				resourceData: MockForemanModelResourceData(s),
			},
			expectedURI:    ModelsURI,
			expectedMethod: http.MethodPost,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanModelRead",
				crudFunc:     resourceForemanModelRead,
				resourceData: MockForemanModelResourceData(s),
			},
			expectedURI:    modelsURIById,
			expectedMethod: http.MethodGet,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanModelUpdate",
				crudFunc:     resourceForemanModelUpdate,
				resourceData: MockForemanModelResourceData(s),
			},
			expectedURI:    modelsURIById,
			expectedMethod: http.MethodPut,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanModelDelete",
				crudFunc:     resourceForemanModelDelete,
				resourceData: MockForemanModelResourceData(s),
			},
			expectedURI:    modelsURIById,
			expectedMethod: http.MethodDelete,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanModelRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanModel{}
	obj.Id = rand.Intn(100)
	s := ForemanModelToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanModelRead",
			crudFunc:     resourceForemanModelRead,
			resourceData: MockForemanModelResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanModelDelete",
			crudFunc:     resourceForemanModelDelete,
			resourceData: MockForemanModelResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanModelRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanModel{}
	obj.Id = rand.Intn(100)
	s := ForemanModelToInstanceState(obj)

	rd := MockForemanModelResourceData(s)
	obj = *buildForemanModel(rd)
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanModelCreate",
				crudFunc:     resourceForemanModelCreate,
				resourceData: MockForemanModelResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanModelUpdate",
				crudFunc:     resourceForemanModelUpdate,
				resourceData: MockForemanModelResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanModelStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanModel{}
	obj.Id = rand.Intn(100)
	s := ForemanModelToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanModelCreate",
			crudFunc:     resourceForemanModelCreate,
			resourceData: MockForemanModelResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanModelRead",
			crudFunc:     resourceForemanModelRead,
			resourceData: MockForemanModelResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanModelUpdate",
			crudFunc:     resourceForemanModelUpdate,
			resourceData: MockForemanModelResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanModelDelete",
			crudFunc:     resourceForemanModelDelete,
			resourceData: MockForemanModelResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanModelEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanModel{}
	obj.Id = rand.Intn(100)
	s := ForemanModelToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanModelCreate",
			crudFunc:     resourceForemanModelCreate,
			resourceData: MockForemanModelResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanModelRead",
			crudFunc:     resourceForemanModelRead,
			resourceData: MockForemanModelResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanModelUpdate",
			crudFunc:     resourceForemanModelUpdate,
			resourceData: MockForemanModelResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanModelMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanModel()
	s := ForemanModelToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanModelCreate",
				crudFunc:     resourceForemanModelCreate,
				resourceData: MockForemanModelResourceData(s),
			},
			responseFile: ModelsTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanModelResourceDataFromFile(
				t,
				ModelsTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanModelResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanModelRead",
				crudFunc:     resourceForemanModelRead,
				resourceData: MockForemanModelResourceData(s),
			},
			responseFile: ModelsTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanModelResourceDataFromFile(
				t,
				ModelsTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanModelResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanModelUpdate",
				crudFunc:     resourceForemanModelUpdate,
				resourceData: MockForemanModelResourceData(s),
			},
			responseFile: ModelsTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanModelResourceDataFromFile(
				t,
				ModelsTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanModelResourceDataCompare,
		},
	}

}

package foreman

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	tfrand "github.com/HanseMerkur/terraform-provider-utils/rand"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// -----------------------------------------------------------------------------
// Test Helper Functions
// -----------------------------------------------------------------------------

const OverrideValuesURI = api.FOREMAN_API_URL_PREFIX + "/smart_class_parameters/%d/override_values"
const OverrideValuesTestDataPath = "testdata/3.1.2/override_values"

// Given a ForemanOverrideValue, create a mock instance state reference
func ForemanOverrideValueToInstanceState(obj api.ForemanOverrideValue) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanOverrideValue
	attr := map[string]string{}
	attr["match.type"] = obj.MatchType
	attr["match.value"] = obj.MatchValue
	attr["omit"] = strconv.FormatBool(obj.Omit)
	attr["smart_class_parameter_id"] = strconv.Itoa(obj.SmartClassParameterId)
	attr["value"] = obj.Value
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanOverrideValue resource, create a
// mock ResourceData reference.
func MockForemanOverrideValueResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanOverrideValue()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a OverrideValue
// ResourceData reference
func MockForemanOverrideValueResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanOverrideValue
	ParseJSONFile(t, path, &obj)
	s := ForemanOverrideValueToInstanceState(obj)
	return MockForemanOverrideValueResourceData(s)
}

// Creates a random ForemanOverrideValue struct
func RandForemanOverrideValue() api.ForemanOverrideValue {
	obj := api.ForemanOverrideValue{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	match_types := [...]string{"fqdn", "hostgroup", "domain", "os"}
	omit_values := [...]bool{true, false}

	obj.MatchType = match_types[rand.Intn(3)]
	obj.MatchValue = tfrand.String(50, tfrand.Lower+" .")
	obj.Omit = omit_values[rand.Intn(len(omit_values))]
	obj.SmartClassParameterId = rand.Intn(20)
	obj.Value = tfrand.String(50, tfrand.Lower+" .")

	return obj
}

// Compares two ResourceData references for a ForemanOverrideValue resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanOverrideValueResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanOverrideValue()
	for key, value := range r.Schema {
		// Skip smart_class_paramater_id as it's not included in the server response
		if key == "smart_class_parameter_id" {
			continue
		}
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
func TestOverrideValueUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanOverrideValue
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanOverrideValue UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanOverrideValue UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanOverrideValue
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanOverrideValue
func TestBuildForemanOverrideValue(t *testing.T) {

	expectedObj := RandForemanOverrideValue()
	expectedState := ForemanOverrideValueToInstanceState(expectedObj)
	expectedResourceData := MockForemanOverrideValueResourceData(expectedState)

	actualObj := *buildForemanOverrideValue(expectedResourceData)

	actualState := ForemanOverrideValueToInstanceState(actualObj)
	actualResourceData := MockForemanOverrideValueResourceData(actualState)

	ForemanOverrideValueResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanOverrideValue
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanOverrideValue_Value(t *testing.T) {

	expectedObj := RandForemanOverrideValue()
	expectedState := ForemanOverrideValueToInstanceState(expectedObj)
	expectedResourceData := MockForemanOverrideValueResourceData(expectedState)

	actualObj := api.ForemanOverrideValue{}
	actualState := ForemanOverrideValueToInstanceState(actualObj)
	actualResourceData := MockForemanOverrideValueResourceData(actualState)

	setResourceDataFromForemanOverrideValue(actualResourceData, &expectedObj)

	ForemanOverrideValueResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanOverrideValueCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanOverrideValue{}
	obj.Id = rand.Intn(100)
	s := ForemanOverrideValueToInstanceState(obj)
	OverrideValuesURIById := fmt.Sprintf(OverrideValuesURI+"/%s", obj.SmartClassParameterId, strconv.Itoa(obj.Id))

	return []TestCaseCorrectURLAndMethod{
		{
			TestCase: TestCase{
				funcName:     "resourceForemanOverrideValueCreate",
				crudFunc:     resourceForemanOverrideValueCreate,
				resourceData: MockForemanOverrideValueResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    fmt.Sprintf(OverrideValuesURI, obj.SmartClassParameterId),
					expectedMethod: http.MethodPost,
				},
			},
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanOverrideValueRead",
				crudFunc:     resourceForemanOverrideValueRead,
				resourceData: MockForemanOverrideValueResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    OverrideValuesURIById,
					expectedMethod: http.MethodGet,
				},
			},
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanOverrideValueUpdate",
				crudFunc:     resourceForemanOverrideValueUpdate,
				resourceData: MockForemanOverrideValueResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    OverrideValuesURIById,
					expectedMethod: http.MethodPut,
				},
			},
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanOverrideValueDelete",
				crudFunc:     resourceForemanOverrideValueDelete,
				resourceData: MockForemanOverrideValueResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    OverrideValuesURIById,
					expectedMethod: http.MethodDelete,
				},
			},
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanOverrideValueRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanOverrideValue{}
	obj.Id = rand.Intn(100)
	s := ForemanOverrideValueToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanOverrideValueRead",
			crudFunc:     resourceForemanOverrideValueRead,
			resourceData: MockForemanOverrideValueResourceData(s),
		},
		{
			funcName:     "resourceForemanOverrideValueDelete",
			crudFunc:     resourceForemanOverrideValueDelete,
			resourceData: MockForemanOverrideValueResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanOverrideValueRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanOverrideValue{}
	obj.Id = rand.Intn(100)
	s := ForemanOverrideValueToInstanceState(obj)

	rd := MockForemanOverrideValueResourceData(s)
	obj = *buildForemanOverrideValue(rd)
	_, _, client := NewForemanAPIAndClient(api.ClientCredentials{}, api.ClientConfig{})
	reqData, _ := client.WrapJSONWithTaxonomy("override_value", obj)

	return []TestCaseRequestData{
		{
			TestCase: TestCase{
				funcName:     "resourceForemanOverrideValueCreate",
				crudFunc:     resourceForemanOverrideValueCreate,
				resourceData: MockForemanOverrideValueResourceData(s),
			},
			expectedData: reqData,
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanOverrideValueUpdate",
				crudFunc:     resourceForemanOverrideValueUpdate,
				resourceData: MockForemanOverrideValueResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanOverrideValueStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanOverrideValue{}
	obj.Id = rand.Intn(100)
	s := ForemanOverrideValueToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanOverrideValueCreate",
			crudFunc:     resourceForemanOverrideValueCreate,
			resourceData: MockForemanOverrideValueResourceData(s),
		},
		{
			funcName:     "resourceForemanOverrideValueRead",
			crudFunc:     resourceForemanOverrideValueRead,
			resourceData: MockForemanOverrideValueResourceData(s),
		},
		{
			funcName:     "resourceForemanOverrideValueUpdate",
			crudFunc:     resourceForemanOverrideValueUpdate,
			resourceData: MockForemanOverrideValueResourceData(s),
		},
		{
			funcName:     "resourceForemanOverrideValueDelete",
			crudFunc:     resourceForemanOverrideValueDelete,
			resourceData: MockForemanOverrideValueResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanOverrideValueEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanOverrideValue{}
	obj.Id = rand.Intn(100)
	s := ForemanOverrideValueToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanOverrideValueCreate",
			crudFunc:     resourceForemanOverrideValueCreate,
			resourceData: MockForemanOverrideValueResourceData(s),
		},
		{
			funcName:     "resourceForemanOverrideValueRead",
			crudFunc:     resourceForemanOverrideValueRead,
			resourceData: MockForemanOverrideValueResourceData(s),
		},
		{
			funcName:     "resourceForemanOverrideValueUpdate",
			crudFunc:     resourceForemanOverrideValueUpdate,
			resourceData: MockForemanOverrideValueResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanOverrideValueMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanOverrideValue()
	s := ForemanOverrideValueToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		{
			TestCase: TestCase{
				funcName:     "resourceForemanOverrideValueCreate",
				crudFunc:     resourceForemanOverrideValueCreate,
				resourceData: MockForemanOverrideValueResourceData(s),
			},
			responseFile: OverrideValuesTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanOverrideValueResourceDataFromFile(
				t,
				OverrideValuesTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanOverrideValueResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		{
			TestCase: TestCase{
				funcName:     "resourceForemanOverrideValueRead",
				crudFunc:     resourceForemanOverrideValueRead,
				resourceData: MockForemanOverrideValueResourceData(s),
			},
			responseFile: OverrideValuesTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanOverrideValueResourceDataFromFile(
				t,
				OverrideValuesTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanOverrideValueResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		{
			TestCase: TestCase{
				funcName:     "resourceForemanOverrideValueUpdate",
				crudFunc:     resourceForemanOverrideValueUpdate,
				resourceData: MockForemanOverrideValueResourceData(s),
			},
			responseFile: OverrideValuesTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanOverrideValueResourceDataFromFile(
				t,
				OverrideValuesTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanOverrideValueResourceDataCompare,
		},
	}

}

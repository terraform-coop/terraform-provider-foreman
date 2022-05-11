package foreman

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"testing"

	"github.com/HanseMerkur/terraform-provider-foreman/foreman/api"
	tfrand "github.com/HanseMerkur/terraform-provider-utils/rand"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// -----------------------------------------------------------------------------
// Test Helper Functions
// -----------------------------------------------------------------------------

const SmartClassParameterURI = "/foreman_puppet/api/puppetclasses/%d/smart_class_parameters"
const SmartClassParameterTestDataPath = "testdata/3.1.2/smart_class_parameters"

// Given a ForemanSmartClassParameter, create a mock instance state reference
func ForemanSmartClassParameterToInstanceState(obj api.ForemanSmartClassParameter) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanSmartClassParameter
	attr := map[string]string{}
	attr["parameter"] = obj.Parameter
	attr["puppetclass_id"] = strconv.Itoa(obj.PuppetClassId)
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanSmartClassParamter resource, create a
// mock ResourceData reference.
func MockForemanSmartClassParameterResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := dataSourceForemanSmartClassParameter()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a smart class paramter
// ResourceData reference
func MockForemanSmartClassParameterResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanSmartClassParameter
	ParseJSONFile(t, path, &obj)
	s := ForemanSmartClassParameterToInstanceState(obj)
	return MockForemanSmartClassParameterResourceData(s)
}

// Creates a random ForemanSmartClassParameter struct
func RandForemanSmartClassParameter() api.ForemanSmartClassParameter {
	obj := api.ForemanSmartClassParameter{}

	fo := RandForemanObject()
	obj.Parameter = tfrand.String(50, tfrand.Lower+" .")
	obj.PuppetClassId = rand.Intn(20)
	obj.ForemanObject = fo

	return obj
}

// Compares two ResourceData references for a ForemanModel resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanSmartClassParameterResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := dataSourceForemanSmartClassParameter()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanSmartClassParameterCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanSmartClassParameter()
	s := ForemanSmartClassParameterToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSmartClassParameterRead",
				crudFunc:     dataSourceForemanSmartClassParameterRead,
				resourceData: MockForemanSmartClassParameterResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    fmt.Sprintf(SmartClassParameterURI, obj.PuppetClassId),
					expectedMethod: http.MethodGet,
				},
			},
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanSmartClassParameterRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanSmartClassParameter()
	s := ForemanSmartClassParameterToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanSmartClassParameterRead",
			crudFunc:     dataSourceForemanSmartClassParameterRead,
			resourceData: MockForemanSmartClassParameterResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanSmartClassParameterStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanSmartClassParameter()
	s := ForemanSmartClassParameterToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanSmartClassParameterRead",
			crudFunc:     dataSourceForemanSmartClassParameterRead,
			resourceData: MockForemanSmartClassParameterResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanSmartClassParameterEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanSmartClassParameter()
	s := ForemanSmartClassParameterToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanSmartClassParameterRead",
			crudFunc:     dataSourceForemanSmartClassParameterRead,
			resourceData: MockForemanSmartClassParameterResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanSmartClassParameterMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanSmartClassParameter()
	s := ForemanSmartClassParameterToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSmartClassParameterRead",
				crudFunc:     dataSourceForemanSmartClassParameterRead,
				resourceData: MockForemanSmartClassParameterResourceData(s),
			},
			responseFile: SmartClassParameterTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSmartClassParameterRead",
				crudFunc:     dataSourceForemanSmartClassParameterRead,
				resourceData: MockForemanSmartClassParameterResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSmartClassParameterRead",
				crudFunc:     dataSourceForemanSmartClassParameterRead,
				resourceData: MockForemanSmartClassParameterResourceData(s),
			},
			responseFile: SmartClassParameterTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanSmartClassParameterResourceDataFromFile(
				t,
				SmartClassParameterTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanSmartClassParameterResourceDataCompare,
		},
	}

}

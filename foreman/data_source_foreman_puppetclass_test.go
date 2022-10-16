package foreman

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
)

// -----------------------------------------------------------------------------
// Test Helper Functions
// -----------------------------------------------------------------------------

const PuppetClassesURI = "/foreman_puppet/api/puppetclasses"
const PuppetClassessTestDataPath = "testdata/3.1.2/puppet_classes"

// Given a ForemanPuppetClass, create a mock instance state reference
func ForemanPuppetClassToInstanceState(obj api.ForemanPuppetClass) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanPuppetClass
	attr := map[string]string{}
	attr["name"] = obj.Name
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanModel resource, create a
// mock ResourceData reference.
func MockForemanPuppetClassResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := dataSourceForemanPuppetClass()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a model
// ResourceData reference
func MockForemanPuppetClasslResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanPuppetClass
	ParseJSONFile(t, path, &obj)
	s := ForemanPuppetClassToInstanceState(obj)
	return MockForemanPuppetClassResourceData(s)
}

// Creates a random ForemanPuppetClass struct
func RandForemanPuppetClass() api.ForemanPuppetClass {
	obj := api.ForemanPuppetClass{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	// The name must be hardcoded to match the test data
	obj.Name = "testing"

	return obj
}

// Creates a random nested ForemanPuppetClass struct
func RandForemanNestedPuppetClass() api.ForemanPuppetClass {
	obj := api.ForemanPuppetClass{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	// The name must be hardcoded to match the test data
	obj.Name = "apache::dev"

	return obj
}

// Compares two ResourceData references for a ForemanModel resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanPuppetClassResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := dataSourceForemanPuppetClass()
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
func DataSourceForemanPuppetClassCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanPuppetClass()
	s := ForemanPuppetClassToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanPuppetClassRead",
				crudFunc:     dataSourceForemanPuppetClassRead,
				resourceData: MockForemanPuppetClassResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    PuppetClassesURI,
					expectedMethod: http.MethodGet,
				},
			},
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanPuppetClassRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanPuppetClass()
	s := ForemanPuppetClassToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "dataSourceForemanPuppetClassRead",
			crudFunc:     dataSourceForemanPuppetClassRead,
			resourceData: MockForemanPuppetClassResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanPuppetClassStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanPuppetClass()
	s := ForemanPuppetClassToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "dataSourceForemanPuppetClassRead",
			crudFunc:     dataSourceForemanPuppetClassRead,
			resourceData: MockForemanPuppetClassResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanPuppetClassEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanPuppetClass()
	s := ForemanPuppetClassToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "dataSourceForemanPuppetClassRead",
			crudFunc:     dataSourceForemanPuppetClassRead,
			resourceData: MockForemanPuppetClassResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanPuppetClassMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanPuppetClass()
	s := ForemanPuppetClassToInstanceState(obj)

	obj_nested := RandForemanNestedPuppetClass()
	s_nested := ForemanPuppetClassToInstanceState(obj_nested)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanPuppetClassRead",
				crudFunc:     dataSourceForemanPuppetClassRead,
				resourceData: MockForemanPuppetClassResourceData(s),
			},
			responseFile: PuppetClassessTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanPuppetClassRead",
				crudFunc:     dataSourceForemanPuppetClassRead,
				resourceData: MockForemanPuppetClassResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanPuppetClassRead",
				crudFunc:     dataSourceForemanPuppetClassRead,
				resourceData: MockForemanPuppetClassResourceData(s),
			},
			responseFile: PuppetClassessTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanPuppetClasslResourceDataFromFile(
				t,
				PuppetClassessTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanPuppetClassResourceDataCompare,
		},
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanPuppetClassRead",
				crudFunc:     dataSourceForemanPuppetClassRead,
				resourceData: MockForemanPuppetClassResourceData(s_nested),
			},
			responseFile: PuppetClassessTestDataPath + "/query_response_colon.json",
			returnError:  false,
			expectedResourceData: MockForemanPuppetClasslResourceDataFromFile(
				t,
				PuppetClassessTestDataPath+"/query_response_colon_state.json",
			),
			compareFunc: ForemanPuppetClassResourceDataCompare,
		},
	}

}

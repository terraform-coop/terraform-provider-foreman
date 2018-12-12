package foreman

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// -----------------------------------------------------------------------------
// Test Helper Functions
// -----------------------------------------------------------------------------

const TemplateKindsURI = api.FOREMAN_API_URL_PREFIX + "/template_kinds"
const TemplateKindsTestDataPath = "testdata/1.11/template_kinds"

// Given a ForemanTemplateKind, create a mock instance state reference
func ForemanTemplateKindToInstanceState(obj api.ForemanTemplateKind) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanTemplateKind
	attr := map[string]string{}
	attr["name"] = obj.Name
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanTemplateKind resource, create a
// mock ResourceData reference.
func MockForemanTemplateKindResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := dataSourceForemanTemplateKind()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates an architecture
// ResourceData reference
func MockForemanTemplateKindResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanTemplateKind
	ParseJSONFile(t, path, &obj)
	s := ForemanTemplateKindToInstanceState(obj)
	return MockForemanTemplateKindResourceData(s)
}

// Creates a random ForemanTemplateKind struct
func RandForemanTemplateKind() api.ForemanTemplateKind {
	obj := api.ForemanTemplateKind{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	return obj
}

// Compares two ResourceData references for a ForemanTemplateKind resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanTemplateKindResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := dataSourceForemanTemplateKind()
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
func DataSourceForemanTemplateKindCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanTemplateKind()
	s := ForemanTemplateKindToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanTemplateKindRead",
				crudFunc:     dataSourceForemanTemplateKindRead,
				resourceData: MockForemanTemplateKindResourceData(s),
			},
			expectedURI:    TemplateKindsURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanTemplateKindRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanTemplateKind()
	s := ForemanTemplateKindToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanTemplateKindRead",
			crudFunc:     dataSourceForemanTemplateKindRead,
			resourceData: MockForemanTemplateKindResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanTemplateKindStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanTemplateKind()
	s := ForemanTemplateKindToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanTemplateKindRead",
			crudFunc:     dataSourceForemanTemplateKindRead,
			resourceData: MockForemanTemplateKindResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanTemplateKindEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanTemplateKind()
	s := ForemanTemplateKindToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanTemplateKindRead",
			crudFunc:     dataSourceForemanTemplateKindRead,
			resourceData: MockForemanTemplateKindResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanTemplateKindMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanTemplateKind()
	s := ForemanTemplateKindToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanTemplateKindRead",
				crudFunc:     dataSourceForemanTemplateKindRead,
				resourceData: MockForemanTemplateKindResourceData(s),
			},
			responseFile: TemplateKindsTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanTemplateKindRead",
				crudFunc:     dataSourceForemanTemplateKindRead,
				resourceData: MockForemanTemplateKindResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanTemplateKindRead",
				crudFunc:     dataSourceForemanTemplateKindRead,
				resourceData: MockForemanTemplateKindResourceData(s),
			},
			responseFile: TemplateKindsTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanTemplateKindResourceDataFromFile(
				t,
				TemplateKindsTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanTemplateKindResourceDataCompare,
		},
	}

}

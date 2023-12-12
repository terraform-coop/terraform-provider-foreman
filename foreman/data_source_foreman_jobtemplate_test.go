package foreman

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"net/http"
	"testing"
)

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanJobTemplateCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanJobTemplate()
	s := ForemanJobTemplateToInstanceState(obj)

	// Currently, "read" uses "query" and therefore does not call the /api/job_templates/<id> endpoint
	jobTemplateURI := "/api/job_templates"

	return []TestCaseCorrectURLAndMethod{
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanJobTemplateRead",
				crudFunc:     dataSourceForemanJobTemplateRead,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    jobTemplateURI,
					expectedMethod: http.MethodGet,
				},
			},
		},
	}

}

func crudTestCaseInner() []TestCase {
	obj := RandForemanJobTemplate()
	s := ForemanJobTemplateToInstanceState(obj)
	return []TestCase{
		{
			funcName:     "dataSourceForemanJobTemplateRead",
			crudFunc:     dataSourceForemanJobTemplateRead,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanJobTemplateRequestDataEmptyTestCases(t *testing.T) []TestCase {
	return crudTestCaseInner()
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanJobTemplateStatusCodeTestCases(t *testing.T) []TestCase {
	return crudTestCaseInner()
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanJobTemplateEmptyResponseTestCases(t *testing.T) []TestCase {
	return crudTestCaseInner()
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanJobTemplateMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanJobTemplate()
	s := ForemanJobTemplateToInstanceState(obj)

	tc := func(s *terraform.InstanceState) TestCase {
		return TestCase{
			funcName:     "dataSourceForemanJobTemplateRead",
			crudFunc:     dataSourceForemanJobTemplateRead,
			resourceData: MockForemanJobTemplateResourceData(s),
		}
	}

	const testDataPathJobTemplates = "testdata/3.6/job_template"

	// Test fail on zero, fail on multiple and successful handling of exactlyon one match
	return []TestCaseMockResponse{
		{
			TestCase:     tc(s),
			responseFile: testDataPathJobTemplates + "/query_response_multi.json",
			returnError:  true,
		},

		{
			TestCase:     tc(s),
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},

		{
			TestCase:     tc(s),
			responseFile: testDataPathJobTemplates + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanJobTemplateResourceDataFromFile(
				t,
				testDataPathJobTemplates+"/query_response_single_state.json",
			),
			compareFunc: ForemanJobTemplateResourceDataCompare,
		},
	}

}

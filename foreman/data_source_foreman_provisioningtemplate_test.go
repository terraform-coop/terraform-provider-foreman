package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanProvisioningTemplateCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanProvisioningTemplate()
	s := ForemanProvisioningTemplateToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanProvisioningTemplateRead",
				crudFunc:     dataSourceForemanProvisioningTemplateRead,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			expectedURI:    ProvisioningTemplatesURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanProvisioningTemplateRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanProvisioningTemplate()
	s := ForemanProvisioningTemplateToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanProvisioningTemplateRead",
			crudFunc:     dataSourceForemanProvisioningTemplateRead,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanProvisioningTemplateStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanProvisioningTemplate()
	s := ForemanProvisioningTemplateToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanProvisioningTemplateRead",
			crudFunc:     dataSourceForemanProvisioningTemplateRead,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanProvisioningTemplateEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanProvisioningTemplate()
	s := ForemanProvisioningTemplateToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanProvisioningTemplateRead",
			crudFunc:     dataSourceForemanProvisioningTemplateRead,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanProvisioningTemplateMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanProvisioningTemplate()
	s := ForemanProvisioningTemplateToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanProvisioningTemplateRead",
				crudFunc:     dataSourceForemanProvisioningTemplateRead,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			responseFile: ProvisioningTemplatesTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanProvisioningTemplateRead",
				crudFunc:     dataSourceForemanProvisioningTemplateRead,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanProvisioningTemplateRead",
				crudFunc:     dataSourceForemanProvisioningTemplateRead,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			responseFile: ProvisioningTemplatesTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanProvisioningTemplateResourceDataFromFile(
				t,
				ProvisioningTemplatesTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanProvisioningTemplateResourceDataCompare,
		},
	}

}

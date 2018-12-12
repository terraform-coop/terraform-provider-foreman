package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanEnvironmentCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanEnvironment()
	s := ForemanEnvironmentToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanEnvironmentRead",
				crudFunc:     dataSourceForemanEnvironmentRead,
				resourceData: MockForemanEnvironmentResourceData(s),
			},
			expectedURI:    EnvironmentsURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanEnvironmentRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanEnvironment()
	s := ForemanEnvironmentToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanEnvironmentRead",
			crudFunc:     dataSourceForemanEnvironmentRead,
			resourceData: MockForemanEnvironmentResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanEnvironmentStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanEnvironment()
	s := ForemanEnvironmentToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanEnvironmentRead",
			crudFunc:     dataSourceForemanEnvironmentRead,
			resourceData: MockForemanEnvironmentResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanEnvironmentEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanEnvironment()
	s := ForemanEnvironmentToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanEnvironmentRead",
			crudFunc:     dataSourceForemanEnvironmentRead,
			resourceData: MockForemanEnvironmentResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanEnvironmentMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanEnvironment()
	s := ForemanEnvironmentToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanEnvironmentRead",
				crudFunc:     dataSourceForemanEnvironmentRead,
				resourceData: MockForemanEnvironmentResourceData(s),
			},
			responseFile: EnvironmentsTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanEnvironmentRead",
				crudFunc:     dataSourceForemanEnvironmentRead,
				resourceData: MockForemanEnvironmentResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanEnvironmentRead",
				crudFunc:     dataSourceForemanEnvironmentRead,
				resourceData: MockForemanEnvironmentResourceData(s),
			},
			responseFile: EnvironmentsTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanEnvironmentResourceDataFromFile(
				t,
				EnvironmentsTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanEnvironmentResourceDataCompare,
		},
	}

}

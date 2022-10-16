package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanArchitectureCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanArchitecture()
	s := ForemanArchitectureToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanArchitectureRead",
				crudFunc:     dataSourceForemanArchitectureRead,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    ArchitecturesURI,
					expectedMethod: http.MethodGet,
				},
			},
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanArchitectureRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanArchitecture()
	s := ForemanArchitectureToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "dataSourceForemanArchitectureRead",
			crudFunc:     dataSourceForemanArchitectureRead,
			resourceData: MockForemanArchitectureResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanArchitectureStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanArchitecture()
	s := ForemanArchitectureToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "dataSourceForemanArchitectureRead",
			crudFunc:     dataSourceForemanArchitectureRead,
			resourceData: MockForemanArchitectureResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanArchitectureEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanArchitecture()
	s := ForemanArchitectureToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "dataSourceForemanArchitectureRead",
			crudFunc:     dataSourceForemanArchitectureRead,
			resourceData: MockForemanArchitectureResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanArchitectureMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanArchitecture()
	s := ForemanArchitectureToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanArchitectureRead",
				crudFunc:     dataSourceForemanArchitectureRead,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			responseFile: ArchitecturesTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanArchitectureRead",
				crudFunc:     dataSourceForemanArchitectureRead,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanArchitectureRead",
				crudFunc:     dataSourceForemanArchitectureRead,
				resourceData: MockForemanArchitectureResourceData(s),
			},
			responseFile: ArchitecturesTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanArchitectureResourceDataFromFile(
				t,
				ArchitecturesTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanArchitectureResourceDataCompare,
		},
	}

}

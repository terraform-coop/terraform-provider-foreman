package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanOperatingSystemCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanOperatingSystem()
	s := ForemanOperatingSystemToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanOperatingSystemRead",
				crudFunc:     dataSourceForemanOperatingSystemRead,
				resourceData: MockForemanOperatingSystemResourceData(s),
			},
			expectedURI:    OperatingSystemsURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanOperatingSystemRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanOperatingSystem()
	s := ForemanOperatingSystemToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanOperatingSystemRead",
			crudFunc:     dataSourceForemanOperatingSystemRead,
			resourceData: MockForemanOperatingSystemResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanOperatingSystemStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanOperatingSystem()
	s := ForemanOperatingSystemToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanOperatingSystemRead",
			crudFunc:     dataSourceForemanOperatingSystemRead,
			resourceData: MockForemanOperatingSystemResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanOperatingSystemEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanOperatingSystem()
	s := ForemanOperatingSystemToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanOperatingSystemRead",
			crudFunc:     dataSourceForemanOperatingSystemRead,
			resourceData: MockForemanOperatingSystemResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanOperatingSystemMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanOperatingSystem()
	s := ForemanOperatingSystemToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanOperatingSystemRead",
				crudFunc:     dataSourceForemanOperatingSystemRead,
				resourceData: MockForemanOperatingSystemResourceData(s),
			},
			responseFile: OperatingSystemsTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanOperatingSystemRead",
				crudFunc:     dataSourceForemanOperatingSystemRead,
				resourceData: MockForemanOperatingSystemResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanOperatingSystemRead",
				crudFunc:     dataSourceForemanOperatingSystemRead,
				resourceData: MockForemanOperatingSystemResourceData(s),
			},
			responseFile: OperatingSystemsTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanOperatingSystemResourceDataFromFile(
				t,
				OperatingSystemsTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanOperatingSystemResourceDataCompare,
		},
	}

}

package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanMediaCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanMedia()
	s := ForemanMediaToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanMediaRead",
				crudFunc:     dataSourceForemanMediaRead,
				resourceData: MockForemanMediaResourceData(s),
			},
			expectedURI:    MediasURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanMediaRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanMedia()
	s := ForemanMediaToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanMediaRead",
			crudFunc:     dataSourceForemanMediaRead,
			resourceData: MockForemanMediaResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanMediaStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanMedia()
	s := ForemanMediaToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanMediaRead",
			crudFunc:     dataSourceForemanMediaRead,
			resourceData: MockForemanMediaResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanMediaEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanMedia()
	s := ForemanMediaToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanMediaRead",
			crudFunc:     dataSourceForemanMediaRead,
			resourceData: MockForemanMediaResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanMediaMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanMedia()
	s := ForemanMediaToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanMediaRead",
				crudFunc:     dataSourceForemanMediaRead,
				resourceData: MockForemanMediaResourceData(s),
			},
			responseFile: MediasTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanMediaRead",
				crudFunc:     dataSourceForemanMediaRead,
				resourceData: MockForemanMediaResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanMediaRead",
				crudFunc:     dataSourceForemanMediaRead,
				resourceData: MockForemanMediaResourceData(s),
			},
			responseFile: MediasTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanMediaResourceDataFromFile(
				t,
				MediasTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanMediaResourceDataCompare,
		},
	}

}

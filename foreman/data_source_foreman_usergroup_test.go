package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanUsergroupCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanUsergroup()
	s := ForemanUsergroupToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanUsergroupRead",
				crudFunc:     dataSourceForemanUsergroupRead,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			expectedURI:    UsergroupsURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanUsergroupRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanUsergroup()
	s := ForemanUsergroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanUsergroupRead",
			crudFunc:     dataSourceForemanUsergroupRead,
			resourceData: MockForemanUsergroupResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanUsergroupStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanUsergroup()
	s := ForemanUsergroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanUsergroupRead",
			crudFunc:     dataSourceForemanUsergroupRead,
			resourceData: MockForemanUsergroupResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanUsergroupEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanUsergroup()
	s := ForemanUsergroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanUsergroupRead",
			crudFunc:     dataSourceForemanUsergroupRead,
			resourceData: MockForemanUsergroupResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanUsergroupMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanUsergroup()
	s := ForemanUsergroupToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanUsergroupRead",
				crudFunc:     dataSourceForemanUsergroupRead,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			responseFile: UsergroupsTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanUsergroupRead",
				crudFunc:     dataSourceForemanUsergroupRead,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanUsergroupRead",
				crudFunc:     dataSourceForemanUsergroupRead,
				resourceData: MockForemanUsergroupResourceData(s),
			},
			responseFile: UsergroupsTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanUsergroupResourceDataFromFile(
				t,
				UsergroupsTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanUsergroupResourceDataCompare,
		},
	}

}

package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanHostgroupCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanHostgroup()
	s := ForemanHostgroupToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanHostgroupRead",
				crudFunc:     dataSourceForemanHostgroupRead,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			expectedURI:    HostgroupsURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanHostgroupRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanHostgroup()
	s := ForemanHostgroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanHostgroupRead",
			crudFunc:     dataSourceForemanHostgroupRead,
			resourceData: MockForemanHostgroupResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanHostgroupStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanHostgroup()
	s := ForemanHostgroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanHostgroupRead",
			crudFunc:     dataSourceForemanHostgroupRead,
			resourceData: MockForemanHostgroupResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanHostgroupEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanHostgroup()
	s := ForemanHostgroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanHostgroupRead",
			crudFunc:     dataSourceForemanHostgroupRead,
			resourceData: MockForemanHostgroupResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanHostgroupMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanHostgroup()
	s := ForemanHostgroupToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanHostgroupRead",
				crudFunc:     dataSourceForemanHostgroupRead,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			responseFile: HostgroupsTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanHostgroupRead",
				crudFunc:     dataSourceForemanHostgroupRead,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanHostgroupRead",
				crudFunc:     dataSourceForemanHostgroupRead,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			responseFile: HostgroupsTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanHostgroupResourceDataFromFile(
				t,
				HostgroupsTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanHostgroupResourceDataCompare,
		},
	}

}

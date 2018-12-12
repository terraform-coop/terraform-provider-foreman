package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanSubnetCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanSubnet()
	s := ForemanSubnetToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSubnetRead",
				crudFunc:     dataSourceForemanSubnetRead,
				resourceData: MockForemanSubnetResourceData(s),
			},
			expectedURI:    SubnetsURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanSubnetRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanSubnet()
	s := ForemanSubnetToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanSubnetRead",
			crudFunc:     dataSourceForemanSubnetRead,
			resourceData: MockForemanSubnetResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanSubnetStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanSubnet()
	s := ForemanSubnetToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanSubnetRead",
			crudFunc:     dataSourceForemanSubnetRead,
			resourceData: MockForemanSubnetResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanSubnetEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanSubnet()
	s := ForemanSubnetToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanSubnetRead",
			crudFunc:     dataSourceForemanSubnetRead,
			resourceData: MockForemanSubnetResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanSubnetMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanSubnet()
	s := ForemanSubnetToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSubnetRead",
				crudFunc:     dataSourceForemanSubnetRead,
				resourceData: MockForemanSubnetResourceData(s),
			},
			responseFile: SubnetsTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSubnetRead",
				crudFunc:     dataSourceForemanSubnetRead,
				resourceData: MockForemanSubnetResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSubnetRead",
				crudFunc:     dataSourceForemanSubnetRead,
				resourceData: MockForemanSubnetResourceData(s),
			},
			responseFile: SubnetsTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanSubnetResourceDataFromFile(
				t,
				SubnetsTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanSubnetResourceDataCompare,
		},
	}

}

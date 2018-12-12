package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanSmartProxyCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanSmartProxy()
	s := ForemanSmartProxyToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSmartProxyRead",
				crudFunc:     dataSourceForemanSmartProxyRead,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			expectedURI:    SmartProxiesURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanSmartProxyRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanSmartProxy()
	s := ForemanSmartProxyToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanSmartProxyRead",
			crudFunc:     dataSourceForemanSmartProxyRead,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanSmartProxyStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanSmartProxy()
	s := ForemanSmartProxyToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanSmartProxyRead",
			crudFunc:     dataSourceForemanSmartProxyRead,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanSmartProxyEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanSmartProxy()
	s := ForemanSmartProxyToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanSmartProxyRead",
			crudFunc:     dataSourceForemanSmartProxyRead,
			resourceData: MockForemanSmartProxyResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanSmartProxyMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanSmartProxy()
	s := ForemanSmartProxyToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSmartProxyRead",
				crudFunc:     dataSourceForemanSmartProxyRead,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			responseFile: SmartProxiesTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSmartProxyRead",
				crudFunc:     dataSourceForemanSmartProxyRead,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSmartProxyRead",
				crudFunc:     dataSourceForemanSmartProxyRead,
				resourceData: MockForemanSmartProxyResourceData(s),
			},
			responseFile: SmartProxiesTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanSmartProxyResourceDataFromFile(
				t,
				SmartProxiesTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanSmartProxyResourceDataCompare,
		},
	}

}

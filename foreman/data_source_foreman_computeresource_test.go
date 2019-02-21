package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanComputeResourceCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanComputeResource()
	s := ForemanComputeResourceToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanComputeResourceRead",
				crudFunc:     dataSourceForemanComputeResourceRead,
				resourceData: MockForemanComputeResourceResourceData(s),
			},
			expectedURI:    ComputeResourcesURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanComputeResourceRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanComputeResource()
	s := ForemanComputeResourceToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanComputeResourceRead",
			crudFunc:     dataSourceForemanComputeResourceRead,
			resourceData: MockForemanComputeResourceResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanComputeResourceStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanComputeResource()
	s := ForemanComputeResourceToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanComputeResourceRead",
			crudFunc:     dataSourceForemanComputeResourceRead,
			resourceData: MockForemanComputeResourceResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanComputeResourceEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanComputeResource()
	s := ForemanComputeResourceToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanComputeResourceRead",
			crudFunc:     dataSourceForemanComputeResourceRead,
			resourceData: MockForemanComputeResourceResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanComputeResourceMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanComputeResource()
	s := ForemanComputeResourceToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanComputeResourceRead",
				crudFunc:     dataSourceForemanComputeResourceRead,
				resourceData: MockForemanComputeResourceResourceData(s),
			},
			responseFile: ComputeResourcesTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanComputeResourceRead",
				crudFunc:     dataSourceForemanComputeResourceRead,
				resourceData: MockForemanComputeResourceResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanComputeResourceRead",
				crudFunc:     dataSourceForemanComputeResourceRead,
				resourceData: MockForemanComputeResourceResourceData(s),
			},
			responseFile: ComputeResourcesTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanComputeResourceResourceDataFromFile(
				t,
				ComputeResourcesTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanComputeResourceResourceDataCompare,
		},
	}

}

package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanModelCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanModel()
	s := ForemanModelToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanModelRead",
				crudFunc:     dataSourceForemanModelRead,
				resourceData: MockForemanModelResourceData(s),
			},
			expectedURI:    ModelsURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanModelRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanModel()
	s := ForemanModelToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanModelRead",
			crudFunc:     dataSourceForemanModelRead,
			resourceData: MockForemanModelResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanModelStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanModel()
	s := ForemanModelToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanModelRead",
			crudFunc:     dataSourceForemanModelRead,
			resourceData: MockForemanModelResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanModelEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanModel()
	s := ForemanModelToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanModelRead",
			crudFunc:     dataSourceForemanModelRead,
			resourceData: MockForemanModelResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanModelMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanModel()
	s := ForemanModelToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanModelRead",
				crudFunc:     dataSourceForemanModelRead,
				resourceData: MockForemanModelResourceData(s),
			},
			responseFile: ModelsTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanModelRead",
				crudFunc:     dataSourceForemanModelRead,
				resourceData: MockForemanModelResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanModelRead",
				crudFunc:     dataSourceForemanModelRead,
				resourceData: MockForemanModelResourceData(s),
			},
			responseFile: ModelsTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanModelResourceDataFromFile(
				t,
				ModelsTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanModelResourceDataCompare,
		},
	}

}

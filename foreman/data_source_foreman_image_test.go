package foreman

import (
	"net/http"
	"strconv"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanImageCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanImage()
	s := ForemanImageToInstanceState(obj)
	imageURIByResource := ImagesURI + "/" + strconv.Itoa(obj.ComputeResourceID) + "/images/"

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanImageRead",
				crudFunc:     dataSourceForemanImageRead,
				resourceData: MockForemanImageResourceData(s),
			},
			expectedURI:    imageURIByResource,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanImageRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanImage()
	s := ForemanImageToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanImageRead",
			crudFunc:     dataSourceForemanImageRead,
			resourceData: MockForemanImageResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanImageStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanImage()
	s := ForemanImageToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanImageRead",
			crudFunc:     dataSourceForemanImageRead,
			resourceData: MockForemanImageResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanImageEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanImage()
	s := ForemanImageToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanImageRead",
			crudFunc:     dataSourceForemanImageRead,
			resourceData: MockForemanImageResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanImageMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanImage()
	s := ForemanImageToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanImageRead",
				crudFunc:     dataSourceForemanImageRead,
				resourceData: MockForemanImageResourceData(s),
			},
			responseFile: ImagesTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanImageRead",
				crudFunc:     dataSourceForemanImageRead,
				resourceData: MockForemanImageResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanImageRead",
				crudFunc:     dataSourceForemanImageRead,
				resourceData: MockForemanImageResourceData(s),
			},
			responseFile: ImagesTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanImageResourceDataFromFile(
				t,
				ImagesTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanImageResourceDataCompare,
		},
	}

}

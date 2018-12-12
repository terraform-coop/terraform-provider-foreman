package foreman

import (
	"net/http"
	"testing"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanDomainCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanDomain()
	s := ForemanDomainToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanDomainRead",
				crudFunc:     dataSourceForemanDomainRead,
				resourceData: MockForemanDomainResourceData(s),
			},
			expectedURI:    DomainsURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanDomainRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanDomain()
	s := ForemanDomainToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanDomainRead",
			crudFunc:     dataSourceForemanDomainRead,
			resourceData: MockForemanDomainResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanDomainStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanDomain()
	s := ForemanDomainToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanDomainRead",
			crudFunc:     dataSourceForemanDomainRead,
			resourceData: MockForemanDomainResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanDomainEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanDomain()
	s := ForemanDomainToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanDomainRead",
			crudFunc:     dataSourceForemanDomainRead,
			resourceData: MockForemanDomainResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanDomainMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanDomain()
	s := ForemanDomainToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with more than one search result for the data
		// source read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanDomainRead",
				crudFunc:     dataSourceForemanDomainRead,
				resourceData: MockForemanDomainResourceData(s),
			},
			responseFile: DomainsTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
		// If the server responds with zero search results for the data source
		// read, then the operation should return an error
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanDomainRead",
				crudFunc:     dataSourceForemanDomainRead,
				resourceData: MockForemanDomainResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
		// If the server responds with exactly one search result for the data source
		// read, then the operation should succeed and the attributes of the
		// ResourceData should be set properly.
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanDomainRead",
				crudFunc:     dataSourceForemanDomainRead,
				resourceData: MockForemanDomainResourceData(s),
			},
			responseFile: DomainsTestDataPath + "/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanDomainResourceDataFromFile(
				t,
				DomainsTestDataPath+"/query_response_single_state.json",
			),
			compareFunc: ForemanDomainResourceDataCompare,
		},
	}

}

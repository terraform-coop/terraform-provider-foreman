package foreman

import (
	"net/http"
	"testing"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func DataSourceForemanPartitionTableCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := RandForemanPartitionTable()
	s := ForemanPartitionTableToInstanceState(obj)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "dataSourceForemanPartitionTableRead",
				crudFunc:     dataSourceForemanPartitionTableRead,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			expectedURI:    PartitionTablesURI,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func DataSourceForemanPartitionTableRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := RandForemanPartitionTable()
	s := ForemanPartitionTableToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanPartitionTableRead",
			crudFunc:     dataSourceForemanPartitionTableRead,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func DataSourceForemanPartitionTableStatusCodeTestCases(t *testing.T) []TestCase {

	obj := RandForemanPartitionTable()
	s := ForemanPartitionTableToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanPartitionTableRead",
			crudFunc:     dataSourceForemanPartitionTableRead,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func DataSourceForemanPartitionTableEmptyResponseTestCases(t *testing.T) []TestCase {

	obj := RandForemanPartitionTable()
	s := ForemanPartitionTableToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "dataSourceForemanPartitionTableRead",
			crudFunc:     dataSourceForemanPartitionTableRead,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func DataSourceForemanPartitionTableMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanPartitionTable()
	s := ForemanPartitionTableToInstanceState(obj)

	testCases := []TestCaseMockResponse{}

	var expectedObj api.ForemanPartitionTable
	var expectedState *terraform.InstanceState
	var expectedData *schema.ResourceData

	// If the server responds with more than one search result for the data
	// source read, then the operation should return an error
	testCases = append(
		testCases,
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanPartitionTableRead",
				crudFunc:     dataSourceForemanPartitionTableRead,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			responseFile: PartitionTablesTestDataPath + "/query_response_multi.json",
			returnError:  true,
		},
	)

	// If the server responds with zero search results for the data source
	// read, then the operation should return an error
	testCases = append(
		testCases,
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanPartitionTableRead",
				crudFunc:     dataSourceForemanPartitionTableRead,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			responseFile: TestDataPath + "/query_response_zero.json",
			returnError:  true,
		},
	)

	// SEE: resource_foreman_partitiontable.go#setResourceDataFromForemanPartitionTable
	ParseJSONFile(t, PartitionTablesTestDataPath+"/query_response_single_state.json", &expectedObj)
	expectedObj.Snippet = obj.Snippet
	expectedObj.Locked = obj.Locked
	expectedObj.AuditComment = obj.AuditComment
	expectedObj.HostgroupIds = obj.HostgroupIds
	expectedObj.HostIds = obj.HostIds
	expectedState = ForemanPartitionTableToInstanceState(expectedObj)
	expectedData = MockForemanPartitionTableResourceData(expectedState)
	// If the server responds with exactly one search result for the data source
	// read, then the operation should succeed and the attributes of the
	// ResourceData should be set properly.
	testCases = append(
		testCases,
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "dataSourceForemanPartitionTableRead",
				crudFunc:     dataSourceForemanPartitionTableRead,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			responseFile:         PartitionTablesTestDataPath + "/query_response_single.json",
			returnError:          false,
			expectedResourceData: expectedData,
			compareFunc:          ForemanPartitionTableResourceDataCompare,
		},
	)

	return testCases

}

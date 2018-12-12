package foreman

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	tfrand "github.com/wayfair/terraform-provider-utils/rand"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// -----------------------------------------------------------------------------
// Test Helper Functions
// -----------------------------------------------------------------------------

const PartitionTablesURI = api.FOREMAN_API_URL_PREFIX + "/ptables"
const PartitionTablesTestDataPath = "testdata/1.11/ptables"

// Given a ForemanPartitionTable, create a mock instance state reference
func ForemanPartitionTableToInstanceState(obj api.ForemanPartitionTable) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanPartitionTable
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["layout"] = obj.Layout
	attr["snippet"] = fmt.Sprintf("%t", obj.Snippet)
	attr["audit_comment"] = obj.AuditComment
	attr["locked"] = fmt.Sprintf("%t", obj.Locked)
	attr["os_family"] = obj.OSFamily

	attr["operatingsystem_ids.#"] = strconv.Itoa(len(obj.OperatingSystemIds))
	for idx, val := range obj.OperatingSystemIds {
		key := fmt.Sprintf("operatingsystem_ids.%d", idx)
		attr[key] = strconv.Itoa(val)
	}

	attr["hostgroup_ids.#"] = strconv.Itoa(len(obj.HostgroupIds))
	for idx, val := range obj.HostgroupIds {
		key := fmt.Sprintf("hostgroup_ids.%d", idx)
		attr[key] = strconv.Itoa(val)
	}

	attr["host_ids.#"] = strconv.Itoa(len(obj.HostIds))
	for idx, val := range obj.HostIds {
		key := fmt.Sprintf("host_ids.%d", idx)
		attr[key] = strconv.Itoa(val)
	}

	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanPartitionTable resource, create a
// mock ResourceData reference.
func MockForemanPartitionTableResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanPartitionTable()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a model
// ResourceData reference
func MockForemanPartitionTableResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanPartitionTable
	ParseJSONFile(t, path, &obj)
	s := ForemanPartitionTableToInstanceState(obj)
	return MockForemanPartitionTableResourceData(s)
}

// Creates a random ForemanPartitionTable struct
func RandForemanPartitionTable() api.ForemanPartitionTable {
	obj := api.ForemanPartitionTable{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Layout = tfrand.String(100, tfrand.Lower+" \r\n.")
	obj.Snippet = rand.Intn(2) > 0
	obj.AuditComment = tfrand.String(100, tfrand.Lower+". ")
	obj.Locked = rand.Intn(2) > 0
	obj.OSFamily = tfrand.String(10, tfrand.Lower)

	obj.OperatingSystemIds = tfrand.IntArrayUnique(rand.Intn(5))
	obj.HostgroupIds = tfrand.IntArrayUnique(rand.Intn(5))
	obj.HostIds = tfrand.IntArrayUnique(rand.Intn(5))

	return obj
}

// Compares two ResourceData references for a ForemanPartitionTable resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanPartitionTableResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

	// compare IDs
	if r1.Id() != r2.Id() {
		t.Fatalf(
			"ResourceData references differ in Id. [%s], [%s]",
			r1.Id(),
			r2.Id(),
		)
	}

	// build the attribute map
	m := map[string]schema.ValueType{}
	r := resourceForemanPartitionTable()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

	var ok1, ok2 bool
	var attr1, attr2 interface{}

	attr1, ok1 = r1.Get("operatingsystem_ids").(*schema.Set)
	attr2, ok2 = r2.Get("operatingsystem_ids").(*schema.Set)
	if ok1 && ok2 {
		attr1Set := attr1.(*schema.Set)
		attr2Set := attr1.(*schema.Set)
		if !attr1Set.Equal(attr2Set) {
			t.Fatalf(
				"ResourceData reference differ in operatingsystem_ids. "+
					"[%v], [%v]",
				attr1Set.List(),
				attr2Set.List(),
			)
		}
	} else if (ok1 && !ok2) || (!ok1 && ok2) {
		t.Fatalf(
			"ResourceData references differ in operatingsystem_ids. "+
				"[%T], [%T]",
			attr1,
			attr2,
		)
	}

	attr1, ok1 = r1.Get("hostgroup_ids").(*schema.Set)
	attr2, ok2 = r2.Get("hostgroup_ids").(*schema.Set)
	if ok1 && ok2 {
		attr1Set := attr1.(*schema.Set)
		attr2Set := attr1.(*schema.Set)
		if !attr1Set.Equal(attr2Set) {
			t.Fatalf(
				"ResourceData reference differ in hostgroup_ids. "+
					"[%v], [%v]",
				attr1Set.List(),
				attr2Set.List(),
			)
		}
	} else if (ok1 && !ok2) || (!ok1 && ok2) {
		t.Fatalf(
			"ResourceData references differ in hostgroup_ids. "+
				"[%T], [%T]",
			attr1,
			attr2,
		)
	}

	attr1, ok1 = r1.Get("host_ids").(*schema.Set)
	attr2, ok2 = r2.Get("host_ids").(*schema.Set)
	if ok1 && ok2 {
		attr1Set := attr1.(*schema.Set)
		attr2Set := attr1.(*schema.Set)
		if !attr1Set.Equal(attr2Set) {
			t.Fatalf(
				"ResourceData reference differ in host_ids. "+
					"[%v], [%v]",
				attr1Set.List(),
				attr2Set.List(),
			)
		}
	} else if (ok1 && !ok2) || (!ok1 && ok2) {
		t.Fatalf(
			"ResourceData references differ in host_ids. "+
				"[%T], [%T]",
			attr1,
			attr2,
		)
	}

}

// -----------------------------------------------------------------------------
// UnmarshalJSON
// -----------------------------------------------------------------------------

// Ensures the JSON unmarshal correctly sets the base attributes from
// ForemanObject
func TestPartitionTableUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanPartitionTable
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanPartitionTable UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanPartitionTable UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanPartitionTable
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanPartitionTable
func TestBuildForemanPartitionTable(t *testing.T) {

	expectedObj := RandForemanPartitionTable()
	expectedState := ForemanPartitionTableToInstanceState(expectedObj)
	expectedResourceData := MockForemanPartitionTableResourceData(expectedState)

	actualObj := *buildForemanPartitionTable(expectedResourceData)

	actualState := ForemanPartitionTableToInstanceState(actualObj)
	actualResourceData := MockForemanPartitionTableResourceData(actualState)

	ForemanPartitionTableResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanPartitionTable
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanPartitionTable_Value(t *testing.T) {

	expectedObj := RandForemanPartitionTable()
	expectedState := ForemanPartitionTableToInstanceState(expectedObj)
	expectedResourceData := MockForemanPartitionTableResourceData(expectedState)

	actualObj := api.ForemanPartitionTable{}

	// SEE: resource_foreman_partitiontable.go#setResourceDataFromForemanPartitionTable
	actualObj.Snippet = expectedObj.Snippet
	actualObj.Locked = expectedObj.Locked
	actualObj.AuditComment = expectedObj.AuditComment
	actualObj.HostgroupIds = expectedObj.HostgroupIds
	actualObj.HostIds = expectedObj.HostIds

	actualState := ForemanPartitionTableToInstanceState(actualObj)
	actualResourceData := MockForemanPartitionTableResourceData(actualState)

	setResourceDataFromForemanPartitionTable(actualResourceData, &expectedObj)

	ForemanPartitionTableResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanPartitionTableCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanPartitionTable{}
	obj.Id = rand.Intn(100)
	s := ForemanPartitionTableToInstanceState(obj)
	partitionTablesURIById := PartitionTablesURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanPartitionTableCreate",
				crudFunc:     resourceForemanPartitionTableCreate,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			expectedURI:    PartitionTablesURI,
			expectedMethod: http.MethodPost,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanPartitionTableRead",
				crudFunc:     resourceForemanPartitionTableRead,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			expectedURI:    partitionTablesURIById,
			expectedMethod: http.MethodGet,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanPartitionTableUpdate",
				crudFunc:     resourceForemanPartitionTableUpdate,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			expectedURI:    partitionTablesURIById,
			expectedMethod: http.MethodPut,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanPartitionTableDelete",
				crudFunc:     resourceForemanPartitionTableDelete,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			expectedURI:    partitionTablesURIById,
			expectedMethod: http.MethodDelete,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanPartitionTableRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanPartitionTable{}
	obj.Id = rand.Intn(100)
	s := ForemanPartitionTableToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanPartitionTableRead",
			crudFunc:     resourceForemanPartitionTableRead,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanPartitionTableDelete",
			crudFunc:     resourceForemanPartitionTableDelete,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanPartitionTableRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanPartitionTable{}
	obj.Id = rand.Intn(100)
	s := ForemanPartitionTableToInstanceState(obj)

	rd := MockForemanPartitionTableResourceData(s)
	obj = *buildForemanPartitionTable(rd)
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanPartitionTableCreate",
				crudFunc:     resourceForemanPartitionTableCreate,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanPartitionTableUpdate",
				crudFunc:     resourceForemanPartitionTableUpdate,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanPartitionTableStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanPartitionTable{}
	obj.Id = rand.Intn(100)
	s := ForemanPartitionTableToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanPartitionTableCreate",
			crudFunc:     resourceForemanPartitionTableCreate,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanPartitionTableRead",
			crudFunc:     resourceForemanPartitionTableRead,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanPartitionTableUpdate",
			crudFunc:     resourceForemanPartitionTableUpdate,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanPartitionTableDelete",
			crudFunc:     resourceForemanPartitionTableDelete,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanPartitionTableEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanPartitionTable{}
	obj.Id = rand.Intn(100)
	s := ForemanPartitionTableToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanPartitionTableCreate",
			crudFunc:     resourceForemanPartitionTableCreate,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanPartitionTableRead",
			crudFunc:     resourceForemanPartitionTableRead,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanPartitionTableUpdate",
			crudFunc:     resourceForemanPartitionTableUpdate,
			resourceData: MockForemanPartitionTableResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanPartitionTableMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanPartitionTable()
	s := ForemanPartitionTableToInstanceState(obj)

	testCases := []TestCaseMockResponse{}

	var expectedObj api.ForemanPartitionTable
	var expectedState *terraform.InstanceState
	var expectedData *schema.ResourceData

	// SEE: resource_foreman_partitiontable.go#setResourceDataFromForemanPartitionTable
	ParseJSONFile(t, PartitionTablesTestDataPath+"/create_response.json", &expectedObj)
	expectedObj.Snippet = obj.Snippet
	expectedObj.Locked = obj.Locked
	expectedObj.AuditComment = obj.AuditComment
	expectedObj.HostgroupIds = obj.HostgroupIds
	expectedObj.HostIds = obj.HostIds
	expectedState = ForemanPartitionTableToInstanceState(expectedObj)
	expectedData = MockForemanPartitionTableResourceData(expectedState)
	// If the server responds with a proper create response, the operation
	// should succeed and the ResourceData's attributes should be updated
	// to server's response
	testCases = append(
		testCases,
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanPartitionTableCreate",
				crudFunc:     resourceForemanPartitionTableCreate,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			responseFile:         PartitionTablesTestDataPath + "/create_response.json",
			returnError:          false,
			expectedResourceData: expectedData,
			compareFunc:          ForemanPartitionTableResourceDataCompare,
		},
	)

	// SEE: resource_foreman_partitiontable.go#setResourceDataFromForemanPartitionTable
	ParseJSONFile(t, PartitionTablesTestDataPath+"/read_response.json", &expectedObj)
	expectedObj.Snippet = obj.Snippet
	expectedObj.Locked = obj.Locked
	expectedObj.AuditComment = obj.AuditComment
	expectedObj.HostgroupIds = obj.HostgroupIds
	expectedObj.HostIds = obj.HostIds
	expectedState = ForemanPartitionTableToInstanceState(expectedObj)
	expectedData = MockForemanPartitionTableResourceData(expectedState)
	// If the server responds with a proper create response, the operation
	// should succeed and the ResourceData's attributes should be updated
	// to server's response
	testCases = append(
		testCases,
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanPartitionTableRead",
				crudFunc:     resourceForemanPartitionTableRead,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			responseFile:         PartitionTablesTestDataPath + "/read_response.json",
			returnError:          false,
			expectedResourceData: expectedData,
			compareFunc:          ForemanPartitionTableResourceDataCompare,
		},
	)

	// SEE: resource_foreman_partitiontable.go#setResourceDataFromForemanPartitionTable
	ParseJSONFile(t, PartitionTablesTestDataPath+"/update_response.json", &expectedObj)
	expectedObj.Snippet = obj.Snippet
	expectedObj.Locked = obj.Locked
	expectedObj.AuditComment = obj.AuditComment
	expectedObj.HostgroupIds = obj.HostgroupIds
	expectedObj.HostIds = obj.HostIds
	expectedState = ForemanPartitionTableToInstanceState(expectedObj)
	expectedData = MockForemanPartitionTableResourceData(expectedState)
	// If the server responds with a proper create response, the operation
	// should succeed and the ResourceData's attributes should be updated
	// to server's response
	testCases = append(
		testCases,
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanPartitionTableUpdate",
				crudFunc:     resourceForemanPartitionTableUpdate,
				resourceData: MockForemanPartitionTableResourceData(s),
			},
			responseFile:         PartitionTablesTestDataPath + "/update_response.json",
			returnError:          false,
			expectedResourceData: expectedData,
			compareFunc:          ForemanPartitionTableResourceDataCompare,
		},
	)

	return testCases

}

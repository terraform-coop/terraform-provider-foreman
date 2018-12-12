package foreman

import (
	"encoding/json"
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

const HostgroupsURI = api.FOREMAN_API_URL_PREFIX + "/hostgroups"
const HostgroupsTestDataPath = "testdata/1.11/hostgroups"

// Given a ForemanHostgroup, create a mock instance state reference
func ForemanHostgroupToInstanceState(obj api.ForemanHostgroup) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanHostgroup
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["title"] = obj.Title
	attr["architecture_id"] = strconv.Itoa(obj.ArchitectureId)
	attr["compute_profile_id"] = strconv.Itoa(obj.ComputeProfileId)
	attr["domain_id"] = strconv.Itoa(obj.DomainId)
	attr["environment_id"] = strconv.Itoa(obj.EnvironmentId)
	attr["medium_id"] = strconv.Itoa(obj.MediaId)
	attr["operatingsystem_id"] = strconv.Itoa(obj.OperatingSystemId)
	attr["parent_id"] = strconv.Itoa(obj.ParentId)
	attr["ptable_id"] = strconv.Itoa(obj.PartitionTableId)
	attr["puppet_ca_proxy_id"] = strconv.Itoa(obj.PuppetCAProxyId)
	attr["puppet_proxy_id"] = strconv.Itoa(obj.PuppetProxyId)
	attr["realm_id"] = strconv.Itoa(obj.RealmId)
	attr["subnet_id"] = strconv.Itoa(obj.SubnetId)
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanHostgroup resource, create a
// mock ResourceData reference.
func MockForemanHostgroupResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanHostgroup()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a hostgroup
// ResourceData reference
func MockForemanHostgroupResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanHostgroup
	ParseJSONFile(t, path, &obj)
	s := ForemanHostgroupToInstanceState(obj)
	return MockForemanHostgroupResourceData(s)
}

// Creates a random ForemanHostgroup struct
func RandForemanHostgroup() api.ForemanHostgroup {
	obj := api.ForemanHostgroup{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Title = tfrand.String(15, tfrand.Lower+"/")
	obj.ArchitectureId = rand.Intn(100)
	obj.ComputeProfileId = rand.Intn(100)
	obj.DomainId = rand.Intn(100)
	obj.EnvironmentId = rand.Intn(100)
	obj.MediaId = rand.Intn(100)
	obj.OperatingSystemId = rand.Intn(100)
	obj.ParentId = rand.Intn(100)
	obj.PartitionTableId = rand.Intn(100)
	obj.PuppetCAProxyId = rand.Intn(100)
	obj.PuppetProxyId = rand.Intn(100)
	obj.RealmId = rand.Intn(100)
	obj.SubnetId = rand.Intn(100)

	return obj
}

// Compares two ResourceData references for a ForemanHostgroup resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanHostgroupResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanHostgroup()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

}

// -----------------------------------------------------------------------------
// UnmarshalJSON
// -----------------------------------------------------------------------------

// Ensures the JSON unmarshal correctly sets the base attributes from
// ForemanObject
func TestHostgroupUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanHostgroup
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanHostgroup UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanHostgroup UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanHostgroup
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanHostgroup
func TestBuildForemanHostgroup(t *testing.T) {

	expectedObj := RandForemanHostgroup()
	expectedState := ForemanHostgroupToInstanceState(expectedObj)
	expectedResourceData := MockForemanHostgroupResourceData(expectedState)

	actualObj := *buildForemanHostgroup(expectedResourceData)

	actualState := ForemanHostgroupToInstanceState(actualObj)
	actualResourceData := MockForemanHostgroupResourceData(actualState)

	ForemanHostgroupResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanHostgroup
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanHostgroup_Value(t *testing.T) {

	expectedObj := RandForemanHostgroup()
	expectedState := ForemanHostgroupToInstanceState(expectedObj)
	expectedResourceData := MockForemanHostgroupResourceData(expectedState)

	actualObj := api.ForemanHostgroup{}
	actualState := ForemanHostgroupToInstanceState(actualObj)
	actualResourceData := MockForemanHostgroupResourceData(actualState)

	setResourceDataFromForemanHostgroup(actualResourceData, &expectedObj)

	ForemanHostgroupResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanHostgroupCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanHostgroup{}
	obj.Id = rand.Intn(100)
	s := ForemanHostgroupToInstanceState(obj)
	hostgroupsURIById := HostgroupsURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHostgroupCreate",
				crudFunc:     resourceForemanHostgroupCreate,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			expectedURI:    HostgroupsURI,
			expectedMethod: http.MethodPost,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHostgroupRead",
				crudFunc:     resourceForemanHostgroupRead,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			expectedURI:    hostgroupsURIById,
			expectedMethod: http.MethodGet,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHostgroupUpdate",
				crudFunc:     resourceForemanHostgroupUpdate,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			expectedURI:    hostgroupsURIById,
			expectedMethod: http.MethodPut,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHostgroupDelete",
				crudFunc:     resourceForemanHostgroupDelete,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			expectedURI:    hostgroupsURIById,
			expectedMethod: http.MethodDelete,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanHostgroupRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanHostgroup{}
	obj.Id = rand.Intn(100)
	s := ForemanHostgroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanHostgroupRead",
			crudFunc:     resourceForemanHostgroupRead,
			resourceData: MockForemanHostgroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostgroupDelete",
			crudFunc:     resourceForemanHostgroupDelete,
			resourceData: MockForemanHostgroupResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanHostgroupRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanHostgroup{}
	obj.Id = rand.Intn(100)
	s := ForemanHostgroupToInstanceState(obj)

	rd := MockForemanHostgroupResourceData(s)
	obj = *buildForemanHostgroup(rd)
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanHostgroupCreate",
				crudFunc:     resourceForemanHostgroupCreate,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanHostgroupUpdate",
				crudFunc:     resourceForemanHostgroupUpdate,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanHostgroupStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanHostgroup{}
	obj.Id = rand.Intn(100)
	s := ForemanHostgroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanHostgroupCreate",
			crudFunc:     resourceForemanHostgroupCreate,
			resourceData: MockForemanHostgroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostgroupRead",
			crudFunc:     resourceForemanHostgroupRead,
			resourceData: MockForemanHostgroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostgroupUpdate",
			crudFunc:     resourceForemanHostgroupUpdate,
			resourceData: MockForemanHostgroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostgroupDelete",
			crudFunc:     resourceForemanHostgroupDelete,
			resourceData: MockForemanHostgroupResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanHostgroupEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanHostgroup{}
	obj.Id = rand.Intn(100)
	s := ForemanHostgroupToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanHostgroupCreate",
			crudFunc:     resourceForemanHostgroupCreate,
			resourceData: MockForemanHostgroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostgroupRead",
			crudFunc:     resourceForemanHostgroupRead,
			resourceData: MockForemanHostgroupResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostgroupUpdate",
			crudFunc:     resourceForemanHostgroupUpdate,
			resourceData: MockForemanHostgroupResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanHostgroupMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanHostgroup()
	s := ForemanHostgroupToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanHostgroupCreate",
				crudFunc:     resourceForemanHostgroupCreate,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			responseFile: HostgroupsTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanHostgroupResourceDataFromFile(
				t,
				HostgroupsTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanHostgroupResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanHostgroupRead",
				crudFunc:     resourceForemanHostgroupRead,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			responseFile: HostgroupsTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanHostgroupResourceDataFromFile(
				t,
				HostgroupsTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanHostgroupResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanHostgroupUpdate",
				crudFunc:     resourceForemanHostgroupUpdate,
				resourceData: MockForemanHostgroupResourceData(s),
			},
			responseFile: HostgroupsTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanHostgroupResourceDataFromFile(
				t,
				HostgroupsTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanHostgroupResourceDataCompare,
		},
	}

}

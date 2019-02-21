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

const HostsURI = api.FOREMAN_API_URL_PREFIX + "/hosts"
const HostsTestDataPath = "testdata/1.11/hosts"

// Given a ForemanHost, create a mock instance state reference
func ForemanHostToInstanceState(obj api.ForemanHost) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanHost
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["domain_id"] = strconv.Itoa(obj.DomainId)
	attr["environment_id"] = strconv.Itoa(obj.EnvironmentId)
	attr["hostgroup_id"] = strconv.Itoa(obj.HostgroupId)
	attr["operatingsystem_id"] = strconv.Itoa(obj.OperatingSystemId)
	attr["medium_id"] = strconv.Itoa(obj.MediumId)
	attr["image_id"] = strconv.Itoa(obj.ImageId)
	attr["interfaces_attributes.#"] = strconv.Itoa(len(obj.InterfacesAttributes))
	for idx, val := range obj.InterfacesAttributes {
		key := fmt.Sprintf("interfaces_attributes.%d.id", idx)
		attr[key] = strconv.Itoa(val.Id)
		key = fmt.Sprintf("interfaces_attributes.%d.ip", idx)
		attr[key] = val.IP
		key = fmt.Sprintf("interfaces_attributes.%d.mac", idx)
		attr[key] = val.MAC
		key = fmt.Sprintf("interfaces_attributes.%d.subnet_id", idx)
		attr[key] = strconv.Itoa(val.SubnetId)
		key = fmt.Sprintf("interfaces_attributes.%d.identifier", idx)
		attr[key] = val.Identifier
		key = fmt.Sprintf("interfaces_attributes.%d.username", idx)
		attr[key] = val.Username
		key = fmt.Sprintf("interfaces_attributes.%d.password", idx)
		attr[key] = val.Password
		key = fmt.Sprintf("interfaces_attributes.%d.type", idx)
		attr[key] = val.Type
		key = fmt.Sprintf("interfaces_attributes.%d.provider", idx)
		attr[key] = val.Provider
		key = fmt.Sprintf("interfaces_attributes.%d.compute_attributes", idx)
		jsonAttr, _ := json.Marshal(val.ComputeAttributes)
		attr[key] = string(jsonAttr)
	}
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanHost resource, create a
// mock ResourceData reference.
func MockForemanHostResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanHost()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a host
// ResourceData reference
func MockForemanHostResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanHost
	ParseJSONFile(t, path, &obj)
	s := ForemanHostToInstanceState(obj)
	return MockForemanHostResourceData(s)
}

// Creates a random ForemanHost struct
func RandForemanHost() api.ForemanHost {
	obj := api.ForemanHost{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Build = rand.Intn(2) > 0
	obj.OperatingSystemId = rand.Intn(100)
	obj.DomainId = rand.Intn(100)
	obj.HostgroupId = rand.Intn(100)
	obj.EnvironmentId = rand.Intn(100)
	obj.MediumId = rand.Intn(100)
	obj.ImageId = rand.Intn(100)

	obj.InterfacesAttributes = make([]api.ForemanInterfacesAttribute, rand.Intn(5))
	for idx, _ := range obj.InterfacesAttributes {
		compAttr := make(map[string]interface{})
		for fil := 0; fil < rand.Intn(3); fil++ {
			compAttr[tfrand.String(5, tfrand.Lower)] = tfrand.String(5, tfrand.Lower)
		}
		obj.InterfacesAttributes[idx] = api.ForemanInterfacesAttribute{
			Id:                rand.Intn(100),
			SubnetId:          rand.Intn(100),
			Identifier:        tfrand.String(10, tfrand.Lower),
			Name:              tfrand.String(10, tfrand.Lower),
			Username:          tfrand.String(10, tfrand.Lower),
			Password:          tfrand.String(10, tfrand.Lower),
			IP:                tfrand.IPv4Str(tfrand.IPv4PrivateClassCStart, tfrand.IPv4PrivateClassCMask),
			MAC:               tfrand.MACAddr48Str(":"),
			Type:              tfrand.String(12, tfrand.Lower),
			Provider:          tfrand.String(12, tfrand.Lower),
			ComputeAttributes: compAttr,
		}
	}

	return obj
}

// Compares two ResourceData references for a ForemanHost resource.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanHostResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanHost()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

	var ok1, ok2 bool
	var attr1, attr2 interface{}

	attr1, ok1 = r1.Get("interfaces_attributes").(*schema.Set)
	attr2, ok2 = r2.Get("interfaces_attributes").(*schema.Set)
	if ok1 && ok2 {
		attr1Set := attr1.(*schema.Set)
		attr2Set := attr1.(*schema.Set)
		if !attr1Set.Equal(attr2Set) {
			t.Fatalf(
				"ResourceData reference differ in interfaces_attributes. "+
					"[%v], [%v]",
				attr1Set.List(),
				attr2Set.List(),
			)
		}
	} else if (ok1 && !ok2) || (!ok1 && ok2) {
		t.Fatalf(
			"ResourceData reference differ in interfaces_attributes. "+
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
func TestHostUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanHost
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanHost UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanHost UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanHost
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanHost
func TestBuildForemanHost(t *testing.T) {

	expectedObj := RandForemanHost()
	expectedState := ForemanHostToInstanceState(expectedObj)
	expectedResourceData := MockForemanHostResourceData(expectedState)

	actualObj := *buildForemanHost(expectedResourceData)

	actualState := ForemanHostToInstanceState(actualObj)
	actualResourceData := MockForemanHostResourceData(actualState)

	ForemanHostResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanHost
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanHost_Value(t *testing.T) {

	expectedObj := RandForemanHost()
	expectedState := ForemanHostToInstanceState(expectedObj)
	expectedResourceData := MockForemanHostResourceData(expectedState)

	actualObj := api.ForemanHost{}
	actualState := ForemanHostToInstanceState(actualObj)
	actualResourceData := MockForemanHostResourceData(actualState)

	setResourceDataFromForemanHost(actualResourceData, &expectedObj)

	ForemanHostResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanHostCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanHost{}
	obj.Id = rand.Intn(100)
	s := ForemanHostToInstanceState(obj)
	hostsURIById := HostsURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHostCreate",
				crudFunc:     resourceForemanHostCreate,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedURI:    HostsURI,
			expectedMethod: http.MethodPost,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHostRead",
				crudFunc:     resourceForemanHostRead,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedURI:    hostsURIById,
			expectedMethod: http.MethodGet,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHostUpdate",
				crudFunc:     resourceForemanHostUpdate,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedURI:    hostsURIById,
			expectedMethod: http.MethodPut,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanHostDelete",
				crudFunc:     resourceForemanHostDelete,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedURI:    hostsURIById,
			expectedMethod: http.MethodDelete,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanHostRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanHost{}
	obj.Id = rand.Intn(100)
	s := ForemanHostToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanHostRead",
			crudFunc:     resourceForemanHostRead,
			resourceData: MockForemanHostResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostDelete",
			crudFunc:     resourceForemanHostDelete,
			resourceData: MockForemanHostResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanHostRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanHost{}
	obj.Id = rand.Intn(100)
	s := ForemanHostToInstanceState(obj)

	rd := MockForemanHostResourceData(s)
	obj = *buildForemanHost(rd)
	// NOTE(ALL): See note in Create and Update functions for build flag
	//   override
	obj.Build = true
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanHostCreate",
				crudFunc:     resourceForemanHostCreate,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanHostUpdate",
				crudFunc:     resourceForemanHostUpdate,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanHostStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanHost{}
	obj.Id = rand.Intn(100)
	s := ForemanHostToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanHostCreate",
			crudFunc:     resourceForemanHostCreate,
			resourceData: MockForemanHostResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostRead",
			crudFunc:     resourceForemanHostRead,
			resourceData: MockForemanHostResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostUpdate",
			crudFunc:     resourceForemanHostUpdate,
			resourceData: MockForemanHostResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostDelete",
			crudFunc:     resourceForemanHostDelete,
			resourceData: MockForemanHostResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanHostEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanHost{}
	obj.Id = rand.Intn(100)
	s := ForemanHostToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanHostCreate",
			crudFunc:     resourceForemanHostCreate,
			resourceData: MockForemanHostResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostRead",
			crudFunc:     resourceForemanHostRead,
			resourceData: MockForemanHostResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanHostUpdate",
			crudFunc:     resourceForemanHostUpdate,
			resourceData: MockForemanHostResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanHostMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanHost()
	s := ForemanHostToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanHostCreate",
				crudFunc:     resourceForemanHostCreate,
				resourceData: MockForemanHostResourceData(s),
			},
			responseFile: HostsTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanHostResourceDataFromFile(
				t,
				HostsTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanHostResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanHostRead",
				crudFunc:     resourceForemanHostRead,
				resourceData: MockForemanHostResourceData(s),
			},
			responseFile: HostsTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanHostResourceDataFromFile(
				t,
				HostsTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanHostResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanHostUpdate",
				crudFunc:     resourceForemanHostUpdate,
				resourceData: MockForemanHostResourceData(s),
			},
			responseFile: HostsTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanHostResourceDataFromFile(
				t,
				HostsTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanHostResourceDataCompare,
		},
	}

}

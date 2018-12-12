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

const SubnetsURI = api.FOREMAN_API_URL_PREFIX + "/subnets"
const SubnetsTestDataPath = "testdata/1.11/subnets"

// Given a ForemanSubnet, create a mock instance state reference
func ForemanSubnetToInstanceState(obj api.ForemanSubnet) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanSubnet
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["network"] = obj.Network
	attr["mask"] = obj.Mask
	attr["gateway"] = obj.Gateway
	attr["dns_primary"] = obj.DnsPrimary
	attr["dns_secondary"] = obj.DnsSecondary
	attr["ipam"] = obj.Ipam
	attr["from"] = obj.From
	attr["to"] = obj.To
	attr["boot_mode"] = obj.BootMode
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanSubnet resource, create a
// mock ResourceData reference.
func MockForemanSubnetResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanSubnet()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a  domain
// ResourceData reference
func MockForemanSubnetResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanSubnet
	ParseJSONFile(t, path, &obj)
	s := ForemanSubnetToInstanceState(obj)
	return MockForemanSubnetResourceData(s)
}

// Creates a random ForemanSubnet struct
func RandForemanSubnet() api.ForemanSubnet {
	obj := api.ForemanSubnet{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Network = tfrand.IPv4Str(tfrand.IPv4PrivateClassCStart, tfrand.IPv4PrivateClassCMask)
	obj.Mask = tfrand.IPv4Str(tfrand.IPv4PrivateClassCStart, tfrand.IPv4PrivateClassCMask)
	obj.Gateway = tfrand.IPv4Str(tfrand.IPv4PrivateClassCStart, tfrand.IPv4PrivateClassCMask)
	obj.DnsPrimary = tfrand.IPv4Str(tfrand.IPv4PrivateClassCStart, tfrand.IPv4PrivateClassCMask)
	obj.DnsSecondary = tfrand.IPv4Str(tfrand.IPv4PrivateClassCStart, tfrand.IPv4PrivateClassCMask)
	obj.Ipam = tfrand.String(5, tfrand.Lower)
	obj.From = tfrand.IPv4Str(tfrand.IPv4PrivateClassCStart, tfrand.IPv4PrivateClassCMask)
	obj.To = tfrand.IPv4Str(tfrand.IPv4PrivateClassCStart, tfrand.IPv4PrivateClassCMask)
	obj.BootMode = tfrand.String(5, tfrand.Lower)

	return obj
}

// Compares two ResourceData references for a ForemanSubnet resource.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanSubnetResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanSubnet()
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
func TestSubnetUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanSubnet
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanSubnet UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanSubnet UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanSubnet
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanSubnet_Value(t *testing.T) {

	expectedObj := RandForemanSubnet()
	expectedState := ForemanSubnetToInstanceState(expectedObj)
	expectedResourceData := MockForemanSubnetResourceData(expectedState)

	actualObj := api.ForemanSubnet{}
	actualState := ForemanSubnetToInstanceState(actualObj)
	actualResourceData := MockForemanSubnetResourceData(actualState)

	setResourceDataFromForemanSubnet(actualResourceData, &expectedObj)

	ForemanSubnetResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanSubnetCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanSubnet{}
	obj.Id = rand.Intn(100)
	s := ForemanSubnetToInstanceState(obj)
	architecturesURIById := SubnetsURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanSubnetRead",
				crudFunc:     resourceForemanSubnetRead,
				resourceData: MockForemanSubnetResourceData(s),
			},
			expectedURI:    architecturesURIById,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanSubnetRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanSubnet{}
	obj.Id = rand.Intn(100)
	s := ForemanSubnetToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanSubnetRead",
			crudFunc:     resourceForemanSubnetRead,
			resourceData: MockForemanSubnetResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanSubnetStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanSubnet{}
	obj.Id = rand.Intn(100)
	s := ForemanSubnetToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanSubnetRead",
			crudFunc:     resourceForemanSubnetRead,
			resourceData: MockForemanSubnetResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanSubnetEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanSubnet{}
	obj.Id = rand.Intn(100)
	s := ForemanSubnetToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanSubnetRead",
			crudFunc:     resourceForemanSubnetRead,
			resourceData: MockForemanSubnetResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanSubnetMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanSubnet()
	s := ForemanSubnetToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanSubnetRead",
				crudFunc:     resourceForemanSubnetRead,
				resourceData: MockForemanSubnetResourceData(s),
			},
			responseFile: SubnetsTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanSubnetResourceDataFromFile(
				t,
				SubnetsTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanSubnetResourceDataCompare,
		},
	}

}

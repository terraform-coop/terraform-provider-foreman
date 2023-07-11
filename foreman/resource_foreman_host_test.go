package foreman

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	tfrand "github.com/HanseMerkur/terraform-provider-utils/rand"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
	attr["managed"] = strconv.FormatBool(obj.Managed)
	if obj.DomainId != nil {
		attr["domain_id"] = strconv.Itoa(*obj.DomainId)
	}
	attr["domain_name"] = obj.DomainName
	attr["build"] = strconv.FormatBool(obj.Build)
	attr["provision_method"] = obj.ProvisionMethod

	if obj.EnvironmentId != nil {
		attr["environment_id"] = strconv.Itoa(*obj.EnvironmentId)
	}
	if obj.HostgroupId != nil {
		attr["hostgroup_id"] = strconv.Itoa(*obj.HostgroupId)
	}
	if obj.OperatingSystemId != nil {
		attr["operatingsystem_id"] = strconv.Itoa(*obj.OperatingSystemId)
	}
	if obj.MediumId != nil {
		attr["medium_id"] = strconv.Itoa(*obj.MediumId)
	}
	if obj.ImageId != nil {
		attr["image_id"] = strconv.Itoa(*obj.ImageId)
	}
	if obj.OwnerId != nil {
		attr["owner_id"] = strconv.Itoa(*obj.OwnerId)
	}
	if obj.ModelId != nil {
		attr["owner_id"] = strconv.Itoa(*obj.ModelId)
	}
	attr["owner_type"] = obj.OwnerType
	attr["interfaces_attributes.#"] = strconv.Itoa(len(obj.InterfacesAttributes))
	attr["retry_count"] = "1"
	compute_attributes, _ := json.Marshal(obj.ComputeAttributes)
	attr["compute_attributes"] = string(compute_attributes)
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

	obj.Build = rand.Float32() < 0.5

	operatingSystemId := rand.Intn(100)
	domainId := rand.Intn(100)
	hostgroupId := rand.Intn(100)
	environmentId := rand.Intn(100)
	mediumId := rand.Intn(100)
	imageId := rand.Intn(100)
	ownerId := rand.Intn(100)

	obj.OperatingSystemId = &operatingSystemId
	obj.DomainId = &domainId
	obj.HostgroupId = &hostgroupId
	obj.EnvironmentId = &environmentId
	obj.MediumId = &mediumId
	obj.ImageId = &imageId
	obj.OwnerId = &ownerId
	obj.OwnerType = "Usergroup"

	hostCompAttr := make(map[string]interface{})
	for fil := 0; fil < rand.Intn(3); fil++ {
		hostCompAttr[tfrand.String(5, tfrand.Lower)] = tfrand.String(5, tfrand.Lower)
	}
	obj.ComputeAttributes = hostCompAttr

	obj.InterfacesAttributes = make([]api.ForemanInterfacesAttribute, rand.Intn(5))
	for idx := range obj.InterfacesAttributes {
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
		// Skip compute_attribs as it gets nulled
		if key == "compute_attributes" {
			continue
		}
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
	s.Attributes["retry_count"] = "0"
	hostsURIById := HostsURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		{
			TestCase: TestCase{
				funcName:     "resourceForemanHostCreate",
				crudFunc:     resourceForemanHostCreate,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    HostsURI,
					expectedMethod: http.MethodPost,
				},
				{
					expectedURI:    HostsURI + "/0/vm_compute_attributes",
					expectedMethod: http.MethodGet,
				},
			},
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanHostRead",
				crudFunc:     resourceForemanHostRead,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    hostsURIById,
					expectedMethod: http.MethodGet,
				},
			},
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanHostUpdate",
				crudFunc:     resourceForemanHostUpdate,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    hostsURIById + "/vm_compute_attributes",
					expectedMethod: http.MethodGet,
				},
			},
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanHostDelete",
				crudFunc:     resourceForemanHostDelete,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    hostsURIById,
					expectedMethod: http.MethodDelete,
				},
			},
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanHostRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanHost{}
	obj.Id = rand.Intn(100)
	s := ForemanHostToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanHostRead",
			crudFunc:     resourceForemanHostRead,
			resourceData: MockForemanHostResourceData(s),
		},
		{
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

	_, _, client := NewForemanAPIAndClient(api.ClientCredentials{}, api.ClientConfig{})
	reqData, _ := client.WrapJSONWithTaxonomy("host", obj)

	return []TestCaseRequestData{
		{
			TestCase: TestCase{
				funcName:     "resourceForemanHostCreate",
				crudFunc:     resourceForemanHostCreate,
				resourceData: MockForemanHostResourceData(s),
			},
			expectedData: reqData,
		},
		{
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
		{
			funcName:     "resourceForemanHostCreate",
			crudFunc:     resourceForemanHostCreate,
			resourceData: MockForemanHostResourceData(s),
		},
		{
			funcName:     "resourceForemanHostRead",
			crudFunc:     resourceForemanHostRead,
			resourceData: MockForemanHostResourceData(s),
		},
		// TestCase{
		// 	funcName:     "resourceForemanHostUpdate",
		// 	crudFunc:     resourceForemanHostUpdate,
		// 	resourceData: MockForemanHostResourceData(s),
		// },
		{
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
		{
			funcName:     "resourceForemanHostCreate",
			crudFunc:     resourceForemanHostCreate,
			resourceData: MockForemanHostResourceData(s),
		},
		{
			funcName:     "resourceForemanHostRead",
			crudFunc:     resourceForemanHostRead,
			resourceData: MockForemanHostResourceData(s),
		},
		// TestCase{
		// 	funcName:     "resourceForemanHostUpdate",
		// 	crudFunc:     resourceForemanHostUpdate,
		// 	resourceData: MockForemanHostResourceData(s),
		// },
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
		{
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
		{
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
		{
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

func testResourceHostStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"method":       "build",
		"manage_build": true,
	}
}
func testResourceHostStateDataV1() map[string]interface{} {
	return map[string]interface{}{
		"method":       "build",
		"manage_build": true,
		"build":        true,
		"managed":      true,
	}
}

func TestResourceHostStateUpgradeV0(t *testing.T) {
	expected := testResourceHostStateDataV1()
	actual, err := resourceForemanHostStateUpgradeV0(context.TODO(), testResourceHostStateDataV0(), nil)

	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}

package foreman

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"

	tfrand "github.com/HanseMerkur/terraform-provider-utils/rand"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const DiscoveryRuleURI = api.FOREMAN_API_URL_PREFIX + "/v2/discovery_rules"
const DiscoveryRuleTestDataPath = "testdata/3.11/discovery_rules"

// ForemanDiscoveryRuleToInstanceState creates a mock instance state reference from a ForemanDiscoveryRule object
func ForemanDiscoveryRuleToInstanceState(obj api.ForemanDiscoveryRule) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanDiscoveryRule
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["search"] = obj.Search
	attr["hostgroup_id"] = strconv.Itoa(obj.Id)
	attr["hostname"] = obj.Hostname
	attr["max_count"] = strconv.Itoa(obj.HostsLimitMaxCount)
	attr["priority"] = strconv.Itoa(obj.Priority)
	attr["enabled"] = strconv.FormatBool(obj.Enabled)
	attr["location_ids"] = intSliceToString(obj.LocationIds)
	attr["hostgroup_ids"] = strconv.Itoa(obj.HostGroupId)
	attr["organization_ids"] = intSliceToString(obj.OrganizationIds)
	state.Attributes = attr
	return &state
}

// intSliceToString converts a slice of integers to a comma-separated string
func intSliceToString(slice []int) string {
	strSlice := make([]string, len(slice))
	for i, v := range slice {
		strSlice[i] = strconv.Itoa(v)
	}
	return strings.Join(strSlice, ",")
}

// MockForemanDiscoveryRuleResourceData creates a mock ResourceData from InstanceState.
func MockForemanDiscoveryRuleResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanDiscoveryRule()
	return r.Data(s)
}

// MockForemanDiscoveryRuleResourceDataFromFile creates a mock ResourceData from a JSON file
func MockForemanDiscoveryRuleResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanDiscoveryRule
	ParseJSONFile(t, path, &obj)
	s := ForemanDiscoveryRuleToInstanceState(obj)
	return MockForemanDiscoveryRuleResourceData(s)
}

// RandForemanDiscoveryRule generates a random ForemanDiscoveryRule object
func RandForemanDiscoveryRule() api.ForemanDiscoveryRule {
	obj := api.ForemanDiscoveryRule{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Name = tfrand.String(20, tfrand.Lower+".")

	return obj
}

// ForemanDiscoveryRuleResourceDataCompare compares two ResourceData references.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanDiscoveryRuleResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanDiscoveryRule()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

}

// TestSetResourceDataFromForemanDiscoveryRule ensures if ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanDiscoveryRule_Value(t *testing.T) {

	expectedObj := RandForemanDiscoveryRule()
	expectedState := ForemanDiscoveryRuleToInstanceState(expectedObj)
	expectedResourceData := MockForemanDiscoveryRuleResourceData(expectedState)

	actualObj := api.ForemanDiscoveryRule{}
	actualState := ForemanDiscoveryRuleToInstanceState(actualObj)
	actualResourceData := MockForemanDiscoveryRuleResourceData(actualState)

	setResourceDataFromForemanDiscoveryRule(actualResourceData, &expectedObj)

	ForemanDiscoveryRuleResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ResourceForemanDiscoveryRuleCreateTestCases Unit Test to check for correct URL and method
// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanDiscoveryRuleCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanDiscoveryRule{}
	obj.Id = rand.Intn(100)
	s := ForemanDiscoveryRuleToInstanceState(obj)
	discovery_rulesURIById := DiscoveryRuleURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		{
			TestCase: TestCase{
				funcName:     "resourceForemanDiscoveryRuleRead",
				crudFunc:     resourceForemanDiscoveryRuleRead,
				resourceData: MockForemanDiscoveryRuleResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    discovery_rulesURIById,
					expectedMethod: http.MethodGet,
				},
			},
		},
	}

}

// ResourceForemanDiscoveryRuleRequestDataEmptyTestCases Unit Test to check for empty request data
// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanDiscoveryRuleRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanDiscoveryRule{}
	obj.Id = rand.Intn(100)
	s := ForemanDiscoveryRuleToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanDiscoveryRuleRead",
			crudFunc:     resourceForemanDiscoveryRuleRead,
			resourceData: MockForemanDiscoveryRuleResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanDiscoveryRuleStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanDiscoveryRule{}
	obj.Id = rand.Intn(100)
	s := ForemanDiscoveryRuleToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanDiscoveryRuleRead",
			crudFunc:     resourceForemanDiscoveryRuleRead,
			resourceData: MockForemanDiscoveryRuleResourceData(s),
		},
	}
}

// ResourceForemanDiscoveryRuleEmptyResponseTestCases Unit Test to check for empty response
// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanDiscoveryRuleEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanDiscoveryRule{}
	obj.Id = rand.Intn(100)
	s := ForemanDiscoveryRuleToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanDiscoveryRuleRead",
			crudFunc:     resourceForemanDiscoveryRuleRead,
			resourceData: MockForemanDiscoveryRuleResourceData(s),
		},
	}
}

// ResourceForemanDiscoveryRuleMockResponseTestCases Unit Test to check agains mock response
// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanDiscoveryRuleMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanDiscoveryRule()
	s := ForemanDiscoveryRuleToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		{
			TestCase: TestCase{
				funcName:     "resourceForemanDiscoveryRuleRead",
				crudFunc:     resourceForemanDiscoveryRuleRead,
				resourceData: MockForemanDiscoveryRuleResourceData(s),
			},
			responseFile: DiscoveryRuleTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanDiscoveryRuleResourceDataFromFile(
				t,
				DiscoveryRuleTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanDiscoveryRuleResourceDataCompare,
		},
	}

}

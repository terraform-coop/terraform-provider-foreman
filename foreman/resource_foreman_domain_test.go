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

const DomainsURI = api.FOREMAN_API_URL_PREFIX + "/domains"
const DomainsTestDataPath = "testdata/1.11/domains"

// Given a ForemanDomain, create a mock instance state reference
func ForemanDomainToInstanceState(obj api.ForemanDomain) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanDomain
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["fullname"] = obj.Fullname
	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanDomain resource, create a
// mock ResourceData reference.
func MockForemanDomainResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanDomain()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a  domain
// ResourceData reference
func MockForemanDomainResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanDomain
	ParseJSONFile(t, path, &obj)
	s := ForemanDomainToInstanceState(obj)
	return MockForemanDomainResourceData(s)
}

// Creates a random ForemanDomain struct
func RandForemanDomain() api.ForemanDomain {
	obj := api.ForemanDomain{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Fullname = tfrand.String(20, tfrand.Lower+".")

	return obj
}

// Compares two ResourceData references for a ForemanDomain resource.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanDomainResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanDomain()
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
func TestDomainUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanDomain
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanDomain UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanDomain UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanDomain
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanDomain_Value(t *testing.T) {

	expectedObj := RandForemanDomain()
	expectedState := ForemanDomainToInstanceState(expectedObj)
	expectedResourceData := MockForemanDomainResourceData(expectedState)

	actualObj := api.ForemanDomain{}
	actualState := ForemanDomainToInstanceState(actualObj)
	actualResourceData := MockForemanDomainResourceData(actualState)

	setResourceDataFromForemanDomain(actualResourceData, &expectedObj)

	ForemanDomainResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanDomainCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanDomain{}
	obj.Id = rand.Intn(100)
	s := ForemanDomainToInstanceState(obj)
	domainsURIById := DomainsURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanDomainRead",
				crudFunc:     resourceForemanDomainRead,
				resourceData: MockForemanDomainResourceData(s),
			},
			expectedURI:    domainsURIById,
			expectedMethod: http.MethodGet,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanDomainRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanDomain{}
	obj.Id = rand.Intn(100)
	s := ForemanDomainToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanDomainRead",
			crudFunc:     resourceForemanDomainRead,
			resourceData: MockForemanDomainResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanDomainStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanDomain{}
	obj.Id = rand.Intn(100)
	s := ForemanDomainToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanDomainRead",
			crudFunc:     resourceForemanDomainRead,
			resourceData: MockForemanDomainResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanDomainEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanDomain{}
	obj.Id = rand.Intn(100)
	s := ForemanDomainToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanDomainRead",
			crudFunc:     resourceForemanDomainRead,
			resourceData: MockForemanDomainResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanDomainMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanDomain()
	s := ForemanDomainToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanDomainRead",
				crudFunc:     resourceForemanDomainRead,
				resourceData: MockForemanDomainResourceData(s),
			},
			responseFile: DomainsTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanDomainResourceDataFromFile(
				t,
				DomainsTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanDomainResourceDataCompare,
		},
	}

}

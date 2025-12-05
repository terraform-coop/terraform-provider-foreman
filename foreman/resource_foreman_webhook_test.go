package foreman

import (
	"math/rand"
	"net/http"
	"strconv"
	"testing"

	tfrand "github.com/HanseMerkur/terraform-provider-utils/rand"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const WebhookURI = api.FOREMAN_API_URL_PREFIX + "/webhooks"
const WebhookTestDataPath = "testdata/3.11/webhooks"

// ForemanWebhookToInstanceState creates a mock instance state reference from a ForemanWebhook object
func ForemanWebhookToInstanceState(obj api.ForemanWebhook) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanWebhook
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["target_url"] = obj.TargetURL
	attr["http_method"] = obj.HTTPMethod
	attr["http_content_type"] = obj.HTTPContentType
	attr["http_headers"] = obj.HTTPHeaders
	attr["event"] = obj.Event
	attr["enabled"] = strconv.FormatBool(obj.Enabled)
	attr["verify_ssl"] = strconv.FormatBool(obj.VerifySSL)
	attr["ssl_ca_certs"] = obj.SSLCACerts
	attr["proxy_authorization"] = strconv.FormatBool(obj.ProxyAuthorization)
	attr["user"] = obj.User
	attr["password"] = obj.Password
	attr["webhook_template_id"] = strconv.Itoa(obj.WebhookTemplateID)

	state.Attributes = attr
	return &state
}

// MockForemanWebhookResourceData creates a mock ResourceData from InstanceState.
func MockForemanWebhookResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanWebhook()
	return r.Data(s)
}

// MockForemanWebhookResourceDataFromFile creates a mock ResourceData from a JSON file
func MockForemanWebhookResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanWebhook
	ParseJSONFile(t, path, &obj)
	s := ForemanWebhookToInstanceState(obj)
	return MockForemanWebhookResourceData(s)
}

// RandForemanWebhook generates a random ForemanWebhook object
func RandForemanWebhook() api.ForemanWebhook {
	obj := api.ForemanWebhook{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Name = tfrand.String(20, tfrand.Lower+".")

	return obj
}

// ForemanWebhookResourceDataCompare compares two ResourceData references.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanWebhookResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanWebhook()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

}

// TestSetResourceDataFromForemanWebhook ensures if ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanWebhook_Value(t *testing.T) {

	expectedObj := RandForemanWebhook()
	expectedState := ForemanWebhookToInstanceState(expectedObj)
	expectedResourceData := MockForemanWebhookResourceData(expectedState)

	actualObj := api.ForemanWebhook{}
	actualState := ForemanWebhookToInstanceState(actualObj)
	actualResourceData := MockForemanWebhookResourceData(actualState)

	setResourceDataFromForemanWebhook(actualResourceData, &expectedObj)

	ForemanWebhookResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ResourceForemanWebhookCreateTestCases Unit Test to check for correct URL and method
// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanWebhookCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanWebhook{}
	obj.Id = rand.Intn(100)
	s := ForemanWebhookToInstanceState(obj)
	webhooksURIById := WebhookURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		{
			TestCase: TestCase{
				funcName:     "resourceForemanWebhookRead",
				crudFunc:     resourceForemanWebhookRead,
				resourceData: MockForemanWebhookResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    webhooksURIById,
					expectedMethod: http.MethodGet,
				},
			},
		},
	}

}

// ResourceForemanWebhookRequestDataEmptyTestCases Unit Test to check for empty request data
// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanWebhookRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanWebhook{}
	obj.Id = rand.Intn(100)
	s := ForemanWebhookToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanWebhookRead",
			crudFunc:     resourceForemanWebhookRead,
			resourceData: MockForemanWebhookResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanWebhookStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanWebhook{}
	obj.Id = rand.Intn(100)
	s := ForemanWebhookToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanWebhookRead",
			crudFunc:     resourceForemanWebhookRead,
			resourceData: MockForemanWebhookResourceData(s),
		},
	}
}

// ResourceForemanWebhookEmptyResponseTestCases Unit Test to check for empty response
// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanWebhookEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanWebhook{}
	obj.Id = rand.Intn(100)
	s := ForemanWebhookToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanWebhookRead",
			crudFunc:     resourceForemanWebhookRead,
			resourceData: MockForemanWebhookResourceData(s),
		},
	}
}

// ResourceForemanWebhookMockResponseTestCases Unit Test to check against mock response
// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanWebhookMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanWebhook()
	s := ForemanWebhookToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		{
			TestCase: TestCase{
				funcName:     "resourceForemanWebhookRead",
				crudFunc:     resourceForemanWebhookRead,
				resourceData: MockForemanWebhookResourceData(s),
			},
			responseFile: WebhookTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanWebhookResourceDataFromFile(
				t,
				WebhookTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanWebhookResourceDataCompare,
		},
	}
}

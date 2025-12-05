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

const WebhookTemplateURI = api.FOREMAN_API_URL_PREFIX + "/webhook_templates"
const WebhookTemplateTestDataPath = "testdata/3.11/webhook_templates"

// ForemanWebhookTemplateToInstanceState creates a mock instance state reference from a ForemanWebhookTemplate object
func ForemanWebhookTemplateToInstanceState(obj api.ForemanWebhookTemplate) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanWebhookTemplate
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["template"] = obj.Template
	attr["snippet"] = strconv.FormatBool(obj.Snippet)
	attr["audit_comment"] = obj.AuditComment
	attr["locked"] = strconv.FormatBool(obj.Locked)
	attr["default"] = strconv.FormatBool(obj.Default)
	attr["description"] = obj.Description
	attr["location_ids"] = intSliceToString(obj.LocationIds)
	attr["organization_ids"] = intSliceToString(obj.OrganizationIds)

	state.Attributes = attr
	return &state
}

// MockForemanWebhookTemplateResourceData creates a mock ResourceData from InstanceState.
func MockForemanWebhookTemplateResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanWebhookTemplate()
	return r.Data(s)
}

// MockForemanWebhookTemplateResourceDataFromFile creates a mock ResourceData from a JSON file
func MockForemanWebhookTemplateResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanWebhookTemplate
	ParseJSONFile(t, path, &obj)
	s := ForemanWebhookTemplateToInstanceState(obj)
	return MockForemanWebhookTemplateResourceData(s)
}

// RandForemanWebhookTemplate generates a random ForemanWebhookTemplate object
func RandForemanWebhookTemplate() api.ForemanWebhookTemplate {
	obj := api.ForemanWebhookTemplate{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Name = tfrand.String(20, tfrand.Lower+".")

	return obj
}

// ForemanWebhookTemplateResourceDataCompare compares two ResourceData references.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanWebhookTemplateResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanWebhookTemplate()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

}

// TestSetResourceDataFromForemanWebhookTemplate ensures if ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanWebhookTemplate_Value(t *testing.T) {

	expectedObj := RandForemanWebhookTemplate()
	expectedState := ForemanWebhookTemplateToInstanceState(expectedObj)
	expectedResourceData := MockForemanWebhookTemplateResourceData(expectedState)

	actualObj := api.ForemanWebhookTemplate{}
	actualState := ForemanWebhookTemplateToInstanceState(actualObj)
	actualResourceData := MockForemanWebhookTemplateResourceData(actualState)

	setResourceDataFromForemanWebhookTemplate(actualResourceData, &expectedObj)

	ForemanWebhookTemplateResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ResourceForemanWebhookTemplateCreateTestCases Unit Test to check for correct URL and method
// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanWebhookTemplateCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanWebhookTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanWebhookTemplateToInstanceState(obj)
	webhookTemplatesURIById := WebhookTemplateURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		{
			TestCase: TestCase{
				funcName:     "resourceForemanWebhookTemplateRead",
				crudFunc:     resourceForemanWebhookTemplateRead,
				resourceData: MockForemanWebhookTemplateResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    webhookTemplatesURIById,
					expectedMethod: http.MethodGet,
				},
			},
		},
	}

}

// ResourceForemanWebhookTemplateRequestDataEmptyTestCases Unit Test to check for empty request data
// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanWebhookTemplateRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanWebhookTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanWebhookTemplateToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanWebhookTemplateRead",
			crudFunc:     resourceForemanWebhookTemplateRead,
			resourceData: MockForemanWebhookTemplateResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanWebhookTemplateStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanWebhookTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanWebhookTemplateToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanWebhookTemplateRead",
			crudFunc:     resourceForemanWebhookTemplateRead,
			resourceData: MockForemanWebhookTemplateResourceData(s),
		},
	}
}

// ResourceForemanWebhookTemplateEmptyResponseTestCases Unit Test to check for empty response
// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanWebhookTemplateEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanWebhookTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanWebhookTemplateToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanWebhookTemplateRead",
			crudFunc:     resourceForemanWebhookTemplateRead,
			resourceData: MockForemanWebhookTemplateResourceData(s),
		},
	}
}

// ResourceForemanWebhookTemplateMockResponseTestCases Unit Test to check against mock response
// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanWebhookTemplateMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanWebhookTemplate()
	s := ForemanWebhookTemplateToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		{
			TestCase: TestCase{
				funcName:     "resourceForemanWebhookTemplateRead",
				crudFunc:     resourceForemanWebhookTemplateRead,
				resourceData: MockForemanWebhookTemplateResourceData(s),
			},
			responseFile: WebhookTemplateTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanWebhookTemplateResourceDataFromFile(
				t,
				WebhookTemplateTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanWebhookTemplateResourceDataCompare,
		},
	}
}

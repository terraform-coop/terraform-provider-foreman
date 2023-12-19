package foreman

import (
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"testing"
)

const JobTemplatesURI = api.FOREMAN_API_URL_PREFIX + "/job_templates"
const JobTemplatesTestDataPath = "testdata/3.6/job_template"

func RandForemanJobTemplate() api.ForemanJobTemplate {
	fo := RandForemanObject()
	return api.ForemanJobTemplate{
		ForemanObject:     fo,
		Description:       "random description",
		DescriptionFormat: "bla",
		Template:          "<% my_var %>",
		Locked:            true,
		JobCategory:       "testing",
		ProviderType:      "providerType123",
		Snippet:           false,
	}
}

func ForemanJobTemplateToInstanceState(obj api.ForemanJobTemplate) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)

	state.Attributes = map[string]string{
		"name":               obj.Name,
		"description":        obj.Description,
		"description_format": obj.DescriptionFormat,
		"template":           obj.Template,
		"job_category":       obj.JobCategory,
		"provider_type":      obj.ProviderType,
		"locked":             strconv.FormatBool(obj.Locked),
		"snippet":            strconv.FormatBool(obj.Snippet),
	}
	return &state
}

func MockForemanJobTemplateResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanJobTemplate()
	return r.Data(s)
}

func MockForemanJobTemplateResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanJobTemplate
	ParseJSONFile(t, path, &obj)
	s := ForemanJobTemplateToInstanceState(obj)
	return MockForemanJobTemplateResourceData(s)
}

func ForemanJobTemplateResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {
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
	r := resourceForemanJobTemplate()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)
}

// JSON Marshaling

// Ensures the JSON unmarshal correctly sets the base attributes from ForemanObject
func TestJobTemplateUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanJobTemplate
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanJobTemplate UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanJobTemplate UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanJobTemplate_Value(t *testing.T) {

	expectedObj := RandForemanJobTemplate()
	expectedState := ForemanJobTemplateToInstanceState(expectedObj)
	expectedResourceData := MockForemanJobTemplateResourceData(expectedState)

	actualObj := api.ForemanJobTemplate{}
	actualState := ForemanJobTemplateToInstanceState(actualObj)
	actualResourceData := MockForemanJobTemplateResourceData(actualState)

	setResourceDataFromForemanJobTemplate(actualResourceData, &expectedObj)

	ForemanJobTemplateResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanJobTemplateCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanJobTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanJobTemplateToInstanceState(obj)
	jobTemplatesURIById := JobTemplatesURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		{
			TestCase: TestCase{
				funcName:     "resourceForemanJobTemplateCreate",
				crudFunc:     resourceForemanJobTemplateCreate,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    JobTemplatesURI,
					expectedMethod: http.MethodPost,
				},
			},
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanJobTemplateRead",
				crudFunc:     resourceForemanJobTemplateRead,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    jobTemplatesURIById,
					expectedMethod: http.MethodGet,
				},
			},
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanJobTemplateUpdate",
				crudFunc:     resourceForemanJobTemplateUpdate,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    jobTemplatesURIById,
					expectedMethod: http.MethodPut,
				},
			},
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanJobTemplateDelete",
				crudFunc:     resourceForemanJobTemplateDelete,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			expectedURIs: []ExpectedUri{
				{
					expectedURI:    jobTemplatesURIById,
					expectedMethod: http.MethodDelete,
				},
			},
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanJobTemplateRequestDataEmptyTestCases(t *testing.T) []TestCase {
	obj := api.ForemanJobTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanJobTemplateToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanJobTemplateRead",
			crudFunc:     resourceForemanJobTemplateRead,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
		{
			funcName:     "resourceForemanJobTemplateDelete",
			crudFunc:     resourceForemanJobTemplateDelete,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanJobTemplateRequestDataTestCases(t *testing.T) []TestCaseRequestData {
	obj := api.ForemanJobTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanJobTemplateToInstanceState(obj)

	rd := MockForemanJobTemplateResourceData(s)
	obj = *buildForemanJobTemplate(rd)
	cred := api.ClientCredentials{}
	conf := api.ClientConfig{}

	_, _, client := NewForemanAPIAndClient(cred, conf)
	reqData, _ := client.WrapJSONWithTaxonomy("job_template", obj)

	return []TestCaseRequestData{
		{
			TestCase: TestCase{
				funcName:     "resourceForemanJobTemplateCreate",
				crudFunc:     resourceForemanJobTemplateCreate,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			expectedData: reqData,
		},
		{
			TestCase: TestCase{
				funcName:     "resourceForemanJobTemplateUpdate",
				crudFunc:     resourceForemanJobTemplateUpdate,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			expectedData: reqData,
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanJobTemplateStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanJobTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanJobTemplateToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanJobTemplateCreate",
			crudFunc:     resourceForemanJobTemplateCreate,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
		{
			funcName:     "resourceForemanJobTemplateRead",
			crudFunc:     resourceForemanJobTemplateRead,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
		{
			funcName:     "resourceForemanJobTemplateUpdate",
			crudFunc:     resourceForemanJobTemplateUpdate,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
		{
			funcName:     "resourceForemanJobTemplateDelete",
			crudFunc:     resourceForemanJobTemplateDelete,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanJobTemplateEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanJobTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanJobTemplateToInstanceState(obj)

	return []TestCase{
		{
			funcName:     "resourceForemanJobTemplateCreate",
			crudFunc:     resourceForemanJobTemplateCreate,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
		{
			funcName:     "resourceForemanJobTemplateRead",
			crudFunc:     resourceForemanJobTemplateRead,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
		{
			funcName:     "resourceForemanJobTemplateUpdate",
			crudFunc:     resourceForemanJobTemplateUpdate,
			resourceData: MockForemanJobTemplateResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanJobTemplateMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanJobTemplate()
	s := ForemanJobTemplateToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		{
			TestCase: TestCase{
				funcName:     "resourceForemanJobTemplateCreate",
				crudFunc:     resourceForemanJobTemplateCreate,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			responseFile: JobTemplatesTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanJobTemplateResourceDataFromFile(
				t,
				JobTemplatesTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanJobTemplateResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		{
			TestCase: TestCase{
				funcName:     "resourceForemanJobTemplateRead",
				crudFunc:     resourceForemanJobTemplateRead,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			responseFile: JobTemplatesTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanJobTemplateResourceDataFromFile(
				t,
				JobTemplatesTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanJobTemplateResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		{
			TestCase: TestCase{
				funcName:     "resourceForemanJobTemplateUpdate",
				crudFunc:     resourceForemanJobTemplateUpdate,
				resourceData: MockForemanJobTemplateResourceData(s),
			},
			responseFile: JobTemplatesTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanJobTemplateResourceDataFromFile(
				t,
				JobTemplatesTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanJobTemplateResourceDataCompare,
		},
	}

}

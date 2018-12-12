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

const ProvisioningTemplatesURI = api.FOREMAN_API_URL_PREFIX + "/provisioning_templates"
const ProvisioningTemplatesTestDataPath = "testdata/1.11/provisioning_templates"

// Given a ForemanProvisioningTemplate, create a mock instance state reference
func ForemanProvisioningTemplateToInstanceState(obj api.ForemanProvisioningTemplate) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = strconv.Itoa(obj.Id)
	// Build the attribute map from ForemanProvisioningTemplate
	attr := map[string]string{}
	attr["name"] = obj.Name
	attr["template"] = obj.Template
	attr["snippet"] = fmt.Sprintf("%t", obj.Snippet)
	attr["audit_comment"] = obj.AuditComment
	attr["locked"] = fmt.Sprintf("%t", obj.Locked)
	attr["template_kind_id"] = strconv.Itoa(obj.TemplateKindId)

	attr["operatingsystem_ids.#"] = strconv.Itoa(len(obj.OperatingSystemIds))
	for idx, val := range obj.OperatingSystemIds {
		key := fmt.Sprintf("operatingsystem_ids.%d", idx)
		attr[key] = strconv.Itoa(val)
	}

	attr["template_combinations_attributes.#"] = strconv.Itoa(len(obj.TemplateCombinationsAttributes))
	for idx, val := range obj.TemplateCombinationsAttributes {
		key := fmt.Sprintf("template_combinations_attributes.%d.id", idx)
		attr[key] = strconv.Itoa(val.Id)
		key = fmt.Sprintf("template_combinations_attributes.%d.hostgroup_id", idx)
		attr[key] = strconv.Itoa(val.HostgroupId)
		key = fmt.Sprintf("template_combinations_attributes.%d.environment_id", idx)
		attr[key] = strconv.Itoa(val.EnvironmentId)
	}

	state.Attributes = attr
	return &state
}

// Given a mock instance state for a ForemanProvisioningTemplate resource, create a
// mock ResourceData reference.
func MockForemanProvisioningTemplateResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := resourceForemanProvisioningTemplate()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a model
// ResourceData reference
func MockForemanProvisioningTemplateResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanProvisioningTemplate
	ParseJSONFile(t, path, &obj)
	s := ForemanProvisioningTemplateToInstanceState(obj)
	return MockForemanProvisioningTemplateResourceData(s)
}

// Creates a random ForemanProvisioningTemplate struct
func RandForemanProvisioningTemplate() api.ForemanProvisioningTemplate {
	obj := api.ForemanProvisioningTemplate{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	obj.Template = tfrand.String(100, tfrand.Lower+" \r\n.")
	obj.Snippet = rand.Intn(2) > 0
	obj.AuditComment = tfrand.String(100, tfrand.Lower+". ")
	obj.Locked = rand.Intn(2) > 0
	obj.TemplateKindId = rand.Intn(100)

	obj.OperatingSystemIds = tfrand.IntArrayUnique(rand.Intn(5))

	obj.TemplateCombinationsAttributes = make([]api.ForemanTemplateCombinationAttribute, rand.Intn(5))
	for idx, _ := range obj.TemplateCombinationsAttributes {
		obj.TemplateCombinationsAttributes[idx] = api.ForemanTemplateCombinationAttribute{
			Id:            rand.Intn(100),
			HostgroupId:   rand.Intn(100),
			EnvironmentId: rand.Intn(100),
		}
	}

	return obj
}

// Compares two ResourceData references for a ForemanProvisioningTemplate resoure.
// If the two references differ in their attributes, the test will raise
// a fatal.
func ForemanProvisioningTemplateResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := resourceForemanProvisioningTemplate()
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

	attr1, ok1 = r1.Get("template_combinations_attributes").(*schema.Set)
	attr2, ok2 = r2.Get("template_combinations_attributes").(*schema.Set)
	if ok1 && ok2 {
		attr1Set := attr1.(*schema.Set)
		attr2Set := attr1.(*schema.Set)
		if !attr1Set.Equal(attr2Set) {
			t.Fatalf(
				"ResourceData reference differ in template_combinations_attributes. "+
					"[%v], [%v]",
				attr1Set.List(),
				attr2Set.List(),
			)
		}
	} else if (ok1 && !ok2) || (!ok1 && ok2) {
		t.Fatalf(
			"ResourceData reference differ in template_combinations_attributes. "+
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
func TestProvisioningTemplateUnmarshalJSON_ForemanObject(t *testing.T) {

	randObj := RandForemanObject()
	randObjBytes, _ := json.Marshal(randObj)

	var obj api.ForemanProvisioningTemplate
	jsonDecErr := json.Unmarshal(randObjBytes, &obj)
	if jsonDecErr != nil {
		t.Errorf(
			"ForemanProvisioningTemplate UnmarshalJSON could not decode base ForemanObject. "+
				"Expected [nil] got [error]. Error value: [%s]",
			jsonDecErr,
		)
	}

	if !reflect.DeepEqual(obj.ForemanObject, randObj) {
		t.Errorf(
			"ForemanProvisioningTemplate UnmarshalJSON did not properly decode base "+
				"ForemanObject properties. Expected [%+v], got [%+v]",
			randObj,
			obj.ForemanObject,
		)
	}

}

// -----------------------------------------------------------------------------
// buildForemanProvisioningTemplate
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being read to
// create a ForemanProvisioningTemplate
func TestBuildForemanProvisioningTemplate(t *testing.T) {

	expectedObj := RandForemanProvisioningTemplate()
	expectedState := ForemanProvisioningTemplateToInstanceState(expectedObj)
	expectedResourceData := MockForemanProvisioningTemplateResourceData(expectedState)

	actualObj := buildForemanProvisioningTemplate(expectedResourceData)

	actualState := ForemanProvisioningTemplateToInstanceState(*actualObj)
	actualResourceData := MockForemanProvisioningTemplateResourceData(actualState)

	ForemanProvisioningTemplateResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// -----------------------------------------------------------------------------
// setResourceDataFromForemanProvisioningTemplate
// -----------------------------------------------------------------------------

// Ensures the ResourceData's attributes are correctly being set
func TestSetResourceDataFromForemanProvisioningTemplate_Value(t *testing.T) {

	expectedObj := RandForemanProvisioningTemplate()
	expectedState := ForemanProvisioningTemplateToInstanceState(expectedObj)
	expectedResourceData := MockForemanProvisioningTemplateResourceData(expectedState)

	actualObj := api.ForemanProvisioningTemplate{}
	actualState := ForemanProvisioningTemplateToInstanceState(actualObj)
	actualResourceData := MockForemanProvisioningTemplateResourceData(actualState)

	setResourceDataFromForemanProvisioningTemplate(actualResourceData, &expectedObj)

	ForemanProvisioningTemplateResourceDataCompare(t, actualResourceData, expectedResourceData)

}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

// SEE: foreman_api_test.go#TestCRUDFunction_CorrectURLAndMethod()
func ResourceForemanProvisioningTemplateCorrectURLAndMethodTestCases(t *testing.T) []TestCaseCorrectURLAndMethod {

	obj := api.ForemanProvisioningTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanProvisioningTemplateToInstanceState(obj)
	provisioningTemplatesURIById := ProvisioningTemplatesURI + "/" + strconv.Itoa(obj.Id)

	return []TestCaseCorrectURLAndMethod{
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanProvisioningTemplateCreate",
				crudFunc:     resourceForemanProvisioningTemplateCreate,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			expectedURI:    ProvisioningTemplatesURI,
			expectedMethod: http.MethodPost,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanProvisioningTemplateRead",
				crudFunc:     resourceForemanProvisioningTemplateRead,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			expectedURI:    provisioningTemplatesURIById,
			expectedMethod: http.MethodGet,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanProvisioningTemplateUpdate",
				crudFunc:     resourceForemanProvisioningTemplateUpdate,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			expectedURI:    provisioningTemplatesURIById,
			expectedMethod: http.MethodPut,
		},
		TestCaseCorrectURLAndMethod{
			TestCase: TestCase{
				funcName:     "resourceForemanProvisioningTemplateDelete",
				crudFunc:     resourceForemanProvisioningTemplateDelete,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			expectedURI:    provisioningTemplatesURIById,
			expectedMethod: http.MethodDelete,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestDataEmpty()
func ResourceForemanProvisioningTemplateRequestDataEmptyTestCases(t *testing.T) []TestCase {

	obj := api.ForemanProvisioningTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanProvisioningTemplateToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanProvisioningTemplateRead",
			crudFunc:     resourceForemanProvisioningTemplateRead,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanProvisioningTemplateDelete",
			crudFunc:     resourceForemanProvisioningTemplateDelete,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_RequestData()
func ResourceForemanProvisioningTemplateRequestDataTestCases(t *testing.T) []TestCaseRequestData {

	obj := api.ForemanProvisioningTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanProvisioningTemplateToInstanceState(obj)

	rd := MockForemanProvisioningTemplateResourceData(s)
	obj = *buildForemanProvisioningTemplate(rd)
	reqData, _ := json.Marshal(obj)

	return []TestCaseRequestData{
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanProvisioningTemplateCreate",
				crudFunc:     resourceForemanProvisioningTemplateCreate,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			expectedData: reqData,
		},
		TestCaseRequestData{
			TestCase: TestCase{
				funcName:     "resourceForemanProvisioningTemplateUpdate",
				crudFunc:     resourceForemanProvisioningTemplateUpdate,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			expectedData: reqData,
		},
	}

}

// SEE: foreman_api_test.go#TestCRUDFunction_StatusCodeError()
func ResourceForemanProvisioningTemplateStatusCodeTestCases(t *testing.T) []TestCase {

	obj := api.ForemanProvisioningTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanProvisioningTemplateToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanProvisioningTemplateCreate",
			crudFunc:     resourceForemanProvisioningTemplateCreate,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanProvisioningTemplateRead",
			crudFunc:     resourceForemanProvisioningTemplateRead,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanProvisioningTemplateUpdate",
			crudFunc:     resourceForemanProvisioningTemplateUpdate,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanProvisioningTemplateDelete",
			crudFunc:     resourceForemanProvisioningTemplateDelete,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_EmptyResponseError()
func ResourceForemanProvisioningTemplateEmptyResponseTestCases(t *testing.T) []TestCase {
	obj := api.ForemanProvisioningTemplate{}
	obj.Id = rand.Intn(100)
	s := ForemanProvisioningTemplateToInstanceState(obj)

	return []TestCase{
		TestCase{
			funcName:     "resourceForemanProvisioningTemplateCreate",
			crudFunc:     resourceForemanProvisioningTemplateCreate,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanProvisioningTemplateRead",
			crudFunc:     resourceForemanProvisioningTemplateRead,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
		TestCase{
			funcName:     "resourceForemanProvisioningTemplateUpdate",
			crudFunc:     resourceForemanProvisioningTemplateUpdate,
			resourceData: MockForemanProvisioningTemplateResourceData(s),
		},
	}
}

// SEE: foreman_api_test.go#TestCRUDFunction_MockResponse()
func ResourceForemanProvisioningTemplateMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanProvisioningTemplate()
	s := ForemanProvisioningTemplateToInstanceState(obj)

	return []TestCaseMockResponse{
		// If the server responds with a proper create response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanProvisioningTemplateCreate",
				crudFunc:     resourceForemanProvisioningTemplateCreate,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			responseFile: ProvisioningTemplatesTestDataPath + "/create_response.json",
			returnError:  false,
			expectedResourceData: MockForemanProvisioningTemplateResourceDataFromFile(
				t,
				ProvisioningTemplatesTestDataPath+"/create_response.json",
			),
			compareFunc: ForemanProvisioningTemplateResourceDataCompare,
		},
		// If the server responds with a proper read response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanProvisioningTemplateRead",
				crudFunc:     resourceForemanProvisioningTemplateRead,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			responseFile: ProvisioningTemplatesTestDataPath + "/read_response.json",
			returnError:  false,
			expectedResourceData: MockForemanProvisioningTemplateResourceDataFromFile(
				t,
				ProvisioningTemplatesTestDataPath+"/read_response.json",
			),
			compareFunc: ForemanProvisioningTemplateResourceDataCompare,
		},
		// If the server responds with a proper update response, the operation
		// should succeed and the ResourceData's attributes should be updated
		// to server's response
		TestCaseMockResponse{
			TestCase: TestCase{
				funcName:     "resourceForemanProvisioningTemplateUpdate",
				crudFunc:     resourceForemanProvisioningTemplateUpdate,
				resourceData: MockForemanProvisioningTemplateResourceData(s),
			},
			responseFile: ProvisioningTemplatesTestDataPath + "/update_response.json",
			returnError:  false,
			expectedResourceData: MockForemanProvisioningTemplateResourceDataFromFile(
				t,
				ProvisioningTemplatesTestDataPath+"/update_response.json",
			),
			compareFunc: ForemanProvisioningTemplateResourceDataCompare,
		},
	}

}

package foreman

import (
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
	"reflect"
	"strconv"
	"testing"
)

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

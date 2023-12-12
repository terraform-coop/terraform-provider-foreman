package foreman

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
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

		// Throw errors due to unknown reasons
		//"template":           obj.Template,
		//"locked":             strconv.FormatBool(obj.Locked),

		"job_category":  obj.JobCategory,
		"provider_type": obj.ProviderType,
		"snippet":       strconv.FormatBool(obj.Snippet),
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

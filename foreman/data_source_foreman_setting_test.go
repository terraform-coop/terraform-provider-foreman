package foreman

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-coop/terraform-provider-foreman/foreman/api"
)

func RandForemanSetting() api.ForemanSetting {
	obj := api.ForemanSetting{}

	fo := RandForemanObject()
	obj.ForemanObject = fo

	// Settings API uses string as Id, overridden
	obj.Id = fmt.Sprintf("randomSetting%d", fo.Id)

	return obj
}

func ForemanSettingToInstanceState(obj api.ForemanSetting) *terraform.InstanceState {
	state := terraform.InstanceState{}
	state.ID = obj.Id

	attr := map[string]string{}
	attr["id"] = obj.Id
	attr["name"] = obj.Name

	if obj.Default != nil {
		attr["default"] = obj.Default.(string)
	}

	attr["description"] = obj.Description
	attr["settings_type"] = obj.SettingsType
	attr["created_at"] = obj.CreatedAt
	attr["updated_at"] = obj.UpdatedAt
	attr["full_name"] = obj.Fullname

	if obj.Value != nil {
		attr["value"] = obj.Value.(string)
	}

	attr["category"] = obj.Category
	attr["category_name"] = obj.CategoryName

	state.Attributes = attr
	return &state
}

// ----------------------------------------------------------------------------
// Test Cases for the Unit Test Framework
// ----------------------------------------------------------------------------

func DataSourceForemanSettingMockResponseTestCases(t *testing.T) []TestCaseMockResponse {

	obj := RandForemanSetting()
	s := ForemanSettingToInstanceState(obj)

	return []TestCaseMockResponse{
		{
			TestCase: TestCase{
				funcName:     "dataSourceForemanSettingRead",
				crudFunc:     dataSourceForemanSettingRead,
				resourceData: MockForemanSettingResourceData(s),
			},
			responseFile: TestDataPath + "/settings/query_response_single.json",
			returnError:  false,
			expectedResourceData: MockForemanSettingResourceDataFromFile(
				t,
				TestDataPath+"/settings/query_response_single_state.json",
			),
			compareFunc: ForemanSettingResourceDataCompare,
		},
	}
}

func MockForemanSettingResourceData(s *terraform.InstanceState) *schema.ResourceData {
	r := dataSourceForemanSetting()
	return r.Data(s)
}

// Reads the JSON for the file at the path and creates a  domain
// ResourceData reference
func MockForemanSettingResourceDataFromFile(t *testing.T, path string) *schema.ResourceData {
	var obj api.ForemanSetting
	ParseJSONFile(t, path, &obj)
	s := ForemanSettingToInstanceState(obj)
	return MockForemanSettingResourceData(s)
}

func ForemanSettingResourceDataCompare(t *testing.T, r1 *schema.ResourceData, r2 *schema.ResourceData) {

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
	r := dataSourceForemanSetting()
	for key, value := range r.Schema {
		m[key] = value.Type
	}

	// compare the rest of the attributes
	CompareResourceDataAttributes(t, m, r1, r2)

}

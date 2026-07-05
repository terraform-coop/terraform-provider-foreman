package goforeman

// Validates every generated entity struct's json.Unmarshal against real
// Foreman API response fixtures captured by the old provider's test suite
// (generated/realworld_testdata/, copied from the old provider's
// foreman/testdata/ - real servers, versions 1.11 through 3.11), not just
// this generator's own synthetic round-trip tests.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// realworldFixture maps one testdata resource directory to the entity type
// its "*_response.json" files should unmarshal into.
type realworldFixture struct {
	dir       string
	newEntity func() interface{}
}

var realworldFixtures = []realworldFixture{
	{"architectures", func() interface{} { return &ForemanArchitecture{} }},
	{"computeresources", func() interface{} { return &ForemanComputeResource{} }},
	{"domains", func() interface{} { return &ForemanDomain{} }},
	{"environments", func() interface{} { return &ForemanEnvironment{} }},
	{"hostgroups", func() interface{} { return &ForemanHostgroup{} }},
	{"hosts", func() interface{} { return &ForemanHost{} }},
	{"media", func() interface{} { return &ForemanMedium{} }},
	{"models", func() interface{} { return &ForemanModel{} }},
	{"operatingsystems", func() interface{} { return &ForemanOperatingSystem{} }},
	{"provisioning_templates", func() interface{} { return &ForemanProvisioningTemplate{} }},
	{"ptables", func() interface{} { return &ForemanPartitionTable{} }},
	{"settings", func() interface{} { return &ForemanSetting{} }},
	{"smart_proxies", func() interface{} { return &ForemanSmartProxy{} }},
	{"subnets", func() interface{} { return &ForemanSubnet{} }},
	{"template_kinds", func() interface{} { return &ForemanTemplateKind{} }},
	{"image", func() interface{} { return &ForemanImage{} }},
	{"override_values", func() interface{} { return &ForemanOverrideValue{} }},
	{"puppet_classes", func() interface{} { return &ForemanPuppetClass{} }},
	{"smart_class_parameters", func() interface{} { return &ForemanSmartClassParameter{} }},
	{"discovery_rules", func() interface{} { return &ForemanDiscoveryRule{} }},
	{"webhook_templates", func() interface{} { return &ForemanWebhookTemplate{} }},
	{"webhooks", func() interface{} { return &ForemanWebhook{} }},
	{"job_template", func() interface{} { return &ForemanJobTemplate{} }},
}

// singleObjectFiles unmarshal directly into the entity type.
var singleObjectFiles = []string{"read_response.json", "create_response.json", "update_response.json"}

// queryResponseFiles are wrapped in the standard {total, subtotal, ...,
// results: [...]} envelope; each element of "results" unmarshals into the
// entity type.
var queryResponseFiles = []string{
	"query_response_single.json", "query_response_multi.json",
	"query_response_colon.json",
}

func TestRealWorldFixtures_SingleObject(t *testing.T) {
	for _, fx := range realworldFixtures {
		matches, err := filepath.Glob(filepath.Join("realworld_testdata", "*", fx.dir))
		require.NoError(t, err)
		for _, dir := range matches {
			for _, name := range singleObjectFiles {
				path := filepath.Join(dir, name)
				data, err := os.ReadFile(path)
				if os.IsNotExist(err) {
					continue
				}
				require.NoError(t, err, path)

				t.Run(filepath.Join(dir, name), func(t *testing.T) {
					entity := fx.newEntity()
					err := json.Unmarshal(data, entity)
					require.NoError(t, err, "unmarshal %s into %T", path, entity)
				})
			}
		}
	}
}

func TestRealWorldFixtures_QueryResponse(t *testing.T) {
	for _, fx := range realworldFixtures {
		if fx.dir == "puppet_classes" {
			// puppetclasses' real index/query response groups results by
			// Puppet environment name ({"testing": [...]}), not the
			// standard flat {"results": [...]} array - see
			// TestRealWorldFixtures_PuppetClassGroupedQuery below.
			continue
		}
		matches, err := filepath.Glob(filepath.Join("realworld_testdata", "*", fx.dir))
		require.NoError(t, err)
		for _, dir := range matches {
			for _, name := range queryResponseFiles {
				path := filepath.Join(dir, name)
				data, err := os.ReadFile(path)
				if os.IsNotExist(err) {
					continue
				}
				require.NoError(t, err, path)

				t.Run(filepath.Join(dir, name), func(t *testing.T) {
					var resp QueryResponse
					require.NoError(t, json.Unmarshal(data, &resp), "unmarshal %s into QueryResponse", path)
					for i, raw := range resp.Results {
						entity := fx.newEntity()
						err := json.Unmarshal(raw, entity)
						require.NoError(t, err, "unmarshal %s results[%d] into %T", path, i, entity)
					}
				})
			}
		}
	}
}

// TestRealWorldFixtures_PuppetClassGroupedQuery validates QueryForemanPuppetClass's
// hand-written unmarshal of the real, environment-grouped response shape
// (see generated/puppet_class.go), using the same fixture QueryForemanPuppetClass
// itself would receive from c.Get.
func TestRealWorldFixtures_PuppetClassGroupedQuery(t *testing.T) {
	matches, err := filepath.Glob(filepath.Join("realworld_testdata", "*", "puppet_classes", "query_response_single.json"))
	require.NoError(t, err)
	require.NotEmpty(t, matches)

	for _, path := range matches {
		data, err := os.ReadFile(path)
		require.NoError(t, err)

		t.Run(path, func(t *testing.T) {
			var response struct {
				Results map[string][]ForemanPuppetClass `json:"results"`
			}
			require.NoError(t, json.Unmarshal(data, &response), "unmarshal %s", path)
			require.NotEmpty(t, response.Results, "expected at least one environment group in %s", path)
		})
	}
}

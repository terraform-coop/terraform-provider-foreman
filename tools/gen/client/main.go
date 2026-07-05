// Command gen generates the Foreman API client library from apipie-rails JSON.
//
// Usage:
//
//	go run tools/gen/client/main.go --input apidoc/v2.json --output generated/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"gopkg.in/yaml.v3"
)

var verbose bool

func main() {
	inputPath := flag.String("input", "apidoc/v2.json", "Path to apidoc/v2.json")
	outputDir := flag.String("output", "generated/", "Output directory for generated client files")
	providerDir := flag.String("provider", "internal/provider/", "Output directory for generated provider files")
	overridesPath := flag.String("overrides", "tools/gen/overrides.yaml", "Path to type overrides file")
	verboseFlag := flag.Bool("verbose", false, "Log skipped resources and generation details")
	flag.Parse()
	verbose = *verboseFlag

	doc, err := parseApipie(*inputPath)
	if err != nil {
		log.Fatalf("Failed to parse apipie JSON: %v", err)
	}

	overrides := loadOverrides(*overridesPath)

	resources := buildResources(doc, overrides)
	hardcoded := hardcodedResources()
	// Auto-derive EntityFields from Fields for hardcoded resources that don't have them
	for i := range hardcoded {
		if len(hardcoded[i].EntityFields) == 0 && len(hardcoded[i].Fields) > 0 {
			hardcoded[i].EntityFields = entityFromRequest(hardcoded[i].Fields)
		}
	}
	resources = append(resources, hardcoded...)

	if err := generateAll(resources, *outputDir, overrides); err != nil {
		log.Fatalf("Failed to generate client code: %v", err)
	}

	if err := generateFrameworkResources(resources, *providerDir, overrides); err != nil {
		log.Fatalf("Failed to generate framework resources: %v", err)
	}

	if err := generateTestFiles(resources, *outputDir, *providerDir, overrides); err != nil {
		log.Fatalf("Failed to generate test files: %v", err)
	}

	log.Printf("Generated %d resources: client in %s, framework in %s", len(resources), *outputDir, *providerDir)
}

// ---------------------------------------------------------------------------
// Apipie JSON model
// ---------------------------------------------------------------------------

type ApipieDoc struct {
	Docs struct {
		Resources map[string]ApipieRawResource `json:"resources"`
	} `json:"docs"`
}

type ApipieRawResource struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Methods []ApipieRawMethod `json:"methods"`
}

type ApipieRawMethod struct {
	Name     string        `json:"name"`
	Apis     []ApipieAPI   `json:"apis"`
	Examples []string      `json:"examples"`
	Params   []ApipieParam `json:"params"`
}

type ApipieAPI struct {
	HTTPMethod string `json:"http_method"`
	APILine    string `json:"api_url"`
}

type ApipieParam struct {
	Name         string        `json:"name"`
	Required     bool          `json:"required"`
	ExpectedType string        `json:"expected_type"`
	Description  string        `json:"description"`
	Validator    string        `json:"validator"`
	Params       []ApipieParam `json:"params"`
	Show         bool          `json:"show"`
}

type ResourceOverride struct {
	ShortName     string            `yaml:"short_name"`
	ExcludeFields []string          `yaml:"exclude_fields"`
	FieldTypes    map[string]string `yaml:"field_types"`
	// ParentEndpoint marks a resource nested under a single numeric parent
	// (e.g. "operatingsystems" for os_default_templates, whose real endpoint
	// is /api/operatingsystems/:operatingsystem_id/os_default_templates).
	// Wires into the same ParentEndpoint mechanism already used by the
	// hardcoded override_value resource: adds a Required "parent_id" int64
	// attribute and threads it through every generated URL.
	ParentEndpoint string `yaml:"parent_endpoint"`
	// FieldAliases maps an entity/response JSON field name to the request JSON
	// field name that represents the same logical attribute, for Rails
	// accepts_nested_attributes_for style APIs where the write key differs
	// from the read key (e.g. hostgroups: request "group_parameters_attributes"
	// reads back as "parameters").
	FieldAliases map[string]string `yaml:"field_aliases"`
	// AttributeRenames overrides the Terraform-facing schema attribute key
	// (and tfsdk struct tag) for a field, keyed by JSONName, without
	// changing the Foreman API wire name. Needed for fields whose JSON name
	// collides with a Terraform-reserved root attribute name, e.g.
	// compute_resources' "provider" field.
	AttributeRenames map[string]string `yaml:"attribute_renames"`
	// NotReturnedOnRead lists fields (by JSONName) Foreman accepts on
	// create/update but never actually includes in its response body,
	// confirmed against a real server (see GenField.NotReturnedOnRead).
	NotReturnedOnRead []string `yaml:"not_returned_on_read"`
}

type Overrides struct {
	SkipResources []string `yaml:"skip_resources"`
	// SkipDataSources marks a resource whose data source is ALSO
	// hand-written, unlike a plain SkipResources entry (which only means the
	// resource_*.go file is hand-written; the data source is still assumed
	// to fit the generic shape unless listed here too).
	SkipDataSources []string                    `yaml:"skip_data_sources"`
	Resources       map[string]ResourceOverride `yaml:"resources"`
}

// ---------------------------------------------------------------------------
// Intermediate representation
// ---------------------------------------------------------------------------

type GenResource struct {
	GoName         string
	ShortName      string // e.g. "domain" for resource type name
	EndpointBase   string
	ParamKey       string
	ParentEndpoint string // e.g. "smart_class_parameters/%d" for nested endpoints
	// OrgScopedQuery marks a Katello resource whose Create/Read/Query
	// endpoints (but not Update/Delete, which address the resource
	// directly by ID) require the provider-configured organization_id as
	// a "?organization_id=%d" query parameter, e.g. katello/products.
	OrgScopedQuery bool
	// SearchField is the entity's JSON field name to search/look up by, for
	// both the generated Query<Resource> client method and the data
	// source's required lookup attribute. Empty means "name" (the
	// overwhelmingly common case); only needs setting when a resource's own
	// display-name-like field is called something else (e.g.
	// smart_class_parameters' "parameter").
	SearchField  string
	HasCreate    bool
	HasUpdate    bool
	HasDelete    bool
	HasRead      bool
	HasIndex     bool
	Fields       []GenField
	EntityFields []GenField
}

// searchField returns res.SearchField, defaulting to "name".
func (res GenResource) searchField() string {
	if res.SearchField == "" {
		return "name"
	}
	return res.SearchField
}

type GenField struct {
	GoName       string
	JSONName     string
	GoType       string
	TFType       string // framework schema type: "String", "Int64", "Bool", "List"
	TFGoType     string // framework Go type: "types.String", "types.Int64", "types.Bool"
	Required     bool
	Description  string
	IsList       bool
	ListElemType string // for lists: "types.StringType", "types.Int64Type"
	// IsParametersMap marks a field that follows Foreman's standard
	// [{"name": ..., "value": ...}] convention (e.g. host_parameters_attributes,
	// group_parameters_attributes, os_parameters_attributes). Such fields are
	// exposed as types.Map (string->string) and bridged via the shared
	// flattenParameters/expandParameters helpers instead of an opaque JSON blob.
	IsParametersMap bool
	// ModelGoName, when set on a request Fields entry, names the EntityFields
	// GoName that the Terraform model actually uses for this logical
	// attribute (see FieldAliases). Codegen reads plan.<ModelGoName> instead
	// of plan.<GoName> when building the request body for such fields.
	ModelGoName string
	// Sensitive marks fields whose JSON name looks like a secret (password,
	// pass, secret), rendered with Sensitive: true in the framework schema.
	Sensitive bool
	// IsNestedList marks a field that is an array of objects with a real,
	// multi-field sub-schema (e.g. host's interfaces_attributes), as opposed
	// to the [{name,value}] convention (IsParametersMap) or an opaque single
	// hash. Such fields are exposed as a types.List of types.Object and
	// bridged via generated flatten/expand helper functions. NestedRequest
	// and NestedEntity describe the sub-object's own request/entity fields,
	// exactly like GenResource.Fields/EntityFields one level down.
	IsNestedList  bool
	NestedRequest []GenField
	NestedEntity  []GenField
	// TFName, when set, overrides JSONName as the Terraform-facing schema
	// attribute key and tfsdk struct tag (see ResourceOverride.AttributeRenames).
	// The Foreman API wire name (JSONName) is left untouched.
	TFName string
	// IsPolymorphicValue marks a standalone Foreman parameter's own "value"
	// field (e.g. common_parameters, parameters, smart_class_parameters -
	// detected by having a "parameter_type"/"hidden_value" sibling, the same
	// convention flattenParameters/expandParameters already handle for
	// nested *_parameters_attributes maps). Foreman parameters are
	// user-typed (string/boolean/integer/array/hash/yaml/json), so apidoc's
	// declared "string" type for this field is only the write-side
	// convenience shape - decoding a non-string response value (e.g. a real
	// JSON boolean/array) into a plain Go string fails the whole struct's
	// json.Unmarshal. The entity/response side is decoded as json.RawMessage
	// and rendered via parameterValueToString instead of stringValue; the
	// request/write side is unaffected (send the Terraform string as-is,
	// exactly like flattenParameters already does).
	IsPolymorphicValue bool
	// NotReturnedOnRead marks a writable field that Foreman accepts on
	// create/update but never actually includes in its response body
	// (confirmed case-by-case against a real server, not something apidoc's
	// examples can show - apidoc has zero response examples for several
	// resources with this quirk). Blindly copying the response's zero value
	// into state on every Create/Read/Update would silently wipe out
	// whatever the user configured. Such fields are skipped in
	// assignEntityFieldsFromResult, leaving the plan/state value from
	// before the API call untouched.
	NotReturnedOnRead bool
}

// tfKey returns the Terraform-facing schema attribute key/tfsdk tag for a
// field: TFName if overridden, otherwise the API's own JSONName.
func tfKey(f GenField) string {
	if f.TFName != "" {
		return f.TFName
	}
	return f.JSONName
}

// nestednGenResource wraps a nested object's request/entity field pair so
// the same modelFields/writeOnlyFields/fieldInModel/fieldIsComputed helpers
// used for top-level resources can be reused one level down.
func nestedGenResource(f GenField) GenResource {
	return GenResource{Fields: f.NestedRequest, EntityFields: f.NestedEntity}
}

// looksSensitive reports whether a JSON field name looks like it holds a
// secret value (password, pass suffix, secret) that should be marked
// Sensitive in the generated Terraform schema.
func looksSensitive(jsonName string) bool {
	lower := strings.ToLower(jsonName)
	return strings.Contains(lower, "password") ||
		strings.HasSuffix(lower, "_pass") ||
		strings.Contains(lower, "secret")
}

// nonNumericIDFields lists "_id"-suffixed fields confirmed to be genuinely
// non-numeric identifiers, not Foreman database primary keys, so they must
// be excluded from looksLikeNumericIDField's naming-convention override:
//   - cluster_id, storage_domain_id, storage_pod_id, project_domain_id,
//     vm_id (compute_resources): passthrough identifiers assigned by the
//     underlying hypervisor/cloud API (vSphere MOIDs, oVirt/OpenStack
//     UUIDs, AWS instance IDs, ...), not Foreman itself.
//   - progress_report_id (hosts): apidoc's own description says
//     "UUID to track orchestration tasks status".
var nonNumericIDFields = map[string]bool{
	"cluster_id":         true,
	"storage_domain_id":  true,
	"storage_pod_id":     true,
	"project_domain_id":  true,
	"vm_id":              true,
	"progress_report_id": true,
}

// looksLikeNumericIDField reports whether a singular "_id" field should be
// trusted as a plain integer FK despite apidoc declaring it a string (see
// nonNumericIDFields for the confirmed exceptions).
func looksLikeNumericIDField(jsonName string) bool {
	return strings.HasSuffix(jsonName, "_id") && !nonNumericIDFields[jsonName]
}

// ---------------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------------

func parseApipie(path string) (*ApipieDoc, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var doc ApipieDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("unmarshaling %s: %w", path, err)
	}
	return &doc, nil
}

func loadOverrides(path string) Overrides {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("No overrides file at %s, using defaults", path)
		return Overrides{Resources: make(map[string]ResourceOverride)}
	}
	var ov Overrides
	if err := yaml.Unmarshal(data, &ov); err != nil {
		log.Printf("Failed to parse overrides %s: %v, using defaults", path, err)
		return Overrides{Resources: make(map[string]ResourceOverride)}
	}
	if ov.Resources == nil {
		ov.Resources = make(map[string]ResourceOverride)
	}
	return ov
}

// ---------------------------------------------------------------------------
// Building resources
// ---------------------------------------------------------------------------

var skipResources = map[string]bool{
	// --- TRULY SKIP: no CRUD, read-only, or UI-only endpoints ---
	"home":                   true, // No entity, just API root
	"ping":                   true, // Health check, no entity
	"status":                 true, // Health check, no entity
	"plugins":                true, // Read-only listing, no CRUD
	"audits":                 true, // Read-only audit log
	"dashboard":              true, // UI-only endpoint
	"fact_values":            true, // Read-only data
	"permissions":            true, // Read-only RBAC data
	"tasks":                  true, // Background task tracking
	"auth_sources":           true, // Read-only, no CRUD
	"auth_source_externals":  true,
	"auth_source_internals":  true,
	"auth_source_ldaps":      true,
	"bookmarks":              true, // UI bookmarks, not manageable
	"config_reports":         true, // Read-only report data
	"filters":                true, // RBAC filters, read-only
	"host_statuses":          true, // Read-only status types
	"hosts_bulk_actions":     true, // Action endpoint, not entity
	"instance_hosts":         true, // Internal
	"interfaces":             true, // Nested under hosts, managed via host's interfaces_attributes
	"locations":              true, // Taxonomy, managed at provider level via addTaxonomy()
	"mail_notifications":     true, // Read-only
	"organizations":          true, // Taxonomy, managed at provider level via addTaxonomy()
	"personal_access_tokens": true, // Security-sensitive
	"registration":           true, // Action endpoint
	"registration_commands":  true, // Action endpoint
	"registration_tokens":    true, // Security-sensitive
	"report_templates":       true, // Read-only
	"roles":                  true, // Read-only RBAC
	"smart_proxy_hosts":      true, // Read-only association
	"ssh_keys":               true, // Security-sensitive
	"table_preferences":      true, // UI preferences
	"template_combinations":  true, // Nested, managed via provisioning_template_ids on OS/hostgroup
	"external_usergroups":    true, // Nested under usergroups
	"compute_attributes":     true, // Read-only
	"template_kinds":         true, // Read-only reference data
	// --- TODO(bridget): these need bridging code, NOT permanent exclusion ---
	// TODO(bridget): realms — old provider did NOT manage realms. If needed, remove from skip list.
	// TODO(bridget): compute_attributes — read-only compute profile attributes. Excluded as read-only.
}

// taxonomyFields excludes only the singular organization_id/location_id:
// the provider-level addTaxonomy() mechanism already injects these into
// every request from the provider's own config, so a per-resource field for
// them would be redundant/conflicting. The plural forms (location_ids/
// organization_ids) are a different, genuine per-resource feature - "which
// locations/organizations can use this record" - unrelated to request
// scoping, and apidoc does declare them for some resources (e.g. users);
// they must NOT be excluded here.
var taxonomyFields = map[string]bool{
	"location_id": true, "organization_id": true,
}

func buildResources(doc *ApipieDoc, overrides Overrides) []GenResource {
	var resources []GenResource

	for rawID, raw := range doc.Docs.Resources {
		if skipResources[rawID] {
			if verbose {
				log.Printf("Skipping %s (in skip list)", rawID)
			}
			continue
		}

		override := overrides.Resources[rawID]

		epBase := rawID
		typeName := singularizePascal(rawID)

		shortName := override.ShortName
		if shortName == "" {
			shortName = strings.ToLower(typeName)
		}

		res := GenResource{
			GoName:       "Foreman" + typeName,
			ShortName:    shortName,
			EndpointBase: epBase,
			ParamKey:     extractParamKey(raw.Methods),
		}

		for _, m := range raw.Methods {
			switch m.Name {
			case "create":
				res.HasCreate = true
			case "show":
				res.HasRead = true
			case "update":
				res.HasUpdate = true
			case "destroy":
				res.HasDelete = true
			case "index":
				res.HasIndex = true
			}
		}

		res.Fields = buildFields(raw.Methods)
		res.EntityFields = buildEntityFields(raw.Methods)
		res.EntityFields = normalizeEntityFieldTypes(res.Fields, res.EntityFields)

		applyFieldOverrides(&res, override)
		applyFieldAliases(&res, override)

		resources = append(resources, res)
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].GoName < resources[j].GoName
	})

	return resources
}

// hardcodedResources returns resources not in the apidoc but needed for parity
// with the old provider (environment, jobtemplate, puppetclass, smartclassparameter, templatekind).
func hardcodedResources() []GenResource {
	return []GenResource{
		{
			GoName:       "ForemanEnvironment",
			ShortName:    "environment",
			EndpointBase: "environments",
			ParamKey:     "environment",
			HasCreate:    true, HasRead: true, HasUpdate: true, HasDelete: true, HasIndex: true,
			Fields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
			},
			// EntityFields derived from Fields via entityFromRequest in buildResources
		},
		{
			GoName:       "ForemanJobTemplate",
			ShortName:    "jobtemplate",
			EndpointBase: "job_templates",
			ParamKey:     "job_template",
			HasCreate:    true, HasRead: true, HasUpdate: true, HasDelete: true, HasIndex: true,
			Fields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "description", GoName: "Description", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "description_format", GoName: "DescriptionFormat", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "template", GoName: "Template", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "locked", GoName: "Locked", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "job_category", GoName: "JobCategory", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "provider_type", GoName: "ProviderType", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "snippet", GoName: "Snippet", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
			},
			// EntityFields derived from Fields via entityFromRequest in buildResources
		},
		{
			GoName:       "ForemanPuppetClass",
			ShortName:    "puppetclass",
			EndpointBase: "puppetclasses",
			ParamKey:     "puppetclass",
			HasCreate:    true, HasRead: true, HasUpdate: true, HasDelete: true, HasIndex: true,
			Fields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
			},
			// EntityFields derived from Fields via entityFromRequest in buildResources
		},
		{
			GoName:       "ForemanSmartClassParameter",
			ShortName:    "smartclassparameter",
			EndpointBase: "smart_class_parameters",
			ParamKey:     "smart_class_parameter",
			// The entity's own display-name-like field is "parameter", not
			// "name" (it has no "name" field at all) - searching "name"
			// unconditionally returned zero results. KNOWN LIMITATION: this
			// searches globally across every puppet class rather than
			// scoping to one via puppetclass_id (the old provider's data
			// source did scope this way), so a parameter name shared by two
			// classes is ambiguous; fully fixing that needs parent-scoped
			// Query support the generic pipeline doesn't have yet.
			SearchField: "parameter",
			// Read-only: matches the old provider exactly, which never had a
			// foreman_smartclassparameter *resource* at all (only this same
			// data source) - not in apidoc (Puppet plugin resource) so there's
			// no source of truth for a real Update request shape, and an
			// earlier HasUpdate:true here always sent an empty request body
			// (Fields was empty too), silently updating nothing on every
			// apply. Same precedent as foreman_templatekind.
			HasCreate: false, HasRead: true, HasUpdate: false, HasDelete: false, HasIndex: true,
			Fields: []GenField{},
			EntityFields: []GenField{
				{JSONName: "parameter", GoName: "Parameter", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "puppetclass_id", GoName: "PuppetclassID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "override", GoName: "Override", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "description", GoName: "Description", GoType: "string", TFType: "String", TFGoType: "types.String"},
				// GoType is json.RawMessage, not string: Foreman parameters
				// are user-typed (string/boolean/integer/array/hash/yaml/
				// json), so a non-string default_value would otherwise fail
				// json.Unmarshal for the whole struct. Rendered via
				// parameterValueToString like common_parameters/parameters'
				// "value" field (see IsPolymorphicValue).
				{JSONName: "default_value", GoName: "DefaultValue", GoType: "json.RawMessage", TFType: "String", TFGoType: "types.String", IsPolymorphicValue: true},
				{JSONName: "hidden_value", GoName: "HiddenValue", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
			},
		},
		{
			GoName:       "ForemanTemplateKind",
			ShortName:    "templatekind",
			EndpointBase: "template_kinds",
			ParamKey:     "template_kind",
			HasCreate:    false, HasRead: true, HasUpdate: false, HasDelete: false, HasIndex: true,
			Fields: []GenField{},
			EntityFields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String"},
			},
		},
		// Plugin resources (not in core Foreman apidoc)
		// Note: override_values skipped — nested endpoint (smart_class_parameters/%d/override_values)
		// KNOWN GAP: the old provider also supported location_ids/organization_ids
		// here, but its own Read implementation had to decode them from a
		// *different* nested shape than it wrote (write: flat int arrays;
		// read: {"organizations": [{"id":.., "name":..}, ...]}) - a genuine
		// asymmetric bridge, not something this hardcoded declarative Fields/
		// EntityFields shape can express without becoming a fully hand-written
		// resource. Not in apidoc (discovery_rules isn't a core-Foreman
		// resource at all) so there's no source of truth to verify the exact
		// shape against; left out rather than guessed.
		{
			GoName:       "ForemanDiscoveryRule",
			ShortName:    "discovery_rule",
			EndpointBase: "discovery_rules",
			ParamKey:     "discovery_rule",
			HasCreate:    true, HasRead: true, HasUpdate: true, HasDelete: true, HasIndex: true,
			Fields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "search", GoName: "Search", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "hostgroup_id", GoName: "HostgroupID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64", Required: true},
				{JSONName: "hostname", GoName: "Hostname", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "max_count", GoName: "HostsLimitMaxCount", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "priority", GoName: "Priority", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "enabled", GoName: "Enabled", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
			},
			EntityFields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "search", GoName: "Search", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "hostgroup_id", GoName: "HostgroupID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "hostname", GoName: "Hostname", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "hosts_limit", GoName: "HostsLimitMaxCount", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "priority", GoName: "Priority", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "enabled", GoName: "Enabled", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
			},
		},
		{
			GoName:       "ForemanWebhook",
			ShortName:    "webhook",
			EndpointBase: "webhooks",
			ParamKey:     "webhook",
			HasCreate:    true, HasRead: true, HasUpdate: true, HasDelete: true, HasIndex: true,
			Fields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "target_url", GoName: "TargetURL", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				// Optional, not Required: Foreman defaults this server-side
				// when omitted (old provider used Optional+Computed).
				{JSONName: "http_method", GoName: "HTTPMethod", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "http_content_type", GoName: "HTTPContentType", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "http_headers", GoName: "HTTPHeaders", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "event", GoName: "Event", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "enabled", GoName: "Enabled", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "verify_ssl", GoName: "VerifySSL", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "ssl_ca_certs", GoName: "SSLCACerts", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "proxy_authorization", GoName: "ProxyAuthorization", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "user", GoName: "User", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "password", GoName: "Password", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "webhook_template_id", GoName: "WebhookTemplateID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
			},
			// EntityFields are different from Fields (User and Password are not returned in response)
			EntityFields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "target_url", GoName: "TargetURL", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "http_method", GoName: "HTTPMethod", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "http_content_type", GoName: "HTTPContentType", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "http_headers", GoName: "HTTPHeaders", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "event", GoName: "Event", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "enabled", GoName: "Enabled", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "verify_ssl", GoName: "VerifySSL", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "ssl_ca_certs", GoName: "SSLCACerts", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "proxy_authorization", GoName: "ProxyAuthorization", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "webhook_template_id", GoName: "WebhookTemplateID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
			},
		},
		{
			GoName:       "ForemanWebhookTemplate",
			ShortName:    "webhooktemplate",
			EndpointBase: "webhook_templates",
			ParamKey:     "webhook_template",
			HasCreate:    true, HasRead: true, HasUpdate: true, HasDelete: true, HasIndex: true,
			Fields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "template", GoName: "Template", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "snippet", GoName: "Snippet", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "audit_comment", GoName: "AuditComment", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "locked", GoName: "Locked", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "default", GoName: "Default", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "description", GoName: "Description", GoType: "string", TFType: "String", TFGoType: "types.String"},
			},
			// EntityFields derived from Fields via entityFromRequest in buildResources
		},
		{
			GoName:         "ForemanOverrideValue",
			ShortName:      "override_value",
			EndpointBase:   "override_values",
			ParamKey:       "override_value",
			ParentEndpoint: "smart_class_parameters",
			HasCreate:      true, HasRead: true, HasUpdate: true, HasDelete: true, HasIndex: false,
			Fields: []GenField{
				{JSONName: "match", GoName: "Match", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "value", GoName: "Value", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "omit", GoName: "Omit", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
			},
			// EntityFields derived from Fields via entityFromRequest in buildResources
		},
		// Katello resources not in the pinned core apidoc/v2.json. Only the
		// plain-CRUD ones fit here: content_view (publish + filter sync),
		// lifecycle_environment (flattens nested prior/successor objects),
		// repository (suppressDownloadConcurrencyDiff plan modifier), and
		// sync_plan (org id is part of the URL path, not a query param) all
		// need logic this generic pipeline doesn't express, and stay
		// hand-written in generated/katello_*.go / internal/provider/*_katello_*.go.
		{
			GoName:         "ForemanKatelloContentCredential",
			ShortName:      "katello_content_credential",
			EndpointBase:   "katello/content_credentials",
			ParamKey:       "content_credential",
			OrgScopedQuery: false,
			HasCreate:      true, HasRead: true, HasUpdate: true, HasDelete: true, HasIndex: true,
			Fields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "content", GoName: "Content", GoType: "string", TFType: "String", TFGoType: "types.String"},
			},
			// EntityFields derived from Fields via entityFromRequest in buildResources
		},
		{
			GoName:         "ForemanKatelloProduct",
			ShortName:      "katello_product",
			EndpointBase:   "katello/products",
			ParamKey:       "product",
			OrgScopedQuery: true,
			HasCreate:      true, HasRead: true, HasUpdate: true, HasDelete: true, HasIndex: true,
			Fields: []GenField{
				{JSONName: "name", GoName: "Name", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
				{JSONName: "description", GoName: "Description", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "label", GoName: "Label", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "gpg_key_id", GoName: "GpgKeyID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "ssl_ca_cert_id", GoName: "SslCaCertID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "ssl_client_cert_id", GoName: "SslClientCertID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "ssl_client_key_id", GoName: "SslClientKeyID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "sync_plan_id", GoName: "SyncPlanID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
			},
			// EntityFields derived from Fields via entityFromRequest in buildResources
		},
	}
}

// entityFromRequest derives EntityFields from Fields by copying all fields
// but removing the Required flag. This reduces duplication for hardcoded resources
// where request and response fields are identical.
func entityFromRequest(fields []GenField) []GenField {
	result := make([]GenField, len(fields))
	for i, f := range fields {
		result[i] = GenField{
			JSONName:        f.JSONName,
			GoName:          f.GoName,
			GoType:          f.GoType,
			TFType:          f.TFType,
			TFGoType:        f.TFGoType,
			Description:     f.Description,
			IsList:          f.IsList,
			ListElemType:    f.ListElemType,
			IsParametersMap: f.IsParametersMap,
			Sensitive:       f.Sensitive,
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Field building
// ---------------------------------------------------------------------------

func buildFields(methods []ApipieRawMethod) []GenField {
	for _, m := range methods {
		if m.Name != "create" && m.Name != "update" {
			continue
		}
		fields := extractHashParams(m.Params)
		if len(fields) > 0 {
			return fields
		}
	}
	return nil
}

func buildEntityFields(methods []ApipieRawMethod) []GenField {
	// Skip index/list methods - they return paginated responses, not entity fields
	// Try to get fields from show/create/update examples first
	for _, m := range methods {
		name := strings.ToLower(m.Name)
		if name == "index" || name == "" {
			continue
		}
		for _, ex := range m.Examples {
			body := parseExampleBody(ex)
			if body == nil {
				continue
			}
			// Skip paginated list responses
			if _, ok := body["results"]; ok {
				continue
			}
			if _, ok := body["total"]; ok {
				continue
			}
			var fields []GenField
			for k, v := range body {
				if k == "id" || k == "name" || k == "created_at" || k == "updated_at" {
					continue
				}
				if taxonomyFields[k] {
					continue
				}
				f := GenField{
					JSONName:  k,
					GoName:    toPascal(k),
					Sensitive: looksSensitive(k),
				}
				setInferredType(&f, v)
				fields = append(fields, f)
			}
			if len(fields) > 0 {
				sortFields(fields)
				return fields
			}
		}
	}
	// Last resort: use all request fields
	return buildFields(methods)
}

// nestedEntitySkipKeys lists keys that show up inside nested response
// objects (e.g. one element of host's "interfaces" array) but are never
// useful as their own Terraform attribute: timestamps, and the parent
// resource's own id/name/fqdn echoed back on each child element.
var nestedEntitySkipKeys = map[string]bool{
	"created_at": true, "updated_at": true,
	"host_id": true, "host_name": true, "fqdn": true,
}

// buildNestedEntityFields builds the entity-side field schema for one
// element of a nested object array (e.g. host's "interfaces"), from the
// union of every example element's keys. "_name" keys that duplicate a
// sibling "_id" key are dropped, mirroring the top-level exclude_fields
// convention for read-only display duplicates.
func buildNestedEntityFields(arr []interface{}) []GenField {
	values := make(map[string]interface{})
	var order []string
	for _, elem := range arr {
		obj, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}
		for k, v := range obj {
			if existing, seen := values[k]; !seen {
				order = append(order, k)
				values[k] = v
			} else if existing == nil && v != nil {
				values[k] = v
			}
		}
	}

	// "id" is legitimate and useful on a nested object (e.g. an interface's
	// own id), unlike the top-level resource id which is handled separately.
	var fields []GenField
	for _, k := range order {
		if k == "created_at" || k == "updated_at" || nestedEntitySkipKeys[k] {
			continue
		}
		if strings.HasSuffix(k, "_name") {
			base := strings.TrimSuffix(k, "_name")
			if _, hasID := values[base+"_id"]; hasID {
				continue
			}
		}
		f := GenField{
			JSONName:  k,
			GoName:    toPascal(k),
			Sensitive: looksSensitive(k),
		}
		if values[k] == nil {
			// A null example value carries no real type information, and
			// "string" (setInferredType's nil default) is often wrong for
			// something like a foreign-key id that just happens to be unset
			// in the example. Leave GoType empty so reconcileFieldPair
			// unconditionally trusts the request-side type instead of
			// requiring isCompatibleType to agree with this guess.
			f.TFType = "String"
			f.TFGoType = "types.String"
		} else {
			setInferredType(&f, values[k])
		}
		fields = append(fields, f)
	}
	sortFields(fields)
	return fields
}

// normalizeEntityFieldTypes normalizes entity field types to match request field types
// where they exist in both. This reduces type mismatches between request and response structs.
// Fields that are genuinely dynamic (e.g., compute_profile_id which is string in response
// but int in request) are left as-is.
func normalizeEntityFieldTypes(fields []GenField, entityFields []GenField) []GenField {
	if len(fields) == 0 || len(entityFields) == 0 {
		return entityFields
	}

	// Build a map of request fields by JSON name
	fieldByName := make(map[string]GenField)
	for _, f := range fields {
		fieldByName[f.JSONName] = f
	}

	result := make([]GenField, len(entityFields))
	for i, ef := range entityFields {
		result[i] = ef
		reqField, ok := fieldByName[ef.JSONName]
		if !ok {
			continue
		}
		reconcileFieldPair(&reqField, &result[i])
	}
	return result
}

// reconcileFieldPair aligns an entity field's shape/type with the request
// field representing the same logical attribute (matched either because
// their JSON names are identical, or via a field_alias in applyFieldAliases).
// The request side is trusted as the source of truth for shape detection
// (IsParametersMap/IsNestedList) because it comes from apidoc's declared
// param schema, while the entity side is inferred from example JSON that is
// frequently empty or missing.
func reconcileFieldPair(reqField *GenField, entityField *GenField) {
	switch {
	case reqField.IsPolymorphicValue:
		// The response side alone is decoded as json.RawMessage (any JSON
		// shape unmarshals into it without error); the write side keeps its
		// plain string type (see IsPolymorphicValue's own doc comment).
		entityField.IsPolymorphicValue = true
		entityField.GoType = "json.RawMessage"
		entityField.TFType = "String"
		entityField.TFGoType = "types.String"
	case reqField.IsParametersMap:
		entityField.IsParametersMap = true
		entityField.IsList = false
		entityField.GoType = "json.RawMessage"
		entityField.TFType = "Map"
		entityField.TFGoType = "types.Map"
	case reqField.IsNestedList:
		entityField.IsNestedList = true
		entityField.IsList = false
		entityField.GoType = "json.RawMessage"
		entityField.TFType = "ListNested"
		entityField.TFGoType = "types.List"
		entityField.NestedRequest = reqField.NestedRequest
		if len(entityField.NestedEntity) == 0 {
			// No usable response example for this field; fall back to the
			// request shape so the resource still has a schema for it.
			entityField.NestedEntity = entityFromRequest(reqField.NestedRequest)
		}
		entityField.NestedEntity = normalizeEntityFieldTypes(reqField.NestedRequest, entityField.NestedEntity)
	default:
		// entityField.GoType == "" means the entity side never actually
		// observed a non-null example value (see buildNestedEntityFields),
		// so there is no real signal to conflict with the request type -
		// trust it unconditionally instead of requiring isCompatibleType.
		// Likewise, a "_id" field's response example is sometimes
		// serialized as a JSON string even though the real attribute is a
		// plain integer FK (the same Rails string-ification apidoc's
		// request-side docs already show) - trust looksLikeNumericIDField
		// over whatever the entity side happened to observe.
		trustRequest := entityField.GoType == "" ||
			isCompatibleType(entityField.GoType, reqField.GoType) ||
			(reqField.GoType == "int64" && looksLikeNumericIDField(reqField.JSONName))
		if trustRequest {
			entityField.GoType = reqField.GoType
			entityField.TFType = goTypeToTFType(reqField.GoType)
			entityField.TFGoType = goTypeToTFGoType(reqField.GoType)
		}
	}
}

// isCompatibleType returns true if two Go types can be safely normalized.
// We only normalize between int64/float64 (both numeric) or within string-like types.
func isCompatibleType(a, b string) bool {
	numericTypes := map[string]bool{"int64": true, "float64": true}
	stringTypes := map[string]bool{"string": true}

	if numericTypes[a] && numericTypes[b] {
		return true
	}
	if stringTypes[a] && stringTypes[b] {
		return true
	}
	return false
}

// tfTypeForIDList returns "Set" for a "_ids"-suffixed int64 list field, and
// "List" otherwise. Foreman returns these many-to-many FK reference
// collections in its own order (not necessarily the order the config wrote
// them in); a List requires an exact order match or Terraform shows a
// perpetual diff, while a Set is naturally order-independent.
func tfTypeForIDList(jsonName string) string {
	if strings.HasSuffix(jsonName, "_ids") {
		return "Set"
	}
	return "List"
}

// tfGoTypeForIDList is the types.Set/types.List counterpart of tfTypeForIDList.
func tfGoTypeForIDList(jsonName string) string {
	if strings.HasSuffix(jsonName, "_ids") {
		return "types.Set"
	}
	return "types.List"
}

// goTypeToTFType converts a Go type string to the corresponding Terraform framework type.
func goTypeToTFType(goType string) string {
	switch goType {
	case "string":
		return "String"
	case "int64":
		return "Int64"
	case "float64":
		return "Float64"
	case "bool":
		return "Bool"
	default:
		return "String"
	}
}

// goTypeToTFGoType converts a Go type string to the corresponding Terraform framework Go type.
func goTypeToTFGoType(goType string) string {
	switch goType {
	case "string":
		return "types.String"
	case "int64":
		return "types.Int64"
	case "float64":
		return "types.Float64"
	case "bool":
		return "types.Bool"
	default:
		return "types.String"
	}
}

func setInferredType(f *GenField, v interface{}) {
	switch val := v.(type) {
	case string:
		if looksLikeNumericIDField(f.JSONName) {
			// Same Rails route-string effect as paramToGenField's request-side
			// case: a response example can show a "_id" field as a quoted
			// string even though the real attribute is a plain integer FK.
			f.GoType = "int64"
			f.TFType = "Int64"
			f.TFGoType = "types.Int64"
			return
		}
		f.GoType = "string"
		f.TFType = "String"
		f.TFGoType = "types.String"
	case bool:
		f.GoType = "bool"
		f.TFType = "Bool"
		f.TFGoType = "types.Bool"
	case float64:
		// Use string conversion to avoid float64 precision loss for large integers
		s := strconv.FormatFloat(val, 'f', -1, 64)
		if strings.Contains(s, ".") {
			f.GoType = "float64"
			f.TFType = "Float64"
			f.TFGoType = "types.Float64"
		} else {
			f.GoType = "int64"
			f.TFType = "Int64"
			f.TFGoType = "types.Int64"
		}
	case map[string]interface{}:
		// Free-form/heterogeneous nested object (e.g. host's permissions,
		// compute_attributes). Expose as an opaque JSON string rather than a
		// typed Map, since Map requires uniform value types and these
		// objects' value types vary per key/resource.
		f.GoType = "json.RawMessage"
		f.TFType = "String"
		f.TFGoType = "types.String"
	case []interface{}:
		f.IsList = true
		if len(val) > 0 {
			switch val[0].(type) {
			case float64:
				f.GoType = "[]int64"
				f.ListElemType = "types.Int64Type"
				f.TFType = tfTypeForIDList(f.JSONName)
				f.TFGoType = tfGoTypeForIDList(f.JSONName)
				return
			case string:
				f.GoType = "[]string"
				f.ListElemType = "types.StringType"
			case map[string]interface{}:
				if isNameValueArrayExample(val) {
					// Foreman's standard [{"name": ..., "value": ...}] convention.
					f.IsParametersMap = true
					f.GoType = "json.RawMessage"
					f.TFType = "Map"
					f.TFGoType = "types.Map"
					f.IsList = false
					return
				}
				// Array of objects with a real multi-field sub-schema (e.g.
				// host's "interfaces"). Build the nested entity schema from
				// the union of every element's keys.
				f.IsNestedList = true
				f.NestedEntity = buildNestedEntityFields(val)
				f.GoType = "json.RawMessage"
				f.TFType = "ListNested"
				f.TFGoType = "types.List"
				f.IsList = false
				return
			default:
				f.GoType = "[]interface{}"
				f.ListElemType = "types.StringType"
			}
		} else {
			f.GoType = "[]interface{}"
			f.ListElemType = "types.StringType"
		}
		f.TFType = "List"
		f.TFGoType = "types.List"
	default:
		// Typically a null example value with no other type signal.
		if looksLikeNumericIDField(f.JSONName) {
			f.GoType = "int64"
			f.TFType = "Int64"
			f.TFGoType = "types.Int64"
			return
		}
		f.GoType = "string"
		f.TFType = "String"
		f.TFGoType = "types.String"
	}
}

// isNameValueParams reports whether a declared apipie array param's nested
// schema matches Foreman's standard [{"name": ..., "value": ...}] convention
// (used for host/hostgroup/os "parameters" style attributes). Extra optional
// sub-params (e.g. parameter_type, hidden_value) are allowed.
func isNameValueParams(params []ApipieParam) bool {
	hasName, hasValue := false, false
	for _, p := range params {
		switch strings.ToLower(p.Name) {
		case "name":
			hasName = true
		case "value":
			hasValue = true
		}
	}
	return hasName && hasValue
}

// isNameValueArrayExample reports whether every object element in a response
// example array matches Foreman's standard [{"name": ..., "value": ...}]
// convention.
func isNameValueArrayExample(arr []interface{}) bool {
	for _, elem := range arr {
		obj, ok := elem.(map[string]interface{})
		if !ok {
			return false
		}
		_, hasName := obj["name"]
		_, hasValue := obj["value"]
		if !hasName || !hasValue {
			return false
		}
	}
	return true
}

func extractParamKey(methods []ApipieRawMethod) string {
	for _, m := range methods {
		if m.Name != "create" && m.Name != "update" {
			continue
		}
		for _, p := range m.Params {
			if p.ExpectedType == "hash" && len(p.Params) > 0 {
				return p.Name
			}
		}
	}
	return ""
}

func extractHashParams(params []ApipieParam) []GenField {
	for _, p := range params {
		if p.ExpectedType == "hash" && len(p.Params) > 0 {
			var fields []GenField
			for _, np := range p.Params {
				if taxonomyFields[np.Name] {
					continue
				}
				f := paramToGenField(np)
				fields = append(fields, f)
			}
			markPolymorphicValueField(fields)
			sortFields(fields)
			return fields
		}
	}
	return nil
}

// markPolymorphicValueField detects Foreman's standalone-parameter-resource
// convention (a "value" field alongside "parameter_type" and/or
// "hidden_value" siblings - the same shape common_parameters, parameters,
// and smart_class_parameters all share) and marks it IsPolymorphicValue, so
// its response side is decoded as json.RawMessage instead of a plain string.
func markPolymorphicValueField(fields []GenField) {
	hasSibling := false
	valueIdx := -1
	for i, f := range fields {
		switch f.JSONName {
		case "value":
			valueIdx = i
		case "parameter_type", "hidden_value":
			hasSibling = true
		}
	}
	if valueIdx >= 0 && hasSibling {
		fields[valueIdx].IsPolymorphicValue = true
	}
}

func paramToGenField(p ApipieParam) GenField {
	f := GenField{
		JSONName:    p.Name,
		GoName:      toPascal(p.Name),
		Required:    p.Required,
		Description: cleanDesc(p.Description),
		Sensitive:   looksSensitive(p.Name),
	}

	switch p.ExpectedType {
	case "string":
		if looksLikeNumericIDField(p.Name) {
			// apidoc reports almost every "_id" field as a plain string,
			// regardless of the underlying column type: Rails route/path
			// segments are lexically strings even when the attribute is a
			// plain integer FK, and apipie-rails' introspection appears to
			// surface that route-level type rather than the real one. Trust
			// the naming convention instead, except for the known handful of
			// genuinely non-numeric external identifiers (see
			// nonNumericIDFields).
			f.GoType = "int64"
			f.TFType = "Int64"
			f.TFGoType = "types.Int64"
		} else {
			f.GoType = "string"
			f.TFType = "String"
			f.TFGoType = "types.String"
		}
	case "numeric":
		f.GoType = "int64"
		f.TFType = "Int64"
		f.TFGoType = "types.Int64"
	case "boolean":
		f.GoType = "bool"
		f.TFType = "Bool"
		f.TFGoType = "types.Bool"
	case "hash":
		// Free-form/heterogeneous nested object (e.g. host's compute_attributes,
		// whose shape varies per compute resource type). Expose as an opaque
		// JSON string rather than a typed Map, since Map requires uniform
		// value types.
		f.GoType = "json.RawMessage"
		f.TFType = "String"
		f.TFGoType = "types.String"
	case "array":
		if strings.HasSuffix(p.Name, "_parameters_attributes") || (len(p.Params) > 0 && isNameValueParams(p.Params)) {
			// Foreman's standard [{"name": ..., "value": ...}] convention.
			// The "_parameters_attributes" suffix alone is enough to trust
			// this even when apidoc declares no sub-params at all (true for
			// domains/subnets - an apidoc documentation gap, not a real
			// difference from hosts/hostgroups/operatingsystems, which
			// happen to have their sub-params fully declared).
			// The request wire type matches flattenParameters' return type;
			// the entity/response side stays json.RawMessage for
			// expandParameters (see setInferredType/normalizeEntityFieldTypes).
			f.IsParametersMap = true
			f.GoType = "[]map[string]interface{}"
			f.TFType = "Map"
			f.TFGoType = "types.Map"
		} else if len(p.Params) > 0 {
			// Array of objects with a real multi-field sub-schema (e.g.
			// host's interfaces_attributes). Recurse to build the nested
			// request schema. Wire type matches the flatten helper's return
			// type (see reconcileFieldPair/addNestedListHelpers); the
			// entity/response side stays json.RawMessage for expand.
			f.IsNestedList = true
			for _, sp := range p.Params {
				f.NestedRequest = append(f.NestedRequest, paramToGenField(sp))
			}
			f.GoType = "[]map[string]interface{}"
			f.TFType = "ListNested"
			f.TFGoType = "types.List"
		} else if strings.Contains(p.Validator, "String") || (!strings.HasSuffix(p.Name, "_ids") && !strings.Contains(p.Validator, "number")) {
			// Foreman's apidoc frequently reports array params as the
			// generic "Must be an array of any type" regardless of actual
			// element type. The "_ids" suffix is Rails' overwhelmingly
			// consistent convention for integer FK arrays; anything else
			// ambiguous (e.g. attached_devices, logs) is far more likely to
			// be a list of identifiers/strings than integers.
			f.IsList = true
			f.GoType = "[]string"
			f.TFType = "List"
			f.TFGoType = "types.List"
			f.ListElemType = "types.StringType"
		} else {
			f.IsList = true
			f.GoType = "[]int64"
			f.TFType = tfTypeForIDList(p.Name)
			f.TFGoType = tfGoTypeForIDList(p.Name)
			f.ListElemType = "types.Int64Type"
		}
	default:
		f.GoType = "string"
		f.TFType = "String"
		f.TFGoType = "types.String"
	}

	return f
}

func parseExampleBody(exStr string) map[string]interface{} {
	// Format: "GET /path\nSTATUS\n{JSON_BODY}"
	lines := strings.SplitN(exStr, "\n", 3)
	if len(lines) < 3 {
		return nil
	}
	// Try the top-level body
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(lines[2]), &body); err != nil {
		return nil
	}
	// If the body has a single key with an object value (e.g. {"domain": {...}}),
	// return the inner object
	if len(body) == 1 {
		for k, v := range body {
			if inner, ok := v.(map[string]interface{}); ok && k != "results" {
				return inner
			}
		}
	}
	return body
}

var htmlTagRe = regexp.MustCompile("<[^>]*>")

func cleanDesc(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = htmlTagRe.ReplaceAllString(s, "")
	// Escape Go string special characters
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.TrimSpace(s)
	return s
}

// ---------------------------------------------------------------------------
// Overrides
// ---------------------------------------------------------------------------

func applyFieldOverrides(res *GenResource, ov ResourceOverride) {
	if ov.ParentEndpoint != "" {
		res.ParentEndpoint = ov.ParentEndpoint
	}

	exclude := make(map[string]bool)
	for _, n := range ov.ExcludeFields {
		exclude[n] = true
	}

	res.Fields = filterFields(res.Fields, exclude, ov.FieldTypes, ov.AttributeRenames)
	res.EntityFields = filterFields(res.EntityFields, exclude, ov.FieldTypes, ov.AttributeRenames)

	if len(ov.NotReturnedOnRead) > 0 {
		notReturned := make(map[string]bool, len(ov.NotReturnedOnRead))
		for _, n := range ov.NotReturnedOnRead {
			notReturned[n] = true
		}
		for i := range res.EntityFields {
			if notReturned[res.EntityFields[i].JSONName] {
				res.EntityFields[i].NotReturnedOnRead = true
			}
		}
	}
}

func filterFields(fields []GenField, exclude map[string]bool, types map[string]string, renames map[string]string) []GenField {
	var result []GenField
	for _, f := range fields {
		if exclude[f.JSONName] {
			continue
		}
		if t, ok := types[f.JSONName]; ok {
			f.GoType = t
			// Keep the framework schema type consistent with the Go type
			// override for canonical types; leave non-canonical overrides
			// (e.g. "int" used purely for the wire struct) untouched.
			if t == "string" || t == "int64" || t == "float64" || t == "bool" {
				f.TFType = goTypeToTFType(t)
				f.TFGoType = goTypeToTFGoType(t)
			}
		}
		if n, ok := renames[f.JSONName]; ok {
			f.TFName = n
		}
		result = append(result, f)
	}
	return result
}

// applyFieldAliases reconciles request/response fields that represent the
// same logical Terraform attribute but use different JSON keys (Rails
// accepts_nested_attributes_for convention, e.g. hostgroups' request key
// "group_parameters_attributes" reads back as "parameters"). The entity
// field is treated as canonical for the Terraform model; the request field
// is tagged with ModelGoName so codegen knows to read the plan value from
// the entity field's Go name instead of its own.
func applyFieldAliases(res *GenResource, ov ResourceOverride) {
	for entityName, requestName := range ov.FieldAliases {
		var entityField *GenField
		for i := range res.EntityFields {
			if res.EntityFields[i].JSONName == entityName {
				entityField = &res.EntityFields[i]
				break
			}
		}
		var reqField *GenField
		for i := range res.Fields {
			if res.Fields[i].JSONName == requestName {
				reqField = &res.Fields[i]
				break
			}
		}
		if entityField == nil || reqField == nil {
			continue
		}
		reconcileFieldPair(reqField, entityField)
		reqField.ModelGoName = entityField.GoName
	}
}

// ---------------------------------------------------------------------------
// Code generation
// ---------------------------------------------------------------------------

func generateAll(resources []GenResource, outputDir string, overrides Overrides) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}

	for _, res := range resources {
		// skip_resources means the client file is hand-written too (e.g. a
		// resource whose shape the generic pipeline can't express, such as
		// autosign's string-typed, parent-scoped, no-show entity).
		skip := false
		for _, s := range overrides.SkipResources {
			if s == res.ShortName || s == res.EndpointBase {
				skip = true
				break
			}
		}
		if skip {
			if verbose {
				log.Printf("Skipping client file %s (in skip_resources, hand-written)", res.ShortName)
			}
			continue
		}
		path := filepath.Join(outputDir, snakeCase(res.GoName)+".go")
		if err := writeGeneratedFileJen(path, generateResourceFile(res)); err != nil {
			return fmt.Errorf("generating %s: %w", res.GoName, err)
		}
	}

	return nil
}

func generateFrameworkResources(resources []GenResource, providerDir string, overrides Overrides) error {
	if err := os.MkdirAll(providerDir, 0755); err != nil {
		return fmt.Errorf("creating provider dir: %w", err)
	}

	for _, res := range resources {
		// skip_resources means "the resource file is hand-written" (e.g. custom
		// parameter bridging); it does not imply the data source is hand-written
		// too, so only the resource_*.go write is skipped here.
		skip := false
		for _, s := range overrides.SkipResources {
			if s == res.ShortName || s == res.EndpointBase {
				skip = true
				break
			}
		}
		// A resource with no create/update/delete has nothing for Terraform
		// to manage - it's a pure lookup, so it should only exist as a data
		// source (e.g. template_kinds: a fixed system enum, not even
		// updatable). Generating an empty-shell resource type for it would
		// just be a worse-UX alias for the data source (needs `terraform
		// import`, participates in plan/apply, but can never actually
		// change anything).
		hasAnyMutation := res.HasCreate || res.HasUpdate || res.HasDelete
		if skip {
			if verbose {
				log.Printf("Skipping resource %s (in skip_resources, hand-written)", res.ShortName)
			}
		} else if !hasAnyMutation {
			if verbose {
				log.Printf("Skipping resource %s (no create/update/delete, data-source-only)", res.ShortName)
			}
		} else {
			resPath := filepath.Join(providerDir, "resource_"+snakeCase(res.GoName)+".go")
			if err := writeGeneratedFileJen(resPath, generateFrameworkResourceFile(res)); err != nil {
				return fmt.Errorf("generating framework resource %s: %w", res.GoName, err)
			}
		}
		skipDS := false
		for _, s := range overrides.SkipDataSources {
			if s == res.ShortName || s == res.EndpointBase {
				skipDS = true
				break
			}
		}
		if skipDS {
			if verbose {
				log.Printf("Skipping data source %s (in skip_data_sources, hand-written)", res.ShortName)
			}
		} else if res.HasIndex && res.ParentEndpoint == "" {
			dsPath := filepath.Join(providerDir, "datasource_"+snakeCase(res.GoName)+".go")
			if err := writeGeneratedFileJen(dsPath, generateDataSourceFile(res)); err != nil {
				return fmt.Errorf("generating framework data source %s: %w", res.GoName, err)
			}
		}
	}

	// Generate provider registration file
	if err := generateProviderFile(resources, providerDir); err != nil {
		return fmt.Errorf("generating provider: %w", err)
	}

	return nil
}

func generateTestFiles(resources []GenResource, outputDir, providerDir string, overrides Overrides) error {
	var kept []GenResource
	for _, res := range resources {
		// Check if this resource is in the skip list
		skip := false
		for _, s := range overrides.SkipResources {
			if s == res.ShortName || s == res.EndpointBase {
				skip = true
				break
			}
		}
		if skip {
			if verbose {
				log.Printf("Skipping test for %s (in skip_resources)", res.ShortName)
			}
			continue
		}
		kept = append(kept, res)

		// Round-trip JSON test in generated/
		path := filepath.Join(outputDir, snakeCase(res.GoName)+"_roundtrip_test.go")
		if err := writeGeneratedFileJen(path, generateRoundTripTestFile(res)); err != nil {
			return fmt.Errorf("generating roundtrip test %s: %w", res.GoName, err)
		}
	}

	// The consolidated test files below assume every resource follows the
	// generic Request/Query shape, which skip_resources entries (hand-written
	// client code) don't - so they're built from `kept`, not `resources`.

	// Consolidated fuzz test in generated/
	if err := writeGeneratedFileJen(filepath.Join(outputDir, "fuzz_test.go"), generateFuzzTestFile(kept)); err != nil {
		return fmt.Errorf("generating fuzz test: %w", err)
	}

	// Consolidated status-code test in generated/
	if err := writeGeneratedFileJen(filepath.Join(outputDir, "statuscode_test.go"), generateStatusCodeTestFile(kept)); err != nil {
		return fmt.Errorf("generating status code test: %w", err)
	}

	// Consolidated acceptance test in internal/provider/
	if err := writeGeneratedFileJen(filepath.Join(providerDir, "acceptance_test.go"), generateAcceptanceTestFile(kept)); err != nil {
		return fmt.Errorf("generating acceptance test: %w", err)
	}

	// Golden file test (once, not per resource)
	if err := writeGeneratedFileJen(filepath.Join(outputDir, "generated_code_test.go"), generateGoldenFileTestFile()); err != nil {
		return fmt.Errorf("generating golden file test: %w", err)
	}

	return nil
}

func generateProviderFile(resources []GenResource, providerDir string) error {
	path := filepath.Join(providerDir, "provider.go")
	return writeGeneratedFileJen(path, generateProviderFileJen(resources))
}

// writeGeneratedFileJen renders a jen.File, gofmts it, and writes it to path.
// It skips files that already exist and do not contain a "Code generated" header,
// treating them as hand-written. Files are written atomically via a temporary
// file to avoid partial writes.
func writeGeneratedFileJen(path string, file *jen.File) error {
	// Skip hand-written files (don't have "Code generated" header)
	if existing, err := os.ReadFile(path); err == nil {
		if !strings.Contains(string(existing), "Code generated") {
			log.Printf("Skipping hand-written %s", path)
			return nil
		}
	}

	var buf strings.Builder
	if err := file.Render(&buf); err != nil {
		return fmt.Errorf("rendering jen: %w", err)
	}

	if err := writeFileAtomic(path, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	log.Printf("Generated %s", path)
	return nil
}

// writeFileAtomic writes data to path atomically by first writing to a sibling
// temp file and then renaming it into place. This prevents partially-written
// files if the generator is interrupted.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		cleanup()
		return fmt.Errorf("chmod temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Naming utilities
// ---------------------------------------------------------------------------

// singularizePascal converts an apidoc resource ID (snake_case, plural) to
// the PascalCase singular used in generated type names.
func singularizePascal(s string) string {
	// Irregulars the generic rules below can't produce. Only IDs that
	// actually reach this function (non-skipped apidoc resources) belong
	// here; a wrong guess for a future apidoc addition shows up immediately
	// as an odd generated type name.
	irregular := map[string]string{
		"http_proxies":         "HTTPProxy",
		"media":                "Medium",
		"operatingsystems":     "OperatingSystem",
		"os_default_templates": "DefaultTemplate",
		"ptables":              "PartitionTable",
	}
	if v, ok := irregular[s]; ok {
		return v
	}

	parts := strings.Split(s, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	result := strings.Join(parts, "")

	switch {
	case strings.HasSuffix(result, "ies"):
		return strings.TrimSuffix(result, "ies") + "y"
	case strings.HasSuffix(result, "s") && !strings.HasSuffix(result, "ss"):
		return strings.TrimSuffix(result, "s")
	}
	return result
}

func snakeCase(s string) string {
	if s == "" {
		return ""
	}
	s = strings.TrimPrefix(s, "Foreman")
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' && s[i-1] >= 'a' && s[i-1] <= 'z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func toPascal(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		if len(p) == 2 && p == strings.ToUpper(p) {
			parts[i] = p // keep acronyms like "id", "os", "ip"
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	result := strings.Join(parts, "")
	// Fix specific cases
	result = strings.ReplaceAll(result, "Id", "ID")
	result = strings.ReplaceAll(result, "Ids", "IDs")
	result = strings.ReplaceAll(result, "Url", "URL")
	result = strings.ReplaceAll(result, "Ip", "IP")
	result = strings.ReplaceAll(result, "Uuid", "UUID")
	result = strings.ReplaceAll(result, "Mac", "MAC")
	result = strings.ReplaceAll(result, "Dns", "DNS")
	result = strings.ReplaceAll(result, "Ssh", "SSH")
	result = strings.ReplaceAll(result, "Tls", "TLS")
	result = strings.ReplaceAll(result, "Ssl", "SSL")
	result = strings.ReplaceAll(result, "Http", "HTTP")
	result = strings.ReplaceAll(result, "Pxe", "PXE")
	return result
}

func sortFields(fields []GenField) {
	sort.Slice(fields, func(i, j int) bool {
		if fields[i].Required != fields[j].Required {
			return fields[i].Required
		}
		return fields[i].JSONName < fields[j].JSONName
	})
}

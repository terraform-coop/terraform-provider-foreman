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
	Name          string            `yaml:"name"`
	ShortName     string            `yaml:"short_name"`
	Endpoint      string            `yaml:"endpoint"`
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
	HasCreate      bool
	HasUpdate      bool
	HasDelete      bool
	HasRead        bool
	HasIndex       bool
	Fields         []GenField
	EntityFields   []GenField
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

// fields to exclude from generated types (API context fields)
var taxonomyFields = map[string]bool{
	"location_id": true, "organization_id": true,
	"location_ids": true, "organization_ids": true,
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

		epBase := override.Endpoint
		if epBase == "" {
			epBase = rawID
		}

		typeName := override.Name
		if typeName == "" {
			typeName = singularizePascal(rawID)
		}

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
				{JSONName: "job_category", GoName: "JobCategory", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "provider_type", GoName: "ProviderType", GoType: "string", TFType: "String", TFGoType: "types.String"},
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
			HasCreate:    false, HasRead: true, HasUpdate: true, HasDelete: false, HasIndex: true,
			Fields: []GenField{},
			EntityFields: []GenField{
				{JSONName: "parameter", GoName: "Parameter", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "puppetclass_id", GoName: "PuppetclassID", GoType: "int64", TFType: "Int64", TFGoType: "types.Int64"},
				{JSONName: "override", GoName: "Override", GoType: "bool", TFType: "Bool", TFGoType: "types.Bool"},
				{JSONName: "description", GoName: "Description", GoType: "string", TFType: "String", TFGoType: "types.String"},
				{JSONName: "default_value", GoName: "DefaultValue", GoType: "string", TFType: "String", TFGoType: "types.String"},
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
				{JSONName: "http_method", GoName: "HTTPMethod", GoType: "string", TFType: "String", TFGoType: "types.String", Required: true},
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
		if entityField.GoType == "" || isCompatibleType(entityField.GoType, reqField.GoType) {
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
			sortFields(fields)
			return fields
		}
	}
	return nil
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
		f.GoType = "string"
		f.TFType = "String"
		f.TFGoType = "types.String"
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
			f.TFType = "List"
			f.TFGoType = "types.List"
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
	if ov.Name != "" {
		res.GoName = "Foreman" + ov.Name
	}
	if ov.Endpoint != "" {
		res.EndpointBase = ov.Endpoint
	}
	if ov.ParentEndpoint != "" {
		res.ParentEndpoint = ov.ParentEndpoint
	}

	exclude := make(map[string]bool)
	for _, n := range ov.ExcludeFields {
		exclude[n] = true
	}

	res.Fields = filterFields(res.Fields, exclude, ov.FieldTypes, ov.AttributeRenames)
	res.EntityFields = filterFields(res.EntityFields, exclude, ov.FieldTypes, ov.AttributeRenames)
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
		if skip {
			if verbose {
				log.Printf("Skipping resource %s (in skip_resources, hand-written)", res.ShortName)
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

func singularizePascal(s string) string {
	// Convert snake_case to PascalCase, then singularize
	// Handle special multi-word names first
	renameMap := map[string]string{
		"operatingsystems":       "OperatingSystem",
		"os_default_templates":   "DefaultTemplate",
		"provisioning_templates": "ProvisioningTemplate",
		"smart_proxies":          "SmartProxy",
		"http_proxies":           "HTTPProxy",
		"compute_profiles":       "ComputeProfile",
		"compute_resources":      "ComputeResource",
		"smart_class_parameters": "SmartClassParameter",
		"common_parameters":      "CommonParameter",
		"template_inputs":        "TemplateInput",
		"partition_tables":       "PartitionTable",
		"default_templates":      "DefaultTemplate",
		"discovery_rules":        "DiscoveryRule",
		"job_templates":          "JobTemplate",
		"webhook_templates":      "WebhookTemplate",
		"registration_commands":  "RegistrationCommand",
		"template_kinds":         "TemplateKind",
		"puppetclasses":          "PuppetClass",
		"override_values":        "OverrideValue",
		"config_reports":         "ConfigReport",
		"fact_values":            "FactValue",
		"auth_source_ldaps":      "AuthSourceLDAP",
		"registration_tokens":    "RegistrationToken",
		"host_statuses":          "HostStatus",
		"hosts_bulk_actions":     "HostsBulkAction",
		"table_preferences":      "TablePreference",
		"template_combinations":  "TemplateCombination",
		"mail_notifications":     "MailNotification",
		"report_templates":       "ReportTemplate",
		"external_usergroups":    "ExternalUsergroup",
		"ssh_keys":               "SSHKey",
		"personal_access_tokens": "PersonalAccessToken",
		"smart_proxy_hosts":      "SmartProxyHost",
		"compute_attributes":     "ComputeAttribute",
	}

	if v, ok := renameMap[s]; ok {
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
	return singularize(result)
}

func singularize(s string) string {
	// Special cases
	if s == "Os" {
		return "OperatingSystem"
	}
	if s == "OsDefaultTemplates" {
		return "DefaultTemplate"
	}
	if s == "Settings" {
		return "Setting"
	}
	if strings.HasSuffix(s, "Statuses") {
		return strings.TrimSuffix(s, "es")
	}

	// Use the resource ID convention for known mappings
	singularMap := map[string]string{
		"Architectures":         "Architecture",
		"Domains":               "Domain",
		"Environments":          "Environment",
		"Hostgroups":            "Hostgroup",
		"Hosts":                 "Host",
		"Media":                 "Medium",
		"Models":                "Model",
		"Subnets":               "Subnet",
		"Proxies":               "Proxy",
		"Parameters":            "Parameter",
		"Ptable":                "PartitionTable",
		"Ptables":               "PartitionTable",
		"Usergroups":            "Usergroup",
		"Users":                 "User",
		"Images":                "Image",
		"Templates":             "Template",
		"TemplateInputs":        "TemplateInput",
		"ComputeResources":      "ComputeResource",
		"ComputeProfiles":       "ComputeProfile",
		"HttpProxies":           "HTTPProxy",
		"SmartProxies":          "SmartProxy",
		"Autosign":              "Autosign",
		"CommonParameters":      "CommonParameter",
		"OverrideValues":        "OverrideValue",
		"OsDefaultTemplates":    "DefaultTemplate",
		"ProvisioningTemplates": "ProvisioningTemplate",
		"TemplateKinds":         "TemplateKind",
		"Puppetclasses":         "PuppetClass",
		"SmartClassParameters":  "SmartClassParameter",
		"DiscoveryRules":        "DiscoveryRule",
		"JobTemplates":          "JobTemplate",
		"Webhooks":              "Webhook",
		"WebhookTemplates":      "WebhookTemplate",
		"DefaultTemplates":      "DefaultTemplate",
		"RegistrationCommands":  "RegistrationCommand",
		"Locations":             "Location",
		"Organizations":         "Organization",
		"Realms":                "Realm",
		"Ping":                  "Ping",
		"Plugins":               "Plugin",
		"Bookmarks":             "Bookmark",
		"Roles":                 "Role",
		"Filters":               "Filter",
		"Permissions":           "Permission",
		"Audits":                "Audit",
		"Tasks":                 "Task",
		"Reports":               "Report",
		"ConfigReports":         "ConfigReport",
		"Settings":              "Setting",
		"Facts":                 "Fact",
		"FactValues":            "FactValue",
	}

	if v, ok := singularMap[s]; ok {
		return v
	}

	// General rules
	if strings.HasSuffix(s, "ies") {
		return strings.TrimSuffix(s, "ies") + "y"
	}
	if strings.HasSuffix(s, "sses") {
		return strings.TrimSuffix(s, "ses")
	}
	if strings.HasSuffix(s, "shes") {
		return strings.TrimSuffix(s, "shes")
	}
	if strings.HasSuffix(s, "ches") {
		return strings.TrimSuffix(s, "ches")
	}
	if strings.HasSuffix(s, "xes") {
		return strings.TrimSuffix(s, "xes")
	}
	if strings.HasSuffix(s, "ses") {
		return strings.TrimSuffix(s, "ses")
	}
	if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		return strings.TrimSuffix(s, "s")
	}

	return s
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

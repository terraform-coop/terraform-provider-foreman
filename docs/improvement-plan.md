# Improvement Plan: terraform-provider-foreman

> **Last updated:** 2026-06-29
>
> Status markers: `[DONE]` `[IN PROGRESS]` `[TODO]`

## Overview

This provider was built under significant time constraints. The five-pillar
strategy below addresses the architectural debt and brings the provider up to
current best practices for both Terraform provider development and Foreman API
integration.

---

## Pillar 1 — Auto-generated Foreman API Client Library `[DONE]`

> **Status:** Complete. Generator at `tools/gen/client/main.go` (~800 lines).
> 23 core resources + 6 Katello resources generated. `apidoc/v2.json` pinned
> from Foreman 3.19 CI. Entity fields extracted from show/create examples.
> Framework resource files also generated from same apidoc input.

### Problem

Currently every Foreman API resource has hand-written CRUD code in
`foreman/api/*.go`. ~40 files with near-identical patterns: wrap JSON, call
endpoint, parse response. Adding a new resource requires ~200 lines of
boilerplate. Hand-written structs drift from what Foreman actually returns.

### Foreman's API Description

Foreman uses [apipie-rails](https://github.com/Apipie/apipie-rails) which
exposes a complete machine-readable description of every endpoint at:

- `/apidoc/v2.json` — Apipie JSON format (used by hammer-cli, python-foreman)
- `/apipie.json?type=swagger` — Dynamic Swagger/OpenAPI 2.0 output

### Existing code generators against this format

| Generator | Lang | Approach | Reason not to use |
|---|---|---|---|
| [apipie-bindings](https://github.com/Apipie/apipie-bindings) v0.7.1 | Ruby | Runtime dispatch from JSON | No Go output |
| [foreman_api](https://github.com/theforeman/foreman_api) | Ruby | Static gen via `bin/generate.rb` | Archived since 2015, Ruby-only |
| [python-foreman](https://github.com/david-caro/python-foreman) | Python | Runtime `generate_func()` from JSON | Python-only, dynamic |
| [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) v2.7.1 | Go | Static OpenAPI 3.x codegen | Foreman's Swagger 2.0 export has no response schemas |
| openapi-generator | Java | General OpenAPI codegen | Same limitation |

No existing tool produces static Go client code from Foreman's API docs.

### Critical gap: no formal response schemas

The apipie JSON has **request parameter definitions** (`params`) with full type
info, but **no formal response schema**. The `examples` field records real
HTTP request/response pairs captured during test runs, but response types must
be inferred.

### Strategy

Build a Go code generator (`internal/tools/gen/client/`) invoked via
`go generate` that:

1.  **Fetches** `/apidoc/v2.json` from a running Foreman instance (or a cached
    copy pinned in the repo).

2.  **Parses** into an intermediate representation with three type layers:

    | Layer | Source | Purpose |
    |---|---|---|
    | **Request type** | `params` — structured param definitions | JSON payload for Create/Update |
    | **Response type** | `examples[n].response.body` — recorded JSON blobs | Full server response shape |
    | **Entity type** | Union of response + request-only fields | Canonical struct used by provider CRUD |

    Response schema is **mined from example response bodies**:
    - Collect every `"response.body"` JSON from the resource's `examples`
    - Unmarshal each into `map[string]interface{}`
    - Take the **union of all top-level keys** across all examples
    - Infer Go types: `"string"`→`string`, integer `number`→`int64`,
      decimal `number`→`float64`, `object`→nested struct, `array`→`[]T`
    - Prepend `ForemanObject` base fields (`id`, `name`, `created_at`,
      `updated_at`)

3.  **Generates** three types per resource + CRUD wrapper functions into
    `generated/`:

    ```go
    // generated/foremandomain.go

    // ForemanDomainRequest — payload sent in POST/PUT bodies
    type ForemanDomainRequest struct {
        Name     string  `json:"name"`
        Fullname *string `json:"fullname,omitempty"`
    }

    // ForemanDomainResponse — full server response (includes server-only fields)
    type ForemanDomainResponse struct {
        ForemanObject
        Fullname string `json:"fullname"`
        DnsId    int    `json:"dns_id"`
    }

    // ForemanDomain — entity type used by the provider (merge of request + response)
    type ForemanDomain struct {
        ForemanObject
        Fullname string `json:"fullname"`
        DnsId    int    `json:"dns_id"`
    }

    // ForemanDomainToRequest — converts entity to request payload
    func (e *ForemanDomain) ForemanDomainToRequest() *ForemanDomainRequest { ... }

    // CRUD methods on *Client
    func (c *Client) CreateDomain(ctx context.Context, body *ForemanDomainRequest) (*ForemanDomain, error) { ... }
    func (c *Client) ReadDomain(ctx context.Context, id int64) (*ForemanDomain, error) { ... }
    func (c *Client) UpdateDomain(ctx context.Context, id int64, body *ForemanDomainRequest) (*ForemanDomain, error) { ... }
    func (c *Client) DeleteDomain(ctx context.Context, id int64) error { ... }
    func (c *Client) QueryDomain(ctx context.Context, name string) (*ForemanDomain, error) { ... }
    ```

4.  **Emits httptest-based unit tests** from the `"examples"` field — each
    recorded request/response pair becomes a test case verifying correct URL,
    method, headers, body, and response parsing.

### Target structure

```
generated/
├── client.go                   # Shared HTTP client (hand-written, stable)
├── foreman_object.go           # Base types, task polling (hand-written)
├── domain.go                   # Generated: types + CRUD + Query
├── domain_test.go              # Generated: httptest-based tests
├── host.go                     # Generated
├── host_test.go                # Generated
├── katello_content_view.go     # Generated
├── ...                         # Every Foreman API resource
└── cmd/
    └── apigen/
        └── main.go             # Code generator (hand-written)
```

### Key design decisions

- **Same Go module** — `generated/` lives inside the provider module, not a
  separate module. This avoids version coordination between two modules and
  mirrors what terraform-provider-aws does with its `skaff/` generator.
- **No third-party codegen framework** — the apipie JSON format is custom
  (unique to apipie-rails). A focused `text/template`-based generator is
  ~500 lines and more maintainable than adapting a general-purpose tool.
- **Type inference is lossy** — JSON `number` maps heuristically: integer
  examples → `int64`, decimal examples → `float64`. A post-processing
  override file (`generated/overrides.yaml`) fixes incorrect types and
  excludes fields that should not appear in Terraform state.
- **Naming convention** — CamelCase from the Apipie `"name"` field
  (`Domain` → `ForemanDomain`).
- **Taxonomy** — Organization/Location IDs are appended at the client level,
  not in generated structs (same as today's `WrapJSONWithTaxonomy`).

---

## Pillar 2 — Migrate to Terraform Plugin Framework `[DONE]`

> **Status:** Complete. SDKv2 completely removed. Framework-only provider in
> `internal/provider/`. All 23 core resources generated from apidoc. 6 Katello
> resources hand-written stubs. `main.go` uses `providerserver.Serve` directly
> (no mux needed — SDKv2 is gone). Module: Go 1.25, framework v1.19.0.
> `.golangci.yml` depguard blocks SDKv2 re-introduction.

### Problem

Currently uses `terraform-plugin-sdk/v2` v2.24.0, which lacks:
- Native null/unknown/known tri-state
- Nested attribute types (uses `TypeList`/`TypeSet` blocks)
- Provider-defined types and custom validators
- Protocol version 6 (plan output, resource identity)
- `terraform-plugin-testing` acceptance test framework

### Current layout (to be replaced)

```
foreman/
├── provider.go                 # schema.Provider definition
├── config.go                   # Client factory
├── resource_helper.go          # Deprecated buildForemanObject
├── api/                        # Hand-written API client
├── resource_foreman_*.go       # 32 resource files (SDKv2)
├── data_source_foreman_*.go    # 33 data source files (SDKv2)
├── foreman_api_test.go         # Shared test framework (SDKv2)
├── *_test.go                   # 42 test files
└── testdata/                   # Mock JSON
```

### Target layout (HashiCorp scaffolding-framework convention)

```
.
├── internal/
│   ├── provider/
│   │   ├── provider.go                 # provider.Provider implementation
│   │   ├── provider_test.go
│   │   ├── resource_domain.go          # One file per resource
│   │   ├── resource_domain_test.go
│   │   ├── resource_host.go
│   │   ├── resource_host_test.go
│   │   ├── datasource_domain.go        # One file per data source
│   │   ├── datasource_domain_test.go
│   │   ├── ...                         # Remaining resources/data sources
│   │   ├── models.go                   # Shared model conversion helpers
│   │   └── models_test.go
│   ├── client/                         # Wraps generated/* client
│   │   ├── client.go
│   │   └── client_test.go
│   └── validators/                     # Custom Terraform validators
│       └── validators.go
├── generated/                           # Pillar 1 — auto-generated API client
│   ├── client.go
│   ├── domain.go
│   ├── host.go
│   └── ...
├── templates/                           # Documentation templates (tfplugindocs)
│   ├── index.md.tmpl
│   └── resources/
├── examples/                            # HCL examples
│   └── ...
├── tools/                               # Helper tools
│   └── gen/
│       └── client/
│           └── main.go                  # Code generator
├── .golangci.yml                        # Linter config
├── .pre-commit-config.yaml
├── GNUmakefile
├── main.go
├── go.mod
└── version/VERSION                      # Single version source
```

Rationale for `internal/`: Packages under `internal/` cannot be imported by
external modules. This is the standard pattern across all HashiCorp providers
(AWS, Azure, Random, Scaffolding) and prevents the provider's internals from
becoming a public API contract.

### Target dependency state

```
github.com/hashicorp/terraform-plugin-framework v1.19.0
github.com/hashicorp/terraform-plugin-go          v0.31.0
github.com/hashicorp/terraform-plugin-testing      v1.15.0
github.com/hashicorp/terraform-plugin-mux          v0.23.0
github.com/hashicorp/terraform-plugin-sdk/v2       v2.40.0  (intermediate, removed in Phase 6)
github.com/hashicorp/terraform-plugin-log          v0.9.0
github.com/hashicorp/terraform-plugin-framework-validators  latest
Go 1.25+
```

All `terraform-plugin-*` versions must be upgraded as a family. Using
mismatched versions causes runtime errors with Terraform v1.12+.

### Concrete resource implementation

Every resource follows this pattern (from HashiCorp's scaffolding-framework
and random provider):

```go
// internal/provider/resource_domain.go

package provider

import (
    "context"
    "fmt"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-log/tflog"
)

// Interface compliance (compile-time check)
var _ resource.Resource = &domainResource{}
var _ resource.ResourceWithImportState = &domainResource{}

func NewDomainResource() resource.Resource {
    return &domainResource{}
}

type domainResource struct {
    client *ForemanClient
}

// Model struct — uses tfsdk tags, not json tags
type domainResourceModel struct {
    ID       types.String `tfsdk:"id"`
    Name     types.String `tfsdk:"name"`
    Fullname types.String `tfsdk:"fullname"`
}

func (r *domainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *domainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Version: 1, // Version from day 1 for future state migrations
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "name": schema.StringAttribute{
                Required: true,
            },
            "fullname": schema.StringAttribute{
                Optional: true,
            },
        },
    }
}

func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan domainResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Convert TF model → generated request type
    body := &generated.ForemanDomainRequest{
        Name:     plan.Name.ValueString(),
        Fullname: plan.Fullname.ValueStringPointer(),
    }

    result, err := r.client.CreateDomain(ctx, body)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", err.Error())
        return
    }

    // Convert generated response → TF model
    plan.ID = types.StringValue(fmt.Sprintf("%d", result.Id))
    plan.Name = types.StringValue(result.Name)
    plan.Fullname = types.StringValue(result.Fullname)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read, Update, Delete, ImportState follow the same pattern.
```

### Provider registration

```go
// internal/provider/provider.go

type ForemanProvider struct {
    version string
}

func New(version string) func() provider.Provider {
    return func() provider.Provider {
        return &ForemanProvider{version: version}
    }
}

func (p *ForemanProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "foreman"
    resp.Version = p.version
}

func (p *ForemanProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewDomainResource,
        NewHostResource,
        // ...
    }
}

func (p *ForemanProvider) ConfigureProvider(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    // Read provider config from req.Config, build ForemanClient
}
```

### Version injection

```go
// main.go
var version string = "dev"

func main() {
    opts := providerserver.ServeOpts{
        Address: "registry.terraform.io/terraform-coop/foreman",
    }
    err := providerserver.Serve(context.Background(), provider.New(version), opts)
    // ...
}
```

```yaml
# .goreleaser.yml
ldflags:
  - '-s -w -X main.version={{.Version}} -X main.commit={{.Commit}}'
```

### Migration strategy

**Phase 1 — Skeleton + mux** (single commit)

1.  Create `internal/provider/provider.go` implementing `provider.Provider`.
2.  Use `terraform-plugin-mux` in `main.go` to serve both old SDKv2 provider
    and new framework provider under the same binary.
3.  Migrate provider config (server, credentials, TLS, logging) to the
    framework's `ConfigureProvider` method.
4.  Move all existing SDKv2 resource files into `internal/provider/` so they
    sit alongside future framework resources.

**Phase 2 — Resource-by-resource** (multiple commits)

For each resource:
1.  Create `internal/provider/resource_<name>.go` with the framework pattern
    above.
2.  Define the model struct with `tfsdk` tags and `tfsdk.Schema`.
3.  Implement `Create`, `Read`, `Update`, `Delete`, `ImportState`.
4.  Wire into the mux and remove the SDKv2 version.

Priority order (by usage frequency):
1.  `foreman_domain`, `foreman_subnet`, `foreman_operatingsystem`
2.  `foreman_hostgroup`, `foreman_host`
3.  `foreman_media`, `foreman_model`, `foreman_architecture`
4.  `foreman_computeresource`, `foreman_computeprofile`
5.  `foreman_user`, `foreman_usergroup`
6.  `foreman_parameter`, `foreman_global_parameter`
7.  `foreman_smartproxy`, `foreman_environment`, `foreman_puppetclass`
8.  `foreman_image`, `foreman_provisioningtemplate`
9.  `foreman_partitiontable`, `foreman_defaulttemplate`
10. `foreman_katello_*` (all Katello resources)
11. `foreman_httpproxy`, `foreman_jobtemplate`, `foreman_templateinput`
12. `foreman_webhook`, `foreman_webhooktemplate`
13. `foreman_discovery_rule`

**Phase 3 — Remove SDKv2** (final commit)

1.  Delete all `resource_foreman_*.go` and `data_source_foreman_*.go` files
    from the old SDKv2 pattern.
2.  Remove the mux, remove SDKv2 dependency from `go.mod`.
3.  Add `depguard` linter rule to block any accidental re-introduction of
    `terraform-plugin-sdk/v2`.

### Key mapping

| SDKv2 concept | Framework equivalent |
|---|---|
| `schema.Resource` | `resource.Resource` interface |
| `schema.ResourceData` | Model struct with `tfsdk` tags + req.State/Plan/Config |
| `d.Get("name").(string)` | `plan.Name.ValueString()` |
| `d.SetId("123")` | `plan.ID = types.StringValue("123")` |
| `schema.EnvDefaultFunc` | `os.Getenv()` in `ConfigureProvider` or schema `.Default` |
| `schema.TypeMap` | `schema.MapAttribute` |
| `schema.TypeList` + `Elem` | `schema.ListNestedAttribute` |
| `schema.Importer` | `resource.ResourceWithImportState` |
| `diag.Diagnostics` | `*resp.Diagnostics` |
| `fmt.Sprintf("%s", d.Id())` | `state.ID.ValueString()` |
| `d.HasChange("name")` | Compare old/new in `Update` via state and plan |

---

## Pillar 3 — Integration Tests with CI Workflow `[IN PROGRESS]`

> **Status:** Infrastructure added. `provider_test.go` with test helpers,
> terraform-plugin-testing v1.15.0 added, CI workflow updated with
> terraform version matrix and integration job (Foreman service container).
> Acceptance test bodies (per-resource TestAcc*) still TODO.

### Problem

Current tests are all mock/httptest-based. No tests against a real Foreman
instance. Integration bugs are caught only by end users.

### Strategy

**Unit tests** — continued and expanded from the generated test suite
(Pillar 1), using `httptest` + generated fixtures.

**Acceptance tests** — `terraform-plugin-testing` with `testcontainers-go`
spinning up a real Foreman instance.

### Acceptance test pattern (per resource)

```go
// internal/provider/resource_domain_test.go
package provider

func TestAccDomain_Basic(t *testing.T) {
    resource.ParallelTest(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
        Steps: []resource.TestStep{
            {
                Config: testConfigDomainBasic,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("foreman_domain.test", "name", "example.com"),
                ),
            },
            // Import test
            {
                ResourceName:      "foreman_domain.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

### Test categories

| Category | Tag | Runner | Purpose |
|---|---|---|---|
| Unit | (none) | `go test ./...` | Generated httptest tests, model round-trips, fuzz |
| Acceptance | `integration` | `go test -tags=integration ./...` | Real Foreman via testcontainers |
| Smoke | (none) | `terraform apply` on `examples/` | Documented configs actually work |
| Upgrade | (none) | `go test -tags=integration` | State migration across versions |

### CI workflow (`.github/workflows/test.yml`)

Must match the scaffolding-framework CI structure:

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - run: go build -v .
      - uses: golangci/golangci-lint-action@v6
        with: { version: v2.12.2 }

  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - run: go generate ./...
      - run: git diff --compact-summary --exit-code  # fail if stale generated code

  test:
    strategy:
      matrix:
        terraform: ['1.12.*', '1.13.*', '1.14.*']
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - uses: hashicorp/setup-terraform@v3
        with: { terraform_version: ${{ matrix.terraform }} }
      - run: go test -race -count=1 -coverprofile=coverage.out ./internal/...

  integration:
    runs-on: ubuntu-latest
    services:
      foreman:
        image: theforeman/foreman:3.13
        ports: [443:443]
        env: { FOREMAN_ADMIN_PASSWORD: changeme }
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.25' }
      - run: go test -v -tags=integration -count=1 -race .terraform Version: ${{ matrix.terraform }}
        env:
          FOREMAN_SERVER_HOSTNAME: localhost
          FOREMAN_CLIENT_USERNAME: admin
          FOREMAN_CLIENT_PASSWORD: changeme
          FOREMAN_CLIENT_TLS_INSECURE: true
```

### Local workflow (`GNUmakefile`)

```makefile
default: build

build:
	go build -v .
install:
	go install -v .
fmt:
	gofmt -s -w -e .
lint:
	golangci-lint run
generate:
	go generate ./...
docs:
	cd tools; go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate
test:
	go test -v -cover -timeout=120s -parallel=10 -count=1 ./...
testacc:
	TF_ACC=1 go test -v -cover -timeout=120m -tags=integration -count=1 ./...
testacc-parallel:
	TF_ACC=1 go test -v -cover -timeout=120m -tags=integration -count=1 -parallel=20 ./...
cover:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | grep total | awk '{print $$3}' | \
	  xargs -I{} sh -c 'test "$$(echo "{}" | tr -d %)" -ge 80'
.PHONY: build install fmt lint generate docs test testacc testacc-parallel cover
```

### Makefile targets

| Target | What it does |
|---|---|
| `build` | `go build -v .` — compile check |
| `install` | `go install -v .` — install to `$GOPATH/bin` |
| `fmt` | `gofmt -s -w -e .` — format all Go files |
| `lint` | `golangci-lint run` — run all linters |
| `generate` | `go generate ./...` — regenerate code + docs |
| `docs` | `tfplugindocs generate` — generate Terraform registry docs |
| `test` | Unit tests, parallel=10, race detection, coverage |
| `testacc` | Acceptance tests (sequential) |
| `testacc-parallel` | Acceptance tests (parallel=20) |
| `cover` | Run coverage and enforce >=80% |

---

## Pillar 4 — Defensive Test Architecture `[IN PROGRESS]`

> **Status:** Generated tests added. Round-trip JSON, fuzz, status code,
> golden file, and acceptance test scaffolding all generated per-resource.
> 337 tests passing. Coverage gate (>=70%) in CI.

### Motivation

LLM-generated Go code is statistically likely to produce:
- Off-by-one or wrong field names in struct conversions
- Missing `nil` checks on pointer/interface returns
- Wrong HTTP methods or endpoint paths
- Incorrect JSON tags / omitempty errors
- Shadowed variables and context misuse
- Ignored error returns
- Wrong type assertions

### Strategy

#### 4a. Testing library

Use `github.com/stretchr/testify` v1.11.1 for all assertions:

- `require` — halts test on failure (setup/teardown)
- `assert` — non-fatal assertions (multiple checks per test)
- `suite` — test suites with shared setup/teardown

#### 4b. Round-trip JSON serialization (every struct)

```go
func TestForemanDomain_RoundTrip(t *testing.T) {
    t.Parallel()
    original := &ForemanDomain{
        ForemanObject: ForemanObject{Id: 1, Name: "example.com"},
        Fullname:     "Example Domain",
    }
    data, err := json.Marshal(original)
    require.NoError(t, err)

    var decoded ForemanDomain
    err = json.Unmarshal(data, &decoded)
    require.NoError(t, err)

    assert.Equal(t, original.Id, decoded.Id)
    assert.Equal(t, original.Name, decoded.Name)
    assert.Equal(t, original.Fullname, decoded.Fullname)
}
```

Generated per-resource. Catches missing JSON tags, wrong types, custom
UnmarshalJSON bugs, omitempty misconfiguration.

#### 4c. Fuzz tests on API response parsing

```go
func FuzzParseDomainResponse(f *testing.F) {
    f.Fuzz(func(t *testing.T, data []byte) {
        var d ForemanDomain
        _ = json.Unmarshal(data, &d) // must not panic
    })
}
```

One per generated struct. Catches panics on unexpected server responses.

#### 4d. Golden file compilation guard

```go
//go:generate go run ./tools/gen/client -input apidoc/v2.json -output ../generated/

func TestGeneratedCodeUpToDate(t *testing.T) {
    // Re-generate into temp dir, diff against committed files.
    // Fail on any difference — forces regeneration before commit.
}
```

Catches drift between Apipie spec and generated code.

#### 4e. Exhaustive status code tests

```go
func TestDomain_AllStatusCodes(t *testing.T) {
    t.Parallel()
    for _, tc := range []struct {
        code int
        wantError bool
    }{
        {200, false}, {201, false},
        {400, true}, {401, true}, {403, true},
        {404, true}, {422, true}, {500, true},
    } {
        t.Run(http.StatusText(tc.code), func(t *testing.T) {
            srv := httptest.NewServer(...)
            ...
        })
    }
}
```

Catches missing 404 handling for delete, wrong error wrapping, missing
async task polling paths.

#### 4f. Table-driven CRUD tests (preserve existing framework)

The existing 5-function framework (URL/method, empty request, request data,
status code, mock response) is strong. Keep it, generated per-resource, and
add:

- **Null/zero-value tests**: every optional field set to Go zero value
- **Boundary tests**: max-length strings, negative IDs, special chars
- **Concurrency tests**: `go test -race` for any shared client state

#### 4g. Race detection on every CI run

```yaml
- run: go test -race -count=1 ./...
```

Every CI job runs with `-race`. Catches concurrent map writes from shared
`Client` or logger state.

#### 4h. Coverage gates

Enforce >=80% on `generated/` client code, >=70% on `internal/provider/`.

```yaml
- run: go test -coverprofile=coverage.out -covermode=atomic ./...
- run: go tool cover -func=coverage.out | grep total | awk '{print $3}' | \
       xargs -I{} sh -c 'test "$(echo "{}" | tr -d %)" -ge 80'
```

#### 4i. Generated test template checklist

Every code generator template must emit tests for:

| Test | Failure mode caught |
|---|---|
| Round-trip JSON | Missing `json` tags, wrong types, omitempty |
| Create HTTP method + path | Wrong endpoint or verb |
| Read HTTP method + path | Wrong endpoint or verb |
| Update HTTP method + path | Wrong endpoint or verb |
| Delete HTTP method + path | Wrong endpoint or verb |
| Query HTTP method + path | Wrong endpoint or params |
| Request body for Create | Wrong JSON key names |
| Request body for Update | Wrong JSON key names |
| Empty body for Read/Delete | Sending body when should not |
| 404 on Read → error | Missing error check |
| 404 on Delete → success | `CheckDeleted` not called |
| Non-2xx → typed error | Wrong error wrapping |
| Empty 200 response | Nil-check on response body |
| Full mock response | State finalization wrong |
| Zero-value optional fields | Nil vs empty slice confusion |
| Fuzz parse | Panic on bad data |
| `t.Parallel()` not set | Forgot parallel marker |

---

## Pillar 5 — pre-commit Hooks `[DONE]`

> **Status:** Complete. `.pre-commit-config.yaml` (TekWizely),
> `.golangci.yml` (depguard blocks SDKv2). CI workflows updated:
> `test.yml` (build/lint/test/generate jobs, Go 1.25),
> `build.yml` (goreleaser v4).

### Hook pipeline

`.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
        args: ['--maxkb=500']

  - repo: https://github.com/TekWizely/pre-commit-golang
    rev: v1.0.0-rc.4
    hooks:
      - id: go-fmt-repo
      - id: go-vet-repo-mod
      - id: go-mod-tidy-repo
      - id: go-test-repo-mod
        args: ['-race', '-count=1', '-timeout=300s', './...']
      - id: go-build-repo-mod
```

TekWizely is the active fork (dnephin/pre-commit-golang archived 2025-03).

### .golangci.yml

```yaml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - staticcheck
    - depguard
    - errcheck
    - misspell
  presets:
    - bugs
    - comment
    - error
    - format
    - import
    - metalinter
    - module
    - performance
    - sql
    - test
    - unused

linters-settings:
  depguard:
    rules:
      no-sdkv2:
        deny:
          - pkg: github.com/hashicorp/terraform-plugin-sdk/v2
            desc: "Use terraform-plugin-framework or terraform-plugin-testing instead"

issues:
  exclude-rules:
    # Generated files: skip vet warnings, still enforce fmt
    - path: generated/
      linters:
        - govet
        - errcheck
```

### CI lint job

```yaml
lint:
  runs-on: ubuntu-latest
  steps:
    - uses: pre-commit/action@v3.0.1
```

Duplicates the pre-commit checks in CI so the gate works even if someone
skips the local hook.

### Generated-code policy

Files under `generated/` carry a `// Code generated ... DO NOT EDIT.` header.
`go-fmt-repo` still formats them (readable generated code is easier to
review). `go-vet-repo-mod` skips them via `issues.exclude-rules` in
`.golangci.yml` (noise from generated code is not useful).

---

## Dependencies & Versions

| Dependency | Version | Status |
|---|---|---|
| Go | 1.25 | Required by plugin-framework v1.19.0 |
| terraform-plugin-framework | v1.19.0 | Stable 2026-03-10 |
| terraform-plugin-go | v0.31.0 | Pair with framework v1.19.0 |
| terraform-plugin-mux | v0.23.0 | Pair with framework v1.19.0 |
| terraform-plugin-testing | v1.15.0 | Pair with framework v1.19.0 |
| terraform-plugin-sdk/v2 | v2.40.0 | Deprecated, removed in Phase 6 |
| terraform-plugin-log | v0.9.0 | Structured logging |
| terraform-plugin-framework-validators | latest | Common validators |
| terraform-plugin-docs | latest | `tfplugindocs` CLI |
| testcontainers-go | v0.43.0 | Integration test containers |
| testify | v1.11.1 | Assertions |
| golangci-lint | v2.12.2 | Linter aggregator |
| oapi-codegen | v2.7.1 | Future fallback (OpenAPI 3.x) |
| pre-commit | v4.2.0+ | Python hook framework |
| pre-commit-hooks | v5.0.0 | Basic file hooks |
| pre-commit-golang | v1.0.0-rc.4 | TekWizely (active fork) |

### Terraform plugin family compatibility

| framework | plugin-go | mux | sdk/v2 | testing |
|---|---|---|---|---|
| v1.19.0 | v0.31.0 | v0.23.0 | v2.40.0 | v1.15.0 |
| v1.15.0 | v0.28.0 | v0.20.0 | v2.37.0 | v1.13.1 |

Always upgrade all five together. Mixing rows causes runtime errors with
Terraform v1.12+ (resource identity feature negotiation).

### Existing Go Foreman clients (not suitable)

| Library | Stars | Status | Missing |
|---|---|---|---|
| `rdeusser/go-foreman` | 2 | Abandoned 2017 | Katello, async tasks, taxonomy |
| `bshuster-repo/foreman-go` | 0 | Alpha, abandoned 2017 | Everything |

No existing library supports Foreman 3.x, Katello, or async task polling.

---

## Migration Order

```
Phase 0  — [DONE] Build code generator in tools/gen/client/. Generate all client
           code into generated/. Add overrides.yaml for type corrections.
           Result: no functional change, all SDKv2 resources still work.

Phase 1  — [DONE] Update Go to 1.25, bump all deps. Create internal/provider/
           with framework skeleton. Wire mux in main.go. Move existing
           SDKv2 files into internal/provider/. Add Makefile, golangci.yml.

Phase 2  — [DONE] Add acceptance test infrastructure: terraform-plugin-testing,
           test helpers in provider_test.go, CI workflow with terraform matrix
           and integration job (Foreman service container).

Phase 3  — [DONE] Migrate resources one-by-one to framework. All 23 core
           resources generated from apidoc. 6 Katello hand-written stubs.
           SDKv2 code fully removed. No mux needed (framework-only).

Phase 4  — [IN PROGRESS] Add defensive tests. Round-trip JSON, fuzz, status code,
           golden file, acceptance test scaffolding all generated per-resource.
           337 tests passing. Coverage gate in CI.

Phase 5  — [DONE] Add pre-commit config, .golangci.yml, CI lint job.

Phase 6  — [DONE] Remove SDKv2 code, remove mux, remove sdk/v2 dep.

All work on feat/rewrite-provider. Never push until Phase 6 is green.
```

## Commit Strategy

```bash
git checkout -b feat/rewrite-provider

# Phase 0 [DONE]
git commit -m "feat(api): add code generator and generated client code"

# Phase 1 [DONE]
git commit -m "feat(provider): update Go to 1.25, add framework skeleton, mux, Makefile"

# Phase 2 [TODO]
git commit -m "test: add acceptance test infrastructure with testcontainers"

# Phase 3 [DONE] — single large commit instead of per-resource
git commit -m "feat: complete provider rewrite with auto-generated client and framework resources"

# Phase 4 [TODO]
git commit -m "test: add round-trip, fuzz, status code, coverage gate tests"

# Phase 5 [DONE]
git commit -m "ci: add pre-commit hooks, golangci.yml lint job"

# Phase 6 [DONE]
git commit -m "refactor: remove SDKv2 code and mux, publish v2"
```

Every commit must pass `go build && go test -race -count=1 ./...` and all
pre-commit hooks. Acceptance tests skip with `-tags=integration`.

---

## Remaining Work Items

### Pillar 3 — Integration Tests `[IN PROGRESS]`

- [x] Add terraform-plugin-testing dependency
- [x] Create acceptance test helpers in `internal/provider/provider_test.go`
- [x] Add terraform version matrix to CI test job
- [x] Add integration CI job with Foreman service container
- [ ] Write `TestAcc*_Basic` tests for each resource (create, read, update, delete, import)

### Pillar 4 — Defensive Tests `[IN PROGRESS]`

- [x] Generate round-trip JSON tests for every generated struct
- [x] Generate fuzz tests on API response parsing (one per struct)
- [x] Add golden file compilation guard (`TestGeneratedCodeUpToDate`)
- [x] Add exhaustive status code tests (200, 201, 400, 401, 403, 404, 422, 500)
- [x] Generate acceptance test scaffolding per resource (TestAcc* with import)
- [x] Enforce coverage gates: >=70% in CI
- [ ] Add null/zero-value tests per field
- [ ] Add boundary tests (max-length strings, negative IDs, special chars)

### Pillar 7 — Field Parity with Old Provider `[TODO]`

> **Status:** All fields identified. TODO markers added to every resource file.
> Implementation not yet started.

**RULE:** No field may be excluded from the Terraform schema with the reason
"manage in UI" or "not manageable". Every field the old provider supported MUST
be present in the new provider. Fields excluded from generation must have
bridging code in hand-written resource files.

#### Host resource — CRITICAL (currently broken)

The host resource currently only exposes computed/read-only fields. Users
cannot set ANY host attributes. This is a regression from the old provider.

**Missing fields requiring bridging code:**

| Field | Type | Status |
|---|---|---|
| `name` | string | TODO(bridget) — Required, hostname |
| `architecture_id` | int64 | TODO(bridget) — Optional |
| `operatingsystem_id` | int64 | TODO(bridget) — Optional |
| `domain_id` | int64 | TODO(bridget) — Optional |
| `environment_id` | int64 | TODO(bridget) — Optional |
| `hostgroup_id` | int64 | TODO(bridget) — Optional |
| `compute_resource_id` | int64 | TODO(bridget) — Optional |
| `compute_profile_id` | int64 | TODO(bridget) — Optional |
| `medium_id` | int64 | TODO(bridget) — Optional |
| `subnet_id` | int64 | TODO(bridget) — Optional |
| `subnet6_id` | int64 | TODO(bridget) — Optional |
| `ptable_id` | int64 | TODO(bridget) — Optional |
| `realm_id` | int64 | TODO(bridget) — Optional |
| `puppet_proxy_id` | int64 | TODO(bridget) — Optional |
| `puppet_ca_proxy_id` | int64 | TODO(bridget) — Optional |
| `owner_id` | int64 | TODO(bridget) — Optional |
| `owner_type` | string | TODO(bridget) — Optional |
| `image_id` | int64 | TODO(bridget) — Optional |
| `model_id` | int64 | TODO(bridget) — Optional |
| `build` | bool | TODO(bridget) — Optional |
| `enabled` | bool | TODO(bridget) — Optional |
| `managed` | bool | TODO(bridget) — Optional |
| `root_pass` | string | TODO(bridget) — Optional+Sensitive |
| `provision_method` | string | TODO(bridget) — Optional |
| `compute_attributes` | string | TODO(bridget) — Optional (JSON) |
| `interfaces_attributes` | []Interface | TODO(bridget) — Complex nested type |
| `host_parameters_attributes` | []Parameter | TODO(bridget) — Complex nested type |

**Excluded from generation (legitimate):**

| Field | Reason |
|---|---|
| `*_name` fields | Read-only display names (use `_id` fields) |
| Nested object duplications | Use `_id` fields instead |
| `token`, `certificate_name`, `capabilities`, `puppet_status` | Read-only computed |
| `all_parameters` | Read-only aggregation |
| `template_combinations` | Managed via other fields |
| Taxonomy fields | Handled at provider level |

#### Hostgroup resource

**Missing fields requiring bridging code:**

| Field | Type | Status |
|---|---|---|
| `compute_resource_id` | int64 | TODO(bridget) — Optional |
| `puppet_proxy_id` | int64 | TODO(bridget) — Optional |
| `puppet_ca_proxy_id` | int64 | TODO(bridget) — Optional |
| `root_pass` | string | TODO(bridget) — Optional+Sensitive |

#### Operating System resource

**Missing fields requiring bridging code:**

| Field | Type | Status |
|---|---|---|
| `os_parameters_attributes` | []Parameter | TODO(bridget) — Array of {name,value,hidden_value} |

#### Architecture resource

**Missing fields requiring bridging code:**

| Field | Type | Status |
|---|---|---|
| `ptables` | []PartitionTable | TODO(bridget) — Managed via operatingsystem.ptable_ids |

#### Katello Repository resource

**Missing fields requiring bridging code:**

| Field | Type | Status |
|---|---|---|
| `ignore_global_proxy` | bool | TODO(bridget) |
| `ignorable_content` | string | TODO(bridget) |
| `verify_ssl_on_sync` | bool | TODO(bridget) |
| `upstream_username` | string | TODO(bridget) |
| `upstream_password` | string | TODO(bridget) — Sensitive |
| `deb_releases` | string | TODO(bridget) |
| `deb_components` | string | TODO(bridget) |
| `deb_architectures` | string | TODO(bridget) |
| `docker_upstream_name` | string | TODO(bridget) |
| `docker_tags_whitelist` | string | TODO(bridget) |
| `ansible_collection_requirements` | string | TODO(bridget) |

#### Katello Lifecycle Environment resource

**Missing fields requiring bridging code:**

| Field | Type | Status |
|---|---|---|
| `prior` | struct | TODO(bridget) — Computed: true |
| `successor` | struct | TODO(bridget) — Computed: true |

#### Katello Content View resource

**Missing fields requiring bridging code:**

| Field | Type | Status |
|---|---|---|
| `repository_ids` | []int | TODO(bridget) — Request + Response |
| `component_ids` | []int | TODO(bridget) — Request + Response |

### Other `[IN PROGRESS]`

- [x] Data sources for all resources (37 total: 28 generated + 6 Katello + 3 plugin)
- [x] Added 5 hardcoded resources not in apidoc (environment, jobtemplate, puppetclass, smartclassparameter, templatekind)
- [x] Plugin resources (discovery_rule, webhook, webhook_template, override_value)
- [x] Katello framework resources (6 hand-written)
- [x] 404 handling (Read→remove from state, Delete→ignore)
- [x] Taxonomy wrapping (org/loc IDs added to request body)
- [x] Wrapper key correctness (resource key in request body)
- [x] Client regression tests (taxonomy, wrapper key, 404 handling)
- [x] Hostgroup parameters (dual-format array/map, hand-written resource)
- [x] Generator: skip_resources option, hand-written file protection
- [x] json.RawMessage support for nested types (array of objects)
- [x] parent_endpoint support for override_value
- [x] Conditional encoding/json import
- [x] Data sources only when HasIndex=true
- [x] TODO markers added to ALL excluded fields requiring bridging code
- [x] Host resource: add all missing _id fields (CRITICAL — currently broken)
- [x] Host resource: add interfaces_attributes bridging code
- [x] Host resource: add host_parameters_attributes bridging code
- [x] Host resource: add compute_attributes bridging code
- [x] Hostgroup resource: add compute_resource_id, puppet_proxy_id, puppet_ca_proxy_id, root_pass
- [x] Operating System resource: add os_parameters_attributes bridging code
- [x] Katello Repository: add 11 missing fields
- [x] Katello Lifecycle Environment: add prior_id/successor_id
- [x] Katello Content View: add repository_ids, component_ids
- [x] Content View filter rules — implemented with nested schema, sync on CRUD
- [ ] HCL examples, tfplugindocs

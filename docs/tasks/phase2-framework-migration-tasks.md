# Phase 2: Terraform Plugin Framework Migration - Detailed Tasks

## Overview
Migrate from terraform-plugin-sdk/v2 to modern terraform-plugin-framework using incremental approach with muxing.

## Prerequisites
- [ ] Go 1.22+ installed
- [ ] Phase 1 completed (or running in parallel)
- [ ] All existing tests passing
- [ ] Feature branch created: `feature/framework-migration`

## Task Breakdown

### Task 2.1: Foundation Setup (Week 5, Days 1-3)

**Goal**: Add framework dependencies and set up muxing infrastructure

#### Subtasks:

1. [ ] Update Go version
   ```bash
   # Update go.mod
   go mod edit -go=1.22
   ```

2. [ ] Add framework dependencies
   ```bash
   go get github.com/hashicorp/terraform-plugin-framework@latest
   go get github.com/hashicorp/terraform-plugin-mux@latest
   go get github.com/hashicorp/terraform-plugin-testing@latest
   go mod tidy
   ```

3. [ ] Create new provider structure
   ```
   foreman/
   ├── api/                    # Existing API client
   ├── framework/              # New framework-based code
   │   ├── provider/
   │   │   └── provider.go    # Framework provider implementation
   │   ├── resources/
   │   │   └── architecture.go # Example migrated resource
   │   ├── datasources/
   │   └── internal/
   │       └── validators/     # Custom validators
   ├── provider.go             # Existing SDK v2 provider
   └── resource_*.go           # Existing SDK v2 resources
   ```

4. [ ] Implement framework provider
   ```go
   // foreman/framework/provider/provider.go
   package provider
   
   import (
       "context"
       
       "github.com/hashicorp/terraform-plugin-framework/datasource"
       "github.com/hashicorp/terraform-plugin-framework/provider"
       "github.com/hashicorp/terraform-plugin-framework/provider/schema"
       "github.com/hashicorp/terraform-plugin-framework/resource"
   )
   
   var _ provider.Provider = &ForemanProvider{}
   
   type ForemanProvider struct {
       version string
   }
   
   func New(version string) func() provider.Provider {
       return func() provider.Provider {
           return &ForemanProvider{
               version: version,
           }
       }
   }
   
   func (p *ForemanProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
       resp.TypeName = "foreman"
       resp.Version = p.version
   }
   
   func (p *ForemanProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
       resp.Schema = schema.Schema{
           Description: "Interact with Foreman API",
           Attributes: map[string]schema.Attribute{
               "server_hostname": schema.StringAttribute{
                   Description: "Hostname of the Foreman server",
                   Required:    true,
               },
               "client_username": schema.StringAttribute{
                   Description: "Username for API authentication",
                   Optional:    true,
               },
               "client_password": schema.StringAttribute{
                   Description: "Password for API authentication",
                   Optional:    true,
                   Sensitive:   true,
               },
               // Add other configuration attributes
           },
       }
   }
   
   func (p *ForemanProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
       // Configuration logic
   }
   
   func (p *ForemanProvider) Resources(ctx context.Context) []func() resource.Resource {
       return []func() resource.Resource{
           // Will add resources as we migrate them
       }
   }
   
   func (p *ForemanProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
       return []func() datasource.DataSource{
           // Will add data sources as we migrate them
       }
   }
   ```

5. [ ] Set up muxed provider in main.go
   ```go
   // main.go
   package main
   
   import (
       "context"
       "flag"
       "log"
       
       "github.com/hashicorp/terraform-plugin-framework/providerserver"
       "github.com/hashicorp/terraform-plugin-go/tfprotov6"
       "github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
       "github.com/hashicorp/terraform-plugin-mux/tf5to6server"
       "github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
       
       "github.com/terraform-coop/terraform-provider-foreman/foreman"
       "github.com/terraform-coop/terraform-provider-foreman/foreman/framework/provider"
   )
   
   var version = "dev"
   
   func main() {
       var debug bool
       flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
       flag.Parse()
       
       ctx := context.Background()
       
       // Upgrade SDK v2 provider to protocol v6
       upgradedSdkProvider, err := tf5to6server.UpgradeServer(
           ctx,
           foreman.Provider().GRPCProvider,
       )
       if err != nil {
           log.Fatal(err)
       }
       
       // Create framework provider
       frameworkProvider := providerserver.NewProtocol6(provider.New(version)())
       
       // Mux the providers
       muxServer, err := tf6muxserver.NewMuxServer(ctx, func() tfprotov6.ProviderServer {
           return upgradedSdkProvider
       }, func() tfprotov6.ProviderServer {
           return frameworkProvider
       })
       if err != nil {
           log.Fatal(err)
       }
       
       var serveOpts []tf6server.ServeOpt
       if debug {
           serveOpts = append(serveOpts, tf6server.WithManagedDebug())
       }
       
       err = tf6server.Serve(
           "registry.terraform.io/terraform-coop/foreman",
           muxServer.ProviderServer,
           serveOpts...,
       )
       if err != nil {
           log.Fatal(err)
       }
   }
   ```

6. [ ] Test muxed setup
   ```bash
   go build
   # Test that provider builds successfully
   ```

**Deliverables**:
- Updated dependencies
- Framework provider skeleton
- Muxed main.go
- Successful build

**Success Criteria**:
- Code compiles without errors
- Provider can be instantiated
- Ready for resource migration

---

### Task 2.2: Provider Configuration Migration (Week 5, Days 4-5)

**Goal**: Migrate provider configuration to framework

#### Subtasks:

1. [ ] Define framework provider schema (expand from skeleton)
2. [ ] Implement provider data structure
3. [ ] Add configuration validation
4. [ ] Migrate client initialization logic
5. [ ] Test provider configuration

**Success Criteria**:
- Provider configuration works
- Client created successfully
- Environment variables supported

---

### Task 2.3: Resource Migration - Simple Resources (Week 6-7)

**Goal**: Migrate 10-15 simple CRUD resources

#### Migration Pattern:

For each resource, follow this pattern:

1. [ ] Create new framework resource file
   ```go
   // foreman/framework/resources/architecture_resource.go
   package resources
   
   import (
       "context"
       
       "github.com/hashicorp/terraform-plugin-framework/resource"
       "github.com/hashicorp/terraform-plugin-framework/resource/schema"
       "github.com/hashicorp/terraform-plugin-framework/types"
   )
   
   var _ resource.Resource = &ArchitectureResource{}
   var _ resource.ResourceWithImportState = &ArchitectureResource{}
   
   type ArchitectureResource struct {
       client *adapter.ForemanClient
   }
   
   type ArchitectureResourceModel struct {
       ID   types.String `tfsdk:"id"`
       Name types.String `tfsdk:"name"`
   }
   
   func NewArchitectureResource() resource.Resource {
       return &ArchitectureResource{}
   }
   
   func (r *ArchitectureResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
       resp.TypeName = req.ProviderTypeName + "_architecture"
   }
   
   func (r *ArchitectureResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
       resp.Schema = schema.Schema{
           Description: "Manages Foreman architectures",
           Attributes: map[string]schema.Attribute{
               "id": schema.StringAttribute{
                   Computed:    true,
                   Description: "Unique identifier for the architecture",
               },
               "name": schema.StringAttribute{
                   Required:    true,
                   Description: "Name of the architecture",
               },
           },
       }
   }
   
   func (r *ArchitectureResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
       if req.ProviderData == nil {
           return
       }
       
       client, ok := req.ProviderData.(*adapter.ForemanClient)
       if !ok {
           resp.Diagnostics.AddError(
               "Unexpected Resource Configure Type",
               "Expected *adapter.ForemanClient",
           )
           return
       }
       
       r.client = client
   }
   
   func (r *ArchitectureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
       var data ArchitectureResourceModel
       resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }
       
       // Create architecture via API
       arch := &adapter.ForemanArchitecture{
           Name: data.Name.ValueString(),
       }
       
       created, err := r.client.CreateArchitecture(ctx, arch)
       if err != nil {
           resp.Diagnostics.AddError("Client Error", err.Error())
           return
       }
       
       data.ID = types.StringValue(strconv.Itoa(created.ID))
       
       resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
   }
   
   // Implement Read, Update, Delete, ImportState methods...
   ```

2. [ ] Register resource in provider
3. [ ] Migrate resource tests
4. [ ] Verify CRUD operations work
5. [ ] Mark SDK resource as deprecated (but keep functional)

#### Resources to Migrate (Priority Order):

**Week 6** (5 resources):
- [ ] Architecture
- [ ] Model
- [ ] CommonParameter
- [ ] TemplateKind
- [ ] OperatingSystem (no complex relationships)

**Week 7** (5 resources):
- [ ] Domain
- [ ] Environment
- [ ] Subnet
- [ ] PartitionTable
- [ ] Media

**Success Criteria**:
- 10 resources migrated to framework
- All tests passing
- No breaking changes for users

---

### Task 2.4: Resource Migration - Medium Complexity (Week 8-10)

**Goal**: Migrate resources with relationships and nested attributes

#### Resources to Migrate:

**Week 8** (4-5 resources):
- [ ] HostGroup
- [ ] ComputeProfile
- [ ] ComputeResource
- [ ] Image

**Week 9** (4-5 resources):
- [ ] SmartProxy
- [ ] ProvisioningTemplate
- [ ] JobTemplate
- [ ] Parameter (with associations)

**Week 10** (4-5 resources):
- [ ] Host (complex, many relationships)
- [ ] ComputeAttribute
- [ ] User
- [ ] UserGroup

---

### Task 2.5: Katello Resources Migration (Week 11-12)

**Goal**: Migrate Katello plugin resources

#### Resources to Migrate:

**Week 11**:
- [ ] Product
- [ ] Repository
- [ ] ContentCredential
- [ ] ContentView
- [ ] ContentViewVersion

**Week 12**:
- [ ] ContentViewFilter
- [ ] ContentViewFilterRule
- [ ] SyncPlan
- [ ] ActivationKey
- [ ] HostCollection

---

### Task 2.6: Data Sources Migration (Week 13-14)

**Goal**: Migrate all data sources to framework

#### Pattern for Data Sources:

```go
// foreman/framework/datasources/architecture_data_source.go
package datasources

import (
    "context"
    
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ArchitectureDataSource{}

type ArchitectureDataSource struct {
    client *adapter.ForemanClient
}

type ArchitectureDataSourceModel struct {
    ID   types.String `tfsdk:"id"`
    Name types.String `tfsdk:"name"`
}

func NewArchitectureDataSource() datasource.DataSource {
    return &ArchitectureDataSource{}
}

// Implement Metadata, Schema, Configure, Read methods...
```

#### Data Sources to Migrate:
- [ ] All data sources corresponding to resources
- [ ] Query-based data sources

---

### Task 2.7: Remove SDK v2 (Week 15)

**Goal**: Complete migration by removing SDK v2 code

#### Subtasks:

1. [ ] Verify all resources migrated
2. [ ] Verify all data sources migrated
3. [ ] Remove mux from main.go
4. [ ] Remove SDK v2 dependency
5. [ ] Delete old provider.go and resource files
6. [ ] Update all imports
7. [ ] Run full test suite
8. [ ] Update documentation

**Success Criteria**:
- No SDK v2 code remaining
- All tests passing
- Documentation updated
- Clean build

---

## Testing Strategy

### For Each Migrated Resource:

1. **Unit Tests**
   ```go
   func TestArchitectureResource_Schema(t *testing.T) {
       // Test schema definition
   }
   ```

2. **Acceptance Tests**
   ```go
   func TestAccArchitectureResource(t *testing.T) {
       resource.Test(t, resource.TestCase{
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               {
                   Config: testAccArchitectureResourceConfig("test-arch"),
                   Check: resource.ComposeAggregateTestCheckFunc(
                       resource.TestCheckResourceAttr("foreman_architecture.test", "name", "test-arch"),
                   ),
               },
               // Test update
               {
                   Config: testAccArchitectureResourceConfig("test-arch-updated"),
                   Check: resource.ComposeAggregateTestCheckFunc(
                       resource.TestCheckResourceAttr("foreman_architecture.test", "name", "test-arch-updated"),
                   ),
               },
               // Test import
               {
                   ResourceName:      "foreman_architecture.test",
                   ImportState:       true,
                   ImportStateVerify: true,
               },
           },
       })
   }
   ```

3. **Integration Tests** (with Phase 3 E2E infrastructure)

---

## Definition of Done

Phase 2 is complete when:

- [ ] All resources migrated to framework
- [ ] All data sources migrated to framework
- [ ] SDK v2 code removed
- [ ] All tests passing
- [ ] Documentation updated
- [ ] User configurations still work without changes
- [ ] Code reviewed and approved

## Rollback Plan

If issues arise:
1. Revert to muxed setup
2. Keep both implementations
3. Debug and fix issues
4. Retry removal of SDK v2

## Tools and References

- [Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- [Migration Guide](https://developer.hashicorp.com/terraform/plugin/framework/migrating)
- [Plugin Mux Documentation](https://github.com/hashicorp/terraform-plugin-mux)
- [Testing Documentation](https://developer.hashicorp.com/terraform/plugin/framework/acctests)

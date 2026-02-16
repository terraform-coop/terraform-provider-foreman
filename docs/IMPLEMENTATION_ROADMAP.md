# Implementation Roadmap & GitHub Issues

This document provides GitHub-ready issue templates for tracking the modernization effort.

## Epic Issues

### Epic 1: API Client Modernization

```markdown
**Title**: Modernize API Client with Generated Code

**Description**:
Replace custom HTTP API client with automatically generated, type-safe client based on Foreman's OpenAPI specifications.

**Goals**:
- Reduce maintenance burden of API client code
- Improve type safety and reduce bugs
- Enable faster updates when Foreman API changes
- Provide foundation for future enhancements

**Approach**:
1. Extract Foreman API specifications from Apipie format
2. Convert Apipie to OpenAPI 3.0 specifications
3. Generate Go client code using oapi-codegen
4. Create adapter layer for backward compatibility
5. Incrementally migrate resources to use generated client

**Dependencies**: None

**Estimated Effort**: 4 weeks

**Related Issues**: #TBD (sub-tasks below)

**Documentation**: See `MODERNIZATION_PLAN.md` and `docs/tasks/phase1-api-client-tasks.md`
```

---

### Epic 2: Terraform Plugin Framework Migration

```markdown
**Title**: Migrate to Terraform Plugin Framework

**Description**:
Migrate from terraform-plugin-sdk/v2 to modern terraform-plugin-framework to benefit from improved abstractions, better maintainability, and ongoing HashiCorp support.

**Goals**:
- Use modern, actively maintained framework
- Improve type safety and error handling
- Better developer experience
- Access to new Terraform features

**Approach**:
1. Set up framework dependencies and muxing infrastructure
2. Migrate provider configuration
3. Incrementally migrate resources (using mux to run both)
4. Migrate data sources
5. Remove SDK v2 code

**Dependencies**: None (can run in parallel with Epic 1)

**Estimated Effort**: 11 weeks

**Related Issues**: #TBD (sub-tasks below)

**Documentation**: See `MODERNIZATION_PLAN.md` and `docs/tasks/phase2-framework-migration-tasks.md`
```

---

### Epic 3: E2E Testing Infrastructure

```markdown
**Title**: Implement E2E Testing with Containerized Foreman

**Description**:
Create comprehensive end-to-end testing infrastructure using containerized Foreman instances to validate provider against real API.

**Goals**:
- Test against real Foreman instances
- Catch integration issues early
- Validate API compatibility
- Improve confidence in releases

**Approach**:
1. Set up Docker Compose for Foreman test environment
2. Create test framework and helpers
3. Write E2E tests for all resources
4. Integrate into CI/CD pipeline
5. Support testing against multiple Foreman versions

**Dependencies**: None (can start immediately, benefits from Epic 1 & 2)

**Estimated Effort**: 5 weeks

**Related Issues**: #TBD (sub-tasks below)

**Documentation**: See `MODERNIZATION_PLAN.md` and `docs/tasks/phase3-e2e-testing-tasks.md`
```

---

## Phase 1: API Client - Detailed Issues

### Issue 1.1: Extract Foreman API Specifications

```markdown
**Title**: Extract Foreman API Specifications from Apipie

**Labels**: api-client, phase1, enhancement

**Description**:
Extract API documentation from Foreman's Apipie endpoints in JSON format for conversion to OpenAPI.

**Tasks**:
- [ ] Set up Foreman test instance (Docker or existing)
- [ ] Extract Foreman core API spec from `/apidoc/api.json`
- [ ] Extract Katello API spec from `/katello/apidoc/api.json`
- [ ] Extract Foreman Tasks API spec from `/foreman_tasks/apidoc/api.json`
- [ ] Extract Puppet Plugin API spec (if available)
- [ ] Document API structure and endpoint coverage
- [ ] Store specs in `api-specs/` directory

**Acceptance Criteria**:
- [ ] All API specs extracted successfully
- [ ] Documentation includes endpoint count and resource types
- [ ] Specs checked into version control

**Estimated Effort**: 2 days

**Related to**: Epic #TBD (API Client Modernization)
```

---

### Issue 1.2: Create Apipie to OpenAPI Converter

```markdown
**Title**: Build Apipie to OpenAPI Conversion Tool

**Labels**: api-client, phase1, tooling, enhancement

**Description**:
Create a Go tool to convert Foreman's Apipie-format API documentation to OpenAPI 3.0 specifications.

**Tasks**:
- [ ] Create project structure in `tools/apipie-to-openapi/`
- [ ] Define Apipie format data structures
- [ ] Define OpenAPI 3.0 output structures
- [ ] Implement conversion logic for:
  - [ ] Endpoints → Paths
  - [ ] Parameters → OpenAPI parameters
  - [ ] Request/response schemas
  - [ ] Authentication schemes
- [ ] Add validation for generated OpenAPI specs
- [ ] Create comprehensive tests
- [ ] Document tool usage

**Acceptance Criteria**:
- [ ] Tool successfully converts sample Apipie doc to OpenAPI
- [ ] Generated OpenAPI validates with standard validators
- [ ] All tests pass
- [ ] Documentation complete

**Estimated Effort**: 3 days

**Related to**: Epic #TBD (API Client Modernization)
```

---

### Issue 1.3: Generate OpenAPI Specifications

```markdown
**Title**: Generate OpenAPI Specs for All Foreman APIs

**Labels**: api-client, phase1, enhancement

**Description**:
Use the converter tool to generate OpenAPI 3.0 specifications for all Foreman APIs.

**Tasks**:
- [ ] Generate Foreman core API OpenAPI spec
- [ ] Generate Katello API OpenAPI spec
- [ ] Generate Foreman Tasks API OpenAPI spec
- [ ] Generate Puppet Plugin API OpenAPI spec
- [ ] Validate all generated specs
- [ ] Manual review for accuracy
- [ ] Create regeneration scripts
- [ ] Add Makefile targets

**Acceptance Criteria**:
- [ ] All specs generated and validated
- [ ] Specs cover all endpoints
- [ ] Easy regeneration process documented
- [ ] Specs checked into version control

**Estimated Effort**: 2 days

**Related to**: Epic #TBD (API Client Modernization)
**Depends on**: #TBD (Issue 1.1, 1.2)
```

---

### Issue 1.4: Set Up oapi-codegen and Generate Client

```markdown
**Title**: Generate Go API Client from OpenAPI Specs

**Labels**: api-client, phase1, enhancement

**Description**:
Set up oapi-codegen and generate type-safe Go client code from OpenAPI specifications.

**Tasks**:
- [ ] Install and configure oapi-codegen
- [ ] Create codegen configurations for each API
- [ ] Generate Foreman core API client
- [ ] Generate Katello API client
- [ ] Generate Foreman Tasks API client
- [ ] Generate Puppet Plugin API client
- [ ] Create generation scripts
- [ ] Add Makefile targets
- [ ] Review generated code

**Acceptance Criteria**:
- [ ] All clients generated successfully
- [ ] Generated code compiles without errors
- [ ] Easy regeneration process
- [ ] Documentation complete

**Estimated Effort**: 3 days

**Related to**: Epic #TBD (API Client Modernization)
**Depends on**: #TBD (Issue 1.3)
```

---

### Issue 1.5: Create API Client Adapter Layer

```markdown
**Title**: Build Compatibility Adapter for Generated Client

**Labels**: api-client, phase1, enhancement

**Description**:
Create an adapter layer that wraps the generated API client and maintains backward compatibility with existing provider code.

**Tasks**:
- [ ] Design adapter interface
- [ ] Implement authentication middleware
- [ ] Implement taxonomy injection (location_id, organization_id)
- [ ] Create backward-compatible facades for existing API methods
- [ ] Implement error handling and conversion
- [ ] Add comprehensive tests
- [ ] Document adapter usage

**Acceptance Criteria**:
- [ ] Adapter maintains API compatibility
- [ ] All authentication methods supported
- [ ] Taxonomy parameters injected correctly
- [ ] All tests pass
- [ ] Documentation complete

**Estimated Effort**: 3 days

**Related to**: Epic #TBD (API Client Modernization)
**Depends on**: #TBD (Issue 1.4)
```

---

### Issue 1.6: Migrate Pilot Resources to Generated Client

```markdown
**Title**: Migrate 3-5 Resources to Use Generated API Client

**Labels**: api-client, phase1, enhancement

**Description**:
Migrate a small set of resources to use the new generated API client to validate the approach.

**Tasks**:
- [ ] Select pilot resources (Architecture, Model, Domain)
- [ ] Migrate Architecture resource
- [ ] Migrate Model resource
- [ ] Migrate Domain resource
- [ ] Update and verify tests for each
- [ ] Create migration guide for developers
- [ ] Document lessons learned

**Acceptance Criteria**:
- [ ] 3 resources successfully migrated
- [ ] All existing tests pass
- [ ] No functional regressions
- [ ] Migration guide created

**Estimated Effort**: 3 days

**Related to**: Epic #TBD (API Client Modernization)
**Depends on**: #TBD (Issue 1.5)
```

---

### Issue 1.7: Complete API Client Migration

```markdown
**Title**: Migrate All Remaining Resources to Generated Client

**Labels**: api-client, enhancement, help wanted

**Description**:
Complete the migration of all remaining resources (35+) to use the generated API client.

**Tasks**:
- [ ] Create tracking checklist of all resources
- [ ] Migrate resources in batches of 5-10
- [ ] Update tests for each batch
- [ ] Verify no regressions
- [ ] Update documentation
- [ ] Remove old custom client code

**Acceptance Criteria**:
- [ ] All resources migrated
- [ ] All tests passing
- [ ] Old client code removed
- [ ] Documentation updated

**Estimated Effort**: 4-6 weeks (can be done incrementally)

**Related to**: Epic #TBD (API Client Modernization)
**Depends on**: #TBD (Issue 1.6)

**Note**: This can be done incrementally over time, in parallel with other phases.
```

---

## Phase 2: Framework Migration - Detailed Issues

### Issue 2.1: Set Up Terraform Plugin Framework

```markdown
**Title**: Add Plugin Framework Dependencies and Muxing

**Labels**: framework, phase2, enhancement

**Description**:
Set up infrastructure for migrating to terraform-plugin-framework using muxing to support both SDK v2 and Framework simultaneously.

**Tasks**:
- [ ] Update Go to 1.22+
- [ ] Add terraform-plugin-framework dependency
- [ ] Add terraform-plugin-mux dependency
- [ ] Create framework provider skeleton
- [ ] Implement muxed provider in main.go
- [ ] Verify provider builds and runs

**Acceptance Criteria**:
- [ ] Framework dependencies added
- [ ] Muxed provider compiles
- [ ] Existing resources still work
- [ ] Ready for resource migration

**Estimated Effort**: 3 days

**Related to**: Epic #TBD (Framework Migration)
```

---

### Issue 2.2: Migrate Provider Configuration to Framework

```markdown
**Title**: Migrate Provider Configuration to Plugin Framework

**Labels**: framework, phase2, enhancement

**Description**:
Migrate provider configuration schema and logic to terraform-plugin-framework.

**Tasks**:
- [ ] Define framework provider schema
- [ ] Implement provider data structure
- [ ] Add configuration validation
- [ ] Migrate client initialization logic
- [ ] Test provider configuration
- [ ] Update documentation

**Acceptance Criteria**:
- [ ] Provider config works in framework
- [ ] All configuration options supported
- [ ] Environment variables work
- [ ] Tests pass

**Estimated Effort**: 2 days

**Related to**: Epic #TBD (Framework Migration)
**Depends on**: #TBD (Issue 2.1)
```

---

### Issue 2.3: Migrate Simple Resources to Framework

```markdown
**Title**: Migrate 10-15 Simple Resources to Plugin Framework

**Labels**: framework, phase2, enhancement

**Description**:
Migrate simple CRUD resources to terraform-plugin-framework.

**Resources to Migrate**:
- [ ] Architecture
- [ ] Model
- [ ] CommonParameter
- [ ] TemplateKind
- [ ] OperatingSystem
- [ ] Domain
- [ ] Environment
- [ ] Subnet
- [ ] PartitionTable
- [ ] Media

**For Each Resource**:
- [ ] Create framework resource implementation
- [ ] Migrate schema to framework types
- [ ] Implement CRUD methods
- [ ] Add ImportState support
- [ ] Migrate tests
- [ ] Verify functionality

**Acceptance Criteria**:
- [ ] All resources migrated successfully
- [ ] All tests passing
- [ ] No functional changes for users

**Estimated Effort**: 2 weeks

**Related to**: Epic #TBD (Framework Migration)
**Depends on**: #TBD (Issue 2.2)
```

---

### Issue 2.4: Migrate Complex Resources to Framework

```markdown
**Title**: Migrate Resources with Relationships to Plugin Framework

**Labels**: framework, phase2, enhancement

**Description**:
Migrate more complex resources with relationships and nested attributes to terraform-plugin-framework.

**Resources to Migrate**:
- [ ] HostGroup
- [ ] ComputeProfile
- [ ] ComputeResource
- [ ] Image
- [ ] SmartProxy
- [ ] ProvisioningTemplate
- [ ] JobTemplate
- [ ] Parameter
- [ ] Host
- [ ] ComputeAttribute
- [ ] User
- [ ] UserGroup

**Acceptance Criteria**:
- [ ] All resources migrated successfully
- [ ] Relationships handled correctly
- [ ] All tests passing

**Estimated Effort**: 3 weeks

**Related to**: Epic #TBD (Framework Migration)
**Depends on**: #TBD (Issue 2.3)
```

---

### Issue 2.5: Migrate Katello Resources to Framework

```markdown
**Title**: Migrate Katello Resources to Plugin Framework

**Labels**: framework, phase2, katello, enhancement

**Description**:
Migrate all Katello plugin resources to terraform-plugin-framework.

**Resources to Migrate**:
- [ ] Product
- [ ] Repository
- [ ] ContentCredential
- [ ] ContentView
- [ ] ContentViewVersion
- [ ] ContentViewFilter
- [ ] ContentViewFilterRule
- [ ] SyncPlan
- [ ] ActivationKey
- [ ] HostCollection

**Acceptance Criteria**:
- [ ] All Katello resources migrated
- [ ] All tests passing
- [ ] Katello workflows validated

**Estimated Effort**: 2 weeks

**Related to**: Epic #TBD (Framework Migration)
**Depends on**: #TBD (Issue 2.4)
```

---

### Issue 2.6: Migrate Data Sources to Framework

```markdown
**Title**: Migrate All Data Sources to Plugin Framework

**Labels**: framework, phase2, datasource, enhancement

**Description**:
Migrate all data sources to terraform-plugin-framework.

**Tasks**:
- [ ] Create data source migration checklist
- [ ] Migrate data sources in batches
- [ ] Update tests for each
- [ ] Verify query functionality

**Acceptance Criteria**:
- [ ] All data sources migrated
- [ ] All tests passing
- [ ] Query parameters work correctly

**Estimated Effort**: 2 weeks

**Related to**: Epic #TBD (Framework Migration)
**Depends on**: #TBD (Issue 2.5)
```

---

### Issue 2.7: Remove SDK v2 and Complete Migration

```markdown
**Title**: Remove terraform-plugin-sdk/v2 and Complete Framework Migration

**Labels**: framework, phase2, cleanup, enhancement

**Description**:
Complete the framework migration by removing all SDK v2 code and dependencies.

**Tasks**:
- [ ] Verify all resources and data sources migrated
- [ ] Remove mux from main.go
- [ ] Remove SDK v2 dependency from go.mod
- [ ] Delete old provider.go and resource files
- [ ] Update all imports
- [ ] Run full test suite
- [ ] Update documentation
- [ ] Create migration announcement

**Acceptance Criteria**:
- [ ] No SDK v2 code remaining
- [ ] All tests passing
- [ ] Clean build
- [ ] Documentation updated
- [ ] Users notified of changes

**Estimated Effort**: 1 week

**Related to**: Epic #TBD (Framework Migration)
**Depends on**: #TBD (Issue 2.6)
```

---

## Phase 3: E2E Testing - Detailed Issues

### Issue 3.1: Set Up Docker Test Infrastructure

```markdown
**Title**: Create Docker Compose Setup for Foreman Testing

**Labels**: testing, phase3, infrastructure, enhancement

**Description**:
Create Docker Compose configuration for running Foreman in a test environment.

**Tasks**:
- [ ] Create docker-compose.test.yml
- [ ] Configure Foreman container
- [ ] Configure PostgreSQL database
- [ ] Add health checks
- [ ] Create helper scripts (start, stop, wait)
- [ ] Test infrastructure locally

**Acceptance Criteria**:
- [ ] Foreman container starts successfully
- [ ] Health checks pass
- [ ] API accessible
- [ ] Scripts work reliably

**Estimated Effort**: 3 days

**Related to**: Epic #TBD (E2E Testing)
```

---

### Issue 3.2: Create E2E Test Framework

```markdown
**Title**: Build Go Test Framework for E2E Tests

**Labels**: testing, phase3, enhancement

**Description**:
Create test framework infrastructure for E2E tests using terraform-plugin-testing.

**Tasks**:
- [ ] Create test package structure
- [ ] Implement test helpers (client, cleanup, fixtures)
- [ ] Configure provider for tests
- [ ] Create provider config helper
- [ ] Add test utilities
- [ ] Document test framework usage

**Acceptance Criteria**:
- [ ] Test framework compiles
- [ ] Can connect to test Foreman
- [ ] Provider instantiation works
- [ ] Documentation complete

**Estimated Effort**: 2 days

**Related to**: Epic #TBD (E2E Testing)
**Depends on**: #TBD (Issue 3.1)
```

---

### Issue 3.3: Write Initial E2E Tests

```markdown
**Title**: Create E2E Tests for Core Resources

**Labels**: testing, phase3, enhancement

**Description**:
Write E2E tests for initial set of resources.

**Tasks**:
- [ ] Architecture resource E2E test
- [ ] Domain resource E2E test
- [ ] HostGroup resource E2E test
- [ ] Model resource E2E test
- [ ] Environment resource E2E test
- [ ] Run tests locally
- [ ] Verify CRUD operations
- [ ] Verify import functionality

**Acceptance Criteria**:
- [ ] 5 resources have E2E tests
- [ ] All tests pass against real Foreman
- [ ] Tests are maintainable

**Estimated Effort**: 1 week

**Related to**: Epic #TBD (E2E Testing)
**Depends on**: #TBD (Issue 3.2)
```

---

### Issue 3.4: Expand E2E Test Coverage

```markdown
**Title**: Add E2E Tests for All Resources

**Labels**: testing, phase3, enhancement, help wanted

**Description**:
Expand E2E test coverage to include all resources and data sources.

**Tasks**:
- [ ] Add tests for all simple resources (10-15 tests)
- [ ] Add tests for complex resources (10-15 tests)
- [ ] Add tests for Katello resources (10 tests)
- [ ] Add tests for data sources (15-20 tests)
- [ ] Add scenario tests (complex workflows)
- [ ] Organize tests by category

**Acceptance Criteria**:
- [ ] >80% of resources have E2E tests
- [ ] All critical paths tested
- [ ] Tests well organized
- [ ] Documentation complete

**Estimated Effort**: 2 weeks

**Related to**: Epic #TBD (E2E Testing)
**Depends on**: #TBD (Issue 3.3)
```

---

### Issue 3.5: Integrate E2E Tests into CI/CD

```markdown
**Title**: Add E2E Tests to GitHub Actions

**Labels**: testing, phase3, ci-cd, enhancement

**Description**:
Integrate E2E tests into GitHub Actions CI/CD pipeline.

**Tasks**:
- [ ] Create e2e-test.yml workflow
- [ ] Configure Docker in GitHub Actions
- [ ] Add Foreman startup and health checks
- [ ] Run E2E tests in CI
- [ ] Collect and upload logs on failure
- [ ] Add test matrix for multiple Foreman versions
- [ ] Optimize test execution time
- [ ] Document CI setup

**Acceptance Criteria**:
- [ ] E2E tests run in CI on PRs
- [ ] Tests complete in <30 minutes
- [ ] Logs captured on failure
- [ ] Multiple Foreman versions tested

**Estimated Effort**: 1 week

**Related to**: Epic #TBD (E2E Testing)
**Depends on**: #TBD (Issue 3.4)
```

---

## Additional Issues

### Documentation and Communication

```markdown
**Title**: Create Migration and Communication Plan

**Labels**: documentation, communication

**Description**:
Create comprehensive documentation and communication plan for modernization effort.

**Tasks**:
- [ ] Write migration guide for contributors
- [ ] Create upgrade guide for users
- [ ] Prepare announcement for community
- [ ] Create FAQ document
- [ ] Set up project board for tracking
- [ ] Create regular progress updates plan

**Acceptance Criteria**:
- [ ] All documentation complete
- [ ] Community informed
- [ ] Clear communication channels

**Estimated Effort**: 1 week
```

---

### Testing and Validation

```markdown
**Title**: Comprehensive Testing Before Release

**Labels**: testing, quality assurance

**Description**:
Perform comprehensive testing before major release of modernized provider.

**Tasks**:
- [ ] Run full test suite
- [ ] Performance testing
- [ ] Compatibility testing with different Foreman versions
- [ ] Manual testing of critical workflows
- [ ] Security audit
- [ ] Documentation review
- [ ] Beta testing with community

**Acceptance Criteria**:
- [ ] All automated tests pass
- [ ] No performance regressions
- [ ] Compatibility verified
- [ ] Security review complete
- [ ] Beta feedback addressed

**Estimated Effort**: 2 weeks
```

---

## Priority and Sequencing

### High Priority (Start Immediately)
1. Phase 1: API Client Modernization (Issues 1.1-1.6)
2. Phase 3: E2E Testing Infrastructure (Issues 3.1-3.3) - can run in parallel

### Medium Priority (After Initial Setup)
1. Phase 2: Framework Migration (Issues 2.1-2.3)
2. Phase 3: Expand E2E Tests (Issue 3.4)

### Lower Priority (Incremental Work)
1. Complete API migration (Issue 1.7)
2. Complete Framework migration (Issues 2.4-2.7)
3. CI Integration (Issue 3.5)

### Dependencies Graph
```
Phase 1: 1.1 → 1.2 → 1.3 → 1.4 → 1.5 → 1.6 → 1.7
                                              ↓
Phase 2: 2.1 → 2.2 → 2.3 → 2.4 → 2.5 → 2.6 → 2.7
         ↓
Phase 3: 3.1 → 3.2 → 3.3 → 3.4 → 3.5
```

## Notes for Issue Creation

When creating these issues in GitHub:

1. **Add appropriate labels**: `enhancement`, `api-client`, `framework`, `testing`, `documentation`, etc.
2. **Link to epic**: Reference the epic issue number
3. **Add to project board**: Create milestones for each phase
4. **Assign owners**: Distribute work among team members
5. **Set estimates**: Add time estimates for planning
6. **Create milestones**: Phase 1 (Month 1), Phase 2 (Months 2-3), Phase 3 (Month 3-4)
7. **Add dependencies**: Use GitHub issue dependencies feature

## Project Tracking

Create a GitHub Project board with columns:
- **Backlog**: All planned issues
- **Ready**: Issues ready to start (dependencies met)
- **In Progress**: Currently being worked on
- **Review**: Awaiting code review
- **Testing**: In testing phase
- **Done**: Completed and merged

## Success Metrics

Track these metrics throughout the project:
- [ ] API client code reduction (target: >50%)
- [ ] Test coverage increase (target: >80% E2E coverage)
- [ ] Framework migration completion (target: 100% resources)
- [ ] CI/CD test execution time (target: <30 minutes)
- [ ] Community feedback (track issues and PRs)

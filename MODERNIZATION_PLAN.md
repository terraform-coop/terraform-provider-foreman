# Terraform Foreman Provider Modernization Plan

## Executive Summary

This document outlines a comprehensive modernization strategy for the terraform-provider-foreman project, addressing three critical areas:

1. **API Client Implementation** - Transition from custom HTTP client to automated code generation
2. **Terraform Framework** - Migrate from Plugin SDK v2 to modern Plugin Framework
3. **E2E Testing Infrastructure** - Implement automated testing against real Foreman instances

## Current State Analysis

### API Implementation
- **Current Approach**: Custom HTTP client with 40+ manually maintained resource files
- **Location**: `/foreman/api/` directory
- **Authentication**: Supports Basic Auth and HTTP Negotiate/SPNEGO
- **API Coverage**: Comprehensive coverage of Foreman v2 API, Katello, Foreman Tasks, and Puppet plugins
- **Issues**: 
  - High maintenance burden (manual updates for every API change)
  - Potential for inconsistencies and drift from upstream API
  - No automatic validation against API specification
  - Significant code duplication across resource files

### Terraform Framework
- **Current Version**: terraform-plugin-sdk/v2 v2.24.0
- **Status**: Legacy framework, in maintenance-only mode
- **Go Version**: 1.18+ (should upgrade to 1.22+)
- **Issues**:
  - No new features or improvements
  - Less type-safe than modern framework
  - Missing modern abstractions and better error handling

### Testing Infrastructure
- **Current Tests**: Unit tests and provider integrity checks
- **CI/CD**: GitHub Actions running tests on Go 1.21+ and 1.22+
- **Gaps**:
  - No integration tests against real Foreman instance
  - No E2E acceptance tests
  - Tests rely on mocked responses or test data
  - Cannot validate actual API compatibility

## Modernization Strategy

### Phase 1: API Client Modernization

#### Option A: OpenAPI-Based Code Generation (Recommended)
**Approach**: Convert Foreman's Apipie documentation to OpenAPI and use oapi-codegen

**Benefits**:
- Industry-standard approach
- Automatic client code generation
- Type-safe API interactions
- Easy to update when API changes
- Rich tooling ecosystem

**Implementation Steps**:
1. Create Apipie-to-OpenAPI converter
2. Extract OpenAPI spec from Foreman instance
3. Generate Go client using oapi-codegen
4. Create adapter layer for Terraform provider
5. Gradually migrate resources to use generated client

**Tools & Technologies**:
- `oapi-codegen` - Go client generator from OpenAPI specs
- Foreman's `/apidoc/api.json` endpoint
- Custom conversion scripts for Apipie → OpenAPI

**Estimated Effort**: 3-4 weeks
- Week 1: Spec extraction and conversion tooling
- Week 2: Client generation and adapter layer
- Week 3-4: Resource migration and testing

#### Option B: Direct Apipie Integration
**Approach**: Build custom code generator that reads Apipie JSON directly

**Benefits**:
- No intermediate conversion step
- Potentially more accurate mapping

**Drawbacks**:
- Custom tooling to maintain
- Less ecosystem support
- Apipie format is Rails-specific

**Estimated Effort**: 4-5 weeks

#### Recommendation: Pursue Option A
OpenAPI is the industry standard and provides better long-term maintainability. The conversion effort is worthwhile for the ecosystem benefits.

### Phase 2: Terraform Plugin Framework Migration

#### Migration Approach: Incremental with Muxing
**Strategy**: Use terraform-plugin-mux to run both SDK v2 and Framework simultaneously

**Benefits**:
- Gradual migration reduces risk
- Can ship improvements incrementally
- Easier to test and validate
- Maintains backward compatibility

**Implementation Steps**:

1. **Setup Foundation** (Week 1)
   - Upgrade Go to 1.22+
   - Add terraform-plugin-framework dependency
   - Add terraform-plugin-mux for hybrid operation
   - Update main.go to serve muxed provider

2. **Migrate Provider Core** (Week 2)
   - Implement provider.Provider interface
   - Migrate provider schema to framework
   - Migrate configuration and client setup
   - Test provider initialization

3. **Migrate Resources** (Weeks 3-8)
   - Start with simpler resources (e.g., Architecture, Model)
   - Migrate 5-10 resources per week
   - Priority order:
     1. Simple CRUD resources (Architecture, Model, Domain)
     2. Resources with relationships (Host Groups)
     3. Complex resources (Hosts, Compute Resources)
     4. Katello resources (Products, Repositories)
   - Maintain parallel test coverage

4. **Migrate Data Sources** (Weeks 9-10)
   - Follow same pattern as resources
   - Update documentation

5. **Remove SDK v2** (Week 11)
   - Remove mux wrapper
   - Remove SDK v2 dependency
   - Final testing and validation

**Tools & Technologies**:
- `terraform-plugin-framework` v1.x
- `terraform-plugin-mux` for hybrid providers
- `terraform-plugin-testing` for acceptance tests

**Estimated Effort**: 11 weeks

**Key Considerations**:
- User-facing Terraform configurations remain compatible
- No breaking changes for users
- Internal implementation changes only
- Comprehensive test coverage required

### Phase 3: E2E Testing Infrastructure

#### Testing Strategy: Containerized Foreman with Docker Compose

**Approach**: Spin up real Foreman instance in Docker for acceptance tests

**Components**:
1. **Docker Compose Setup**
   - Foreman container (official or custom image)
   - PostgreSQL database
   - Test data initialization scripts
   - Network configuration

2. **Test Framework**
   - Use terraform-plugin-testing framework
   - Acceptance tests with TF_ACC=1
   - Real Terraform operations (apply, plan, destroy)
   - Verification against actual API

3. **CI/CD Integration**
   - GitHub Actions workflow for E2E tests
   - Container lifecycle management
   - Test result reporting
   - Artifact collection (logs, screenshots)

**Implementation Steps**:

1. **Infrastructure Setup** (Week 1)
   ```yaml
   # docker-compose.test.yml
   services:
     postgres:
       image: postgres:13
       environment:
         POSTGRES_DB: foreman
         POSTGRES_USER: foreman
         POSTGRES_PASSWORD: foreman
     
     foreman:
       image: theforeman/foreman:latest
       depends_on:
         - postgres
       environment:
         DATABASE_URL: postgresql://foreman:foreman@postgres/foreman
       ports:
         - "3000:3000"
       healthcheck:
         test: ["CMD", "curl", "-f", "http://localhost:3000/"]
         interval: 10s
         timeout: 5s
         retries: 5
   ```

2. **Test Utilities** (Week 1)
   - Helper functions for test setup/teardown
   - Common test fixtures and data
   - Foreman instance health checks
   - Authentication helpers

3. **Initial Test Suite** (Week 2)
   - Architecture resource tests
   - Domain resource tests
   - Host group tests
   - Basic CRUD validation

4. **Expand Test Coverage** (Weeks 3-4)
   - Add tests for all resources
   - Test complex scenarios (relationships, dependencies)
   - Test error conditions
   - Test update/import scenarios

5. **CI Integration** (Week 5)
   ```yaml
   # .github/workflows/e2e-test.yml
   name: E2E Tests
   on:
     pull_request:
       branches: [master]
   jobs:
     e2e-test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         - uses: actions/setup-go@v3
           with:
             go-version: '>=1.22'
         - name: Start Foreman
           run: docker-compose -f docker-compose.test.yml up -d
         - name: Wait for Foreman
           run: ./scripts/wait-for-foreman.sh
         - name: Run E2E Tests
           run: TF_ACC=1 go test -v -timeout 30m ./...
           env:
             FOREMAN_URL: http://localhost:3000
             FOREMAN_USERNAME: admin
             FOREMAN_PASSWORD: changeme
         - name: Collect Logs
           if: failure()
           run: docker-compose -f docker-compose.test.yml logs
         - name: Cleanup
           if: always()
           run: docker-compose -f docker-compose.test.yml down -v
   ```

**Tools & Technologies**:
- Docker & Docker Compose
- GitHub Actions
- terraform-plugin-testing
- Official Foreman Docker images or custom builds

**Estimated Effort**: 5 weeks

**Key Considerations**:
- Test isolation and cleanup
- Parallel test execution
- Test data management
- Performance optimization
- Foreman version compatibility testing

## Implementation Timeline

### Overall Schedule: 19 Weeks (~4.5 Months)

```
Phase 1: API Client Modernization     [Weeks 1-4]
Phase 2: Framework Migration           [Weeks 5-15]
Phase 3: E2E Testing Infrastructure    [Weeks 11-15] (Parallel with Phase 2)
Final Integration & Documentation      [Weeks 16-19]
```

### Detailed Timeline

| Week | Phase | Tasks |
|------|-------|-------|
| 1-2 | Phase 1 | Apipie to OpenAPI conversion, client generation setup |
| 3-4 | Phase 1 | Generate client, create adapter, migrate sample resources |
| 5 | Phase 2 | Framework setup, provider core migration |
| 6-10 | Phase 2 | Resource migration (5-10 per week) |
| 11 | Phase 3 | Docker setup, test infrastructure (parallel with Phase 2) |
| 12-13 | Phase 3 | Write initial test suite (parallel with Phase 2) |
| 14-15 | Phase 2 & 3 | Complete resource migration, expand tests |
| 16-17 | All | Integration testing, bug fixes |
| 18-19 | All | Documentation, final validation, release prep |

## Dependencies and Prerequisites

### Technical Requirements
- Go 1.22+ for optimal framework support
- Docker for local testing
- Access to Foreman instance for API documentation extraction
- CI/CD environment with Docker support

### Knowledge Requirements
- OpenAPI/Swagger specification format
- Terraform Plugin Framework APIs
- Docker and container orchestration
- Go testing frameworks

### External Dependencies
- Foreman API stability (avoid major version changes during migration)
- oapi-codegen tool compatibility
- Docker image availability for Foreman

## Risk Assessment and Mitigation

### High Risks

1. **API Spec Conversion Accuracy**
   - Risk: Apipie → OpenAPI conversion loses information
   - Mitigation: Extensive validation, manual review, test coverage
   - Impact: High (affects all API operations)

2. **Breaking Changes During Migration**
   - Risk: Framework migration introduces bugs
   - Mitigation: Incremental approach with muxing, comprehensive tests
   - Impact: High (affects all users)

3. **E2E Test Infrastructure Reliability**
   - Risk: Flaky tests, environment issues
   - Mitigation: Robust setup/teardown, health checks, retries
   - Impact: Medium (affects CI/CD confidence)

### Medium Risks

1. **Timeline Overruns**
   - Risk: Complexity underestimated
   - Mitigation: Buffer time, incremental releases
   - Impact: Medium (delays benefits)

2. **Resource Constraints**
   - Risk: Insufficient development capacity
   - Mitigation: Community involvement, phased approach
   - Impact: Medium (slows progress)

3. **Foreman API Changes**
   - Risk: Upstream API changes during migration
   - Mitigation: Pin Foreman version for testing, monitor changes
   - Impact: Medium (requires adaptation)

### Low Risks

1. **Tool Compatibility Issues**
   - Risk: Code generation tools have limitations
   - Mitigation: Evaluate alternatives, custom adaptations
   - Impact: Low (workarounds available)

## Success Metrics

### API Client Modernization
- [ ] Automated client generation from Foreman API spec
- [ ] 100% resource coverage with generated client
- [ ] Reduced API client code by >50%
- [ ] API version updates possible in <1 day

### Framework Migration
- [ ] All resources using Plugin Framework
- [ ] Zero breaking changes for users
- [ ] Test coverage maintained or improved
- [ ] Better error messages and diagnostics

### E2E Testing
- [ ] Tests running against real Foreman instance
- [ ] >80% resource coverage in E2E tests
- [ ] E2E tests in CI pipeline
- [ ] Test execution time <30 minutes

## Resource Requirements

### Development Time
- **Phase 1**: 160 hours (1 developer, 4 weeks)
- **Phase 2**: 440 hours (1 developer, 11 weeks)
- **Phase 3**: 200 hours (1 developer, 5 weeks, partial overlap)
- **Total**: ~800 hours

### Infrastructure
- GitHub Actions minutes for CI/CD
- Docker registry storage (if custom images)
- Test environment resources

### Community Involvement
- Code reviews from maintainers
- Testing from community users
- Documentation updates
- Issue triage and bug fixes

## Maintenance and Long-Term Benefits

### Reduced Maintenance Burden
- API changes: Days instead of weeks
- Type safety prevents many bugs
- Automated testing catches regressions
- Modern framework receives ongoing support

### Improved User Experience
- Better error messages
- More consistent behavior
- Faster bug fixes
- Modern Terraform features support

### Developer Experience
- Easier onboarding for contributors
- Standard tooling and patterns
- Better IDE support
- Comprehensive test coverage

## Conclusion

This modernization plan addresses critical technical debt in the terraform-provider-foreman project while maintaining backward compatibility and user trust. The phased approach allows for incremental progress and risk mitigation.

### Recommended Next Steps

1. **Immediate Actions**
   - Review and approve this plan
   - Set up project tracking (GitHub Projects/Issues)
   - Allocate development resources
   - Create feature branches for each phase

2. **Phase 1 Kickoff**
   - Set up development environment
   - Extract Foreman API documentation
   - Begin Apipie → OpenAPI conversion
   - Create proof-of-concept with 2-3 resources

3. **Communication**
   - Announce modernization effort to community
   - Create RFC for major changes
   - Document migration guide for contributors
   - Regular progress updates

### Getting Started

To begin implementation, start with Phase 1 (API Client Modernization) as it provides the foundation for improved maintainability and sets the stage for the other phases.

See the companion documents for detailed implementation guides:
- `docs/api-client-migration.md` (to be created)
- `docs/framework-migration.md` (to be created)
- `docs/e2e-testing-guide.md` (to be created)

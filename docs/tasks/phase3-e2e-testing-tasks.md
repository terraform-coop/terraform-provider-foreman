# Phase 3: E2E Testing Infrastructure - Detailed Tasks

## Overview
Implement comprehensive end-to-end testing infrastructure with containerized Foreman instances.

## Prerequisites
- [ ] Docker installed and running
- [ ] Docker Compose v2+ installed
- [ ] GitHub Actions environment (for CI integration)
- [ ] Feature branch created: `feature/e2e-testing`

## Task Breakdown

### Task 3.1: Docker Infrastructure Setup (Week 11, Days 1-3)

**Goal**: Create Docker Compose setup for Foreman testing environment

#### Subtasks:

1. [ ] Create test infrastructure directory
   ```bash
   mkdir -p test/docker
   cd test/docker
   ```

2. [ ] Create docker-compose.test.yml
   ```yaml
   # test/docker/docker-compose.test.yml
   version: '3.8'
   
   services:
     postgres:
       image: postgres:13-alpine
       environment:
         POSTGRES_DB: foreman
         POSTGRES_USER: foreman
         POSTGRES_PASSWORD: foreman
       volumes:
         - postgres_data:/var/lib/postgresql/data
       healthcheck:
         test: ["CMD-SHELL", "pg_isready -U foreman"]
         interval: 5s
         timeout: 5s
         retries: 5
       networks:
         - foreman_network
   
     foreman:
       image: quay.io/foreman/foreman:latest
       depends_on:
         postgres:
           condition: service_healthy
       environment:
         DATABASE_URL: postgresql://foreman:foreman@postgres/foreman
         FOREMAN_ADMIN_PASSWORD: changeme
         FOREMAN_INITIAL_ADMIN_USERNAME: admin
         FOREMAN_INITIAL_ADMIN_PASSWORD: changeme
         RAILS_ENV: production
       ports:
         - "3000:3000"
       volumes:
         - ./init-scripts:/docker-entrypoint-initdb.d
       healthcheck:
         test: ["CMD", "curl", "-f", "http://localhost:3000/api/status"]
         interval: 10s
         timeout: 5s
         retries: 30
         start_period: 60s
       networks:
         - foreman_network
   
   networks:
     foreman_network:
       driver: bridge
   
   volumes:
     postgres_data:
   ```

3. [ ] Create initialization scripts
   ```bash
   # test/docker/init-scripts/01-setup-foreman.sh
   #!/bin/bash
   set -e
   
   echo "Waiting for Foreman to be ready..."
   until curl -f http://localhost:3000/api/status; do
     sleep 5
   done
   
   echo "Setting up test data..."
   # Add any initial test data setup here
   ```

4. [ ] Create helper scripts
   ```bash
   # test/docker/start-test-env.sh
   #!/bin/bash
   set -e
   
   echo "Starting Foreman test environment..."
   docker-compose -f docker-compose.test.yml up -d
   
   echo "Waiting for services to be healthy..."
   docker-compose -f docker-compose.test.yml ps
   
   ./wait-for-foreman.sh
   
   echo "Test environment ready!"
   echo "Foreman URL: http://localhost:3000"
   echo "Username: admin"
   echo "Password: changeme"
   ```

   ```bash
   # test/docker/wait-for-foreman.sh
   #!/bin/bash
   set -e
   
   MAX_RETRIES=60
   RETRY_COUNT=0
   
   echo "Waiting for Foreman API to be available..."
   
   until curl -f -s -u admin:changeme http://localhost:3000/api/v2/status > /dev/null; do
     RETRY_COUNT=$((RETRY_COUNT + 1))
     if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
       echo "Foreman failed to start within expected time"
       docker-compose -f docker-compose.test.yml logs foreman
       exit 1
     fi
     echo "Waiting... (attempt $RETRY_COUNT/$MAX_RETRIES)"
     sleep 5
   done
   
   echo "Foreman is ready!"
   ```

   ```bash
   # test/docker/stop-test-env.sh
   #!/bin/bash
   set -e
   
   echo "Stopping Foreman test environment..."
   docker-compose -f docker-compose.test.yml down -v
   
   echo "Test environment stopped and cleaned up!"
   ```

5. [ ] Make scripts executable
   ```bash
   chmod +x test/docker/*.sh
   ```

6. [ ] Test Docker setup locally
   ```bash
   cd test/docker
   ./start-test-env.sh
   
   # Verify Foreman is accessible
   curl -u admin:changeme http://localhost:3000/api/v2/status
   
   # Clean up
   ./stop-test-env.sh
   ```

**Deliverables**:
- Docker Compose configuration
- Helper scripts for environment management
- Verified working Foreman container

**Success Criteria**:
- Foreman container starts successfully
- API is accessible
- Health checks pass
- Can create/read via API

---

### Task 3.2: Test Framework Setup (Week 11, Days 4-5)

**Goal**: Create Go test infrastructure for E2E tests

#### Subtasks:

1. [ ] Create test package structure
   ```
   test/
   ├── docker/              # Docker setup (from Task 3.1)
   ├── e2e/
   │   ├── provider_test.go # Provider setup tests
   │   ├── architecture_test.go
   │   ├── domain_test.go
   │   ├── helpers/
   │   │   ├── client.go    # Test client helpers
   │   │   ├── cleanup.go   # Cleanup utilities
   │   │   └── fixtures.go  # Test data fixtures
   │   └── testdata/
   │       └── configs/     # Terraform configs for tests
   └── README.md
   ```

2. [ ] Create test helpers
   ```go
   // test/e2e/helpers/client.go
   package helpers
   
   import (
       "context"
       "os"
       "testing"
       
       "github.com/terraform-coop/terraform-provider-foreman/foreman/api/adapter"
   )
   
   // GetTestClient creates a client for E2E tests
   func GetTestClient(t *testing.T) *adapter.ForemanClient {
       t.Helper()
       
       config := adapter.ClientConfig{
           BaseURL:            getEnvOrDefault("FOREMAN_URL", "http://localhost:3000"),
           Username:           getEnvOrDefault("FOREMAN_USERNAME", "admin"),
           Password:           getEnvOrDefault("FOREMAN_PASSWORD", "changeme"),
           TLSInsecureEnabled: true, // For test environment
       }
       
       client, err := adapter.NewClient(config)
       if err != nil {
           t.Fatalf("Failed to create test client: %v", err)
       }
       
       return client
   }
   
   func getEnvOrDefault(key, defaultVal string) string {
       if val := os.Getenv(key); val != "" {
           return val
       }
       return defaultVal
   }
   ```

   ```go
   // test/e2e/helpers/cleanup.go
   package helpers
   
   import (
       "context"
       "testing"
   )
   
   // CleanupArchitecture removes test architecture
   func CleanupArchitecture(t *testing.T, client *adapter.ForemanClient, id int) {
       t.Helper()
       
       ctx := context.Background()
       err := client.DeleteArchitecture(ctx, id)
       if err != nil {
           t.Logf("Warning: Failed to cleanup architecture %d: %v", id, err)
       }
   }
   
   // RegisterCleanup registers a cleanup function to run after test
   func RegisterCleanup(t *testing.T, fn func()) {
       t.Cleanup(fn)
   }
   ```

3. [ ] Create provider test configuration
   ```go
   // test/e2e/provider_test.go
   package e2e
   
   import (
       "context"
       "os"
       "testing"
       
       "github.com/hashicorp/terraform-plugin-framework/providerserver"
       "github.com/hashicorp/terraform-plugin-go/tfprotov6"
       "github.com/hashicorp/terraform-plugin-testing/helper/resource"
       
       "github.com/terraform-coop/terraform-provider-foreman/foreman/framework/provider"
       "github.com/terraform-coop/terraform-provider-foreman/test/e2e/helpers"
   )
   
   var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
       "foreman": providerserver.NewProtocol6WithError(provider.New("test")()),
   }
   
   func TestMain(m *testing.M) {
       // Skip E2E tests if TF_ACC is not set
       if os.Getenv("TF_ACC") != "1" {
           os.Exit(0)
       }
       
       // Verify test environment is available
       client := helpers.GetTestClient(&testing.T{})
       if client == nil {
           panic("Cannot connect to test Foreman instance")
       }
       
       // Run tests
       os.Exit(m.Run())
   }
   ```

4. [ ] Create provider config for tests
   ```go
   // test/e2e/helpers/provider_config.go
   package helpers
   
   import "fmt"
   
   // ProviderConfig generates provider configuration for tests
   func ProviderConfig() string {
       return fmt.Sprintf(`
   provider "foreman" {
     server_hostname = "%s"
     client_username = "%s"
     client_password = "%s"
     tls_insecure    = true
   }
   `, 
           getEnvOrDefault("FOREMAN_URL", "http://localhost:3000"),
           getEnvOrDefault("FOREMAN_USERNAME", "admin"),
           getEnvOrDefault("FOREMAN_PASSWORD", "changeme"),
       )
   }
   ```

**Deliverables**:
- Test package structure
- Helper functions for E2E tests
- Provider test configuration

**Success Criteria**:
- Test framework compiles
- Can connect to Foreman
- Provider can be instantiated

---

### Task 3.3: Initial Test Suite (Week 12)

**Goal**: Write E2E tests for initial resources

#### Subtasks:

1. [ ] Create Architecture resource E2E test
   ```go
   // test/e2e/architecture_test.go
   package e2e
   
   import (
       "fmt"
       "testing"
       
       "github.com/hashicorp/terraform-plugin-testing/helper/resource"
       "github.com/terraform-coop/terraform-provider-foreman/test/e2e/helpers"
   )
   
   func TestAccArchitectureResource(t *testing.T) {
       resource.Test(t, resource.TestCase{
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               // Create and Read testing
               {
                   Config: testAccArchitectureResourceConfig("x86_64_test"),
                   Check: resource.ComposeAggregateTestCheckFunc(
                       resource.TestCheckResourceAttr("foreman_architecture.test", "name", "x86_64_test"),
                       resource.TestCheckResourceAttrSet("foreman_architecture.test", "id"),
                   ),
               },
               // Update testing
               {
                   Config: testAccArchitectureResourceConfig("x86_64_test_updated"),
                   Check: resource.ComposeAggregateTestCheckFunc(
                       resource.TestCheckResourceAttr("foreman_architecture.test", "name", "x86_64_test_updated"),
                   ),
               },
               // ImportState testing
               {
                   ResourceName:      "foreman_architecture.test",
                   ImportState:       true,
                   ImportStateVerify: true,
               },
           },
       })
   }
   
   func testAccArchitectureResourceConfig(name string) string {
       return helpers.ProviderConfig() + fmt.Sprintf(`
   resource "foreman_architecture" "test" {
     name = %[1]q
   }
   `, name)
   }
   ```

2. [ ] Create Domain resource E2E test
   ```go
   // test/e2e/domain_test.go
   package e2e
   
   import (
       "fmt"
       "testing"
       
       "github.com/hashicorp/terraform-plugin-testing/helper/resource"
       "github.com/terraform-coop/terraform-provider-foreman/test/e2e/helpers"
   )
   
   func TestAccDomainResource(t *testing.T) {
       resource.Test(t, resource.TestCase{
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               {
                   Config: testAccDomainResourceConfig("example.com", "Example Domain"),
                   Check: resource.ComposeAggregateTestCheckFunc(
                       resource.TestCheckResourceAttr("foreman_domain.test", "name", "example.com"),
                       resource.TestCheckResourceAttr("foreman_domain.test", "fullname", "Example Domain"),
                       resource.TestCheckResourceAttrSet("foreman_domain.test", "id"),
                   ),
               },
               {
                   Config: testAccDomainResourceConfig("example.com", "Updated Example Domain"),
                   Check: resource.ComposeAggregateTestCheckFunc(
                       resource.TestCheckResourceAttr("foreman_domain.test", "fullname", "Updated Example Domain"),
                   ),
               },
               {
                   ResourceName:      "foreman_domain.test",
                   ImportState:       true,
                   ImportStateVerify: true,
               },
           },
       })
   }
   
   func testAccDomainResourceConfig(name, fullname string) string {
       return helpers.ProviderConfig() + fmt.Sprintf(`
   resource "foreman_domain" "test" {
     name     = %[1]q
     fullname = %[2]q
   }
   `, name, fullname)
   }
   ```

3. [ ] Create HostGroup resource E2E test
   ```go
   // test/e2e/hostgroup_test.go
   package e2e
   
   import (
       "fmt"
       "testing"
       
       "github.com/hashicorp/terraform-plugin-testing/helper/resource"
       "github.com/terraform-coop/terraform-provider-foreman/test/e2e/helpers"
   )
   
   func TestAccHostGroupResource(t *testing.T) {
       resource.Test(t, resource.TestCase{
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               {
                   Config: testAccHostGroupResourceConfig("test-hostgroup", "Test Host Group"),
                   Check: resource.ComposeAggregateTestCheckFunc(
                       resource.TestCheckResourceAttr("foreman_hostgroup.test", "name", "test-hostgroup"),
                       resource.TestCheckResourceAttr("foreman_hostgroup.test", "title", "Test Host Group"),
                       resource.TestCheckResourceAttrSet("foreman_hostgroup.test", "id"),
                   ),
               },
               {
                   ResourceName:      "foreman_hostgroup.test",
                   ImportState:       true,
                   ImportStateVerify: true,
               },
           },
       })
   }
   
   func testAccHostGroupResourceConfig(name, title string) string {
       return helpers.ProviderConfig() + fmt.Sprintf(`
   resource "foreman_hostgroup" "test" {
     name  = %[1]q
     title = %[2]q
   }
   `, name, title)
   }
   ```

4. [ ] Run tests locally
   ```bash
   # Start test environment
   cd test/docker
   ./start-test-env.sh
   
   # Run E2E tests
   cd ../..
   TF_ACC=1 go test -v ./test/e2e/... -timeout 30m
   
   # Cleanup
   cd test/docker
   ./stop-test-env.sh
   ```

**Deliverables**:
- E2E tests for 3-5 resources
- All tests passing locally
- Test documentation

**Success Criteria**:
- Tests create/read/update/delete resources
- Tests can import existing resources
- All tests pass against real Foreman

---

### Task 3.4: Expand Test Coverage (Week 13-14)

**Goal**: Add E2E tests for all resources

#### Week 13 Tasks:
- [ ] Add tests for all simple resources (10-15 tests)
- [ ] Add tests for medium complexity resources (5-10 tests)
- [ ] Add tests for resources with relationships

#### Week 14 Tasks:
- [ ] Add tests for Katello resources (10 tests)
- [ ] Add tests for data sources (10-15 tests)
- [ ] Add complex scenario tests (combinations of resources)

#### Test Organization:
```
test/e2e/
├── resources/
│   ├── architecture_test.go
│   ├── domain_test.go
│   ├── host_test.go
│   ├── hostgroup_test.go
│   └── ... (all resources)
├── datasources/
│   ├── architecture_test.go
│   ├── domain_test.go
│   └── ... (all data sources)
└── scenarios/
    ├── complete_host_test.go
    ├── katello_workflow_test.go
    └── ... (complex scenarios)
```

**Success Criteria**:
- >80% of resources have E2E tests
- All critical paths tested
- Tests are maintainable and documented

---

### Task 3.5: CI/CD Integration (Week 15)

**Goal**: Integrate E2E tests into GitHub Actions

#### Subtasks:

1. [ ] Create E2E test workflow
   ```yaml
   # .github/workflows/e2e-test.yml
   name: E2E Tests
   
   on:
     pull_request:
       branches: [master]
       paths:
         - '**.go'
         - 'test/**'
         - '.github/workflows/e2e-test.yml'
     push:
       branches: [master]
   
   env:
     GO_VERSION: '1.22'
   
   jobs:
     e2e-test:
       runs-on: ubuntu-latest
       timeout-minutes: 60
       
       steps:
         - name: Checkout code
           uses: actions/checkout@v4
         
         - name: Set up Go
           uses: actions/setup-go@v4
           with:
             go-version: ${{ env.GO_VERSION }}
         
         - name: Start Foreman test environment
           working-directory: test/docker
           run: |
             docker-compose -f docker-compose.test.yml up -d
             ./wait-for-foreman.sh
         
         - name: Run E2E tests
           env:
             TF_ACC: "1"
             FOREMAN_URL: "http://localhost:3000"
             FOREMAN_USERNAME: "admin"
             FOREMAN_PASSWORD: "changeme"
           run: |
             go test -v -timeout 45m ./test/e2e/... \
               -coverprofile=coverage-e2e.out
         
         - name: Upload coverage
           uses: codecov/codecov-action@v3
           with:
             files: ./coverage-e2e.out
             flags: e2e-tests
         
         - name: Collect logs on failure
           if: failure()
           run: |
             mkdir -p test-logs
             docker-compose -f test/docker/docker-compose.test.yml logs > test-logs/docker-logs.txt
         
         - name: Upload logs
           if: failure()
           uses: actions/upload-artifact@v3
           with:
             name: test-logs
             path: test-logs/
             retention-days: 7
         
         - name: Cleanup
           if: always()
           working-directory: test/docker
           run: docker-compose -f docker-compose.test.yml down -v
   ```

2. [ ] Add test matrix for multiple Foreman versions
   ```yaml
   jobs:
     e2e-test:
       strategy:
         matrix:
           foreman_version: ['latest', '3.8', '3.9']
       # ... rest of job definition
       
       steps:
         # ... other steps
         
         - name: Start Foreman test environment
           env:
             FOREMAN_VERSION: ${{ matrix.foreman_version }}
           working-directory: test/docker
           run: |
             docker-compose -f docker-compose.test.yml up -d
   ```

3. [ ] Update docker-compose to support version variable
   ```yaml
   # test/docker/docker-compose.test.yml
   services:
     foreman:
       image: quay.io/foreman/foreman:${FOREMAN_VERSION:-latest}
       # ... rest of config
   ```

4. [ ] Add README for E2E tests
   ```markdown
   # E2E Testing Guide
   
   ## Running Tests Locally
   
   1. Start test environment:
      ```bash
      cd test/docker
      ./start-test-env.sh
      ```
   
   2. Run tests:
      ```bash
      TF_ACC=1 go test -v ./test/e2e/...
      ```
   
   3. Cleanup:
      ```bash
      cd test/docker
      ./stop-test-env.sh
      ```
   
   ## CI/CD
   
   E2E tests run automatically on pull requests via GitHub Actions.
   
   ## Debugging
   
   See logs: `docker-compose -f test/docker/docker-compose.test.yml logs`
   ```

5. [ ] Test CI integration
   - [ ] Create test PR
   - [ ] Verify workflow runs
   - [ ] Verify tests execute
   - [ ] Verify cleanup happens

**Deliverables**:
- GitHub Actions workflow for E2E tests
- Support for multiple Foreman versions
- Documentation for running tests
- Verified CI integration

**Success Criteria**:
- E2E tests run in CI
- Tests complete in <30 minutes
- Logs captured on failure
- Cleanup always happens

---

## Test Data Management

### Strategy:

1. **Ephemeral Data**: Tests create and destroy their own data
2. **Isolation**: Each test uses unique names/IDs
3. **Cleanup**: Always cleanup in `t.Cleanup()` or defer
4. **Fixtures**: Shared test data in `test/e2e/helpers/fixtures.go`

### Example Fixture:
```go
// test/e2e/helpers/fixtures.go
package helpers

import "fmt"

// UniqueArchitectureName generates unique architecture name for tests
func UniqueArchitectureName(prefix string) string {
    return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// StandardTestArchitecture creates a standard test architecture
func StandardTestArchitecture(t *testing.T, client *adapter.ForemanClient) *adapter.ForemanArchitecture {
    t.Helper()
    
    arch := &adapter.ForemanArchitecture{
        Name: UniqueArchitectureName("test-arch"),
    }
    
    created, err := client.CreateArchitecture(context.Background(), arch)
    if err != nil {
        t.Fatalf("Failed to create test architecture: %v", err)
    }
    
    t.Cleanup(func() {
        CleanupArchitecture(t, client, created.ID)
    })
    
    return created
}
```

---

## Definition of Done

Phase 3 is complete when:

- [ ] Docker Compose setup working
- [ ] Test framework in place
- [ ] >80% resource coverage in E2E tests
- [ ] E2E tests integrated in CI
- [ ] Tests run successfully against real Foreman
- [ ] Documentation complete
- [ ] Test execution time <30 minutes

## Performance Optimization

### Strategies:
1. **Parallel Execution**: Run independent tests in parallel
2. **Resource Reuse**: Share Foreman instance across tests
3. **Fast Cleanup**: Efficient resource deletion
4. **Selective Testing**: Run only affected tests in CI

### Parallel Tests:
```go
func TestAccResourcesParallel(t *testing.T) {
    t.Run("Architecture", func(t *testing.T) {
        t.Parallel()
        TestAccArchitectureResource(t)
    })
    
    t.Run("Domain", func(t *testing.T) {
        t.Parallel()
        TestAccDomainResource(t)
    })
}
```

---

## Troubleshooting Guide

### Common Issues:

1. **Foreman won't start**
   - Check Docker logs: `docker-compose logs foreman`
   - Verify ports not in use: `lsof -i :3000`
   - Check disk space: `df -h`

2. **Tests timeout**
   - Increase timeout: `-timeout 60m`
   - Check Foreman health: `curl http://localhost:3000/api/status`
   - Review test logs

3. **Cleanup failures**
   - Manual cleanup: `./stop-test-env.sh`
   - Remove volumes: `docker volume prune`

---

## Tools and References

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Terraform Plugin Testing](https://developer.hashicorp.com/terraform/plugin/testing)
- [Foreman Docker Images](https://quay.io/repository/foreman/foreman)
- [GitHub Actions Docker](https://docs.github.com/en/actions/using-containerized-services)

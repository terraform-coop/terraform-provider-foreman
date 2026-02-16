# Phase 1: API Client Modernization - Detailed Tasks

## Overview
Transform the custom HTTP API client into an automatically generated, type-safe client based on Foreman's API specification.

## Prerequisites
- [ ] Go 1.22+ installed
- [ ] Access to Foreman instance (for API spec extraction)
- [ ] Development environment configured
- [ ] Feature branch created: `feature/api-client-modernization`

## Task Breakdown

### Task 1.1: API Specification Extraction (Week 1, Days 1-2)

**Goal**: Extract Foreman's API specification in Apipie format

#### Subtasks:
1. [ ] Set up local Foreman instance or use test instance
   - Deploy using Docker: `docker run -d --name foreman theforeman/foreman:latest`
   - Or use existing test/staging environment
   - Verify access to `/apidoc` endpoint

2. [ ] Extract API documentation
   ```bash
   # Download Apipie JSON specification
   curl -u admin:password \
     http://foreman.example.com/apidoc/api.json \
     -o foreman-api-apipie.json
   ```

3. [ ] Analyze Apipie structure
   - [ ] Document resource types
   - [ ] Identify authentication patterns
   - [ ] Map API versions
   - [ ] List all endpoints (expect 300-500+)

4. [ ] Extract Katello plugin API
5. [ ] Extract Foreman Tasks API

**Deliverables**:
- Apipie JSON files for Foreman core, Katello, and plugins
- Documentation of API structure
- List of endpoints by resource type

**Success Criteria**:
- Complete API specification extracted
- All plugins included
- Documentation reviewed

---

### Task 1.2: Apipie to OpenAPI Converter (Week 1, Days 3-5)

**Goal**: Create tool to convert Apipie format to OpenAPI 3.0

See full details in the complete document.

---

### Task 1.3: Generate OpenAPI Specifications (Week 2, Days 1-2)

**Goal**: Generate OpenAPI specs for all Foreman APIs

---

### Task 1.4: Client Code Generation Setup (Week 2, Days 3-5)

**Goal**: Set up oapi-codegen and generate initial Go client

---

### Task 1.5: Create Adapter Layer (Week 3, Days 1-3)

**Goal**: Build compatibility layer between generated client and existing provider code

---

### Task 1.6: Migrate Sample Resources (Week 3, Day 4-5 & Week 4)

**Goal**: Migrate 3-5 simple resources to validate approach

---

### Task 1.7: Integration and Testing (Week 4)

**Goal**: Comprehensive validation of new API client

---

## Definition of Done

Phase 1 is complete when:

- [ ] Apipie to OpenAPI converter is functional and tested
- [ ] OpenAPI specifications generated for all Foreman APIs
- [ ] oapi-codegen successfully generates Go client code
- [ ] Adapter layer provides backward compatibility
- [ ] 3-5 resources migrated and tested
- [ ] All existing tests pass
- [ ] Documentation updated
- [ ] Code reviewed and approved
- [ ] Plan for remaining resource migration documented

## Tools and References

### Tools
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) - Go client generator
- [OpenAPI Validator](https://apitools.dev/swagger-parser/online/) - Validate specs
- [Swagger Editor](https://editor.swagger.io/) - Edit and visualize OpenAPI

### References
- [Foreman API Documentation](https://apidocs.theforeman.org/)
- [OpenAPI 3.0 Specification](https://swagger.io/specification/)
- [oapi-codegen Documentation](https://github.com/oapi-codegen/oapi-codegen/blob/master/README.md)

For detailed subtask breakdowns, see the full version in repository documentation.

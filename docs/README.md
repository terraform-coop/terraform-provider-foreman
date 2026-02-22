# Terraform Provider Foreman - Modernization Documentation

This directory contains comprehensive documentation for the modernization initiative of the terraform-provider-foreman project.

## 📚 Documentation Overview

### Primary Documents

1. **[MODERNIZATION_PLAN.md](../MODERNIZATION_PLAN.md)** - Executive summary and high-level strategy
   - Current state analysis
   - Modernization strategy for all three phases
   - Timeline and risk assessment
   - Success metrics

2. **[IMPLEMENTATION_ROADMAP.md](IMPLEMENTATION_ROADMAP.md)** - GitHub-ready issue templates
   - Epic issues for each phase
   - Detailed sub-task issues
   - Priority and sequencing
   - Project tracking guidelines

### Phase-Specific Task Documents

3. **[Phase 1: API Client Modernization](tasks/phase1-api-client-tasks.md)**
   - Detailed task breakdown for API client transformation
   - 7 major tasks with subtasks
   - Tools and references
   - 4-week timeline

4. **[Phase 2: Framework Migration](tasks/phase2-framework-migration-tasks.md)**
   - Detailed task breakdown for Terraform framework migration
   - Resource-by-resource migration plan
   - Testing strategy
   - 11-week timeline

5. **[Phase 3: E2E Testing](tasks/phase3-e2e-testing-tasks.md)**
   - Detailed task breakdown for E2E testing infrastructure
   - Docker setup and test framework
   - CI/CD integration
   - 5-week timeline

### Implementation Guides

6. **[Downloading API Specs from GitHub Actions](guides/downloading-api-specs.md)**
   - Step-by-step guide for obtaining Foreman API specifications
   - GitHub Actions artifact download instructions
   - Version mapping and organization
   - Alternative methods (running instance)

## 🎯 Quick Start Guide

### For Project Managers

Start here to understand the overall initiative:
1. Read [MODERNIZATION_PLAN.md](../MODERNIZATION_PLAN.md) - Executive Summary section
2. Review the Timeline in the plan (19 weeks total)
3. Check the Risk Assessment section
4. Use [IMPLEMENTATION_ROADMAP.md](IMPLEMENTATION_ROADMAP.md) to create GitHub issues

### For Developers

Start here to begin implementation:
1. Read relevant phase document based on your assigned work:
   - API work: [Phase 1 Tasks](tasks/phase1-api-client-tasks.md)
   - Framework work: [Phase 2 Tasks](tasks/phase2-framework-migration-tasks.md)
   - Testing work: [Phase 3 Tasks](tasks/phase3-e2e-testing-tasks.md)
2. Check prerequisites in each document
3. **For Phase 1**: Start with [Downloading API Specs Guide](guides/downloading-api-specs.md)
4. Follow the task breakdown step-by-step

### For Contributors

1. Review [MODERNIZATION_PLAN.md](../MODERNIZATION_PLAN.md) to understand the vision
2. Check open issues on GitHub (will be created from roadmap)
3. Pick an issue that matches your skills
4. Follow the detailed task guide for that phase

## 📋 Implementation Phases

### Phase 1: API Client Modernization (4 weeks)
**Goal**: Replace custom HTTP client with auto-generated code

**Key Outcomes**:
- Automated client generation from Foreman API specs
- Reduced maintenance burden
- Type-safe API interactions
- Foundation for future improvements

**Start with**: 
1. [Downloading API Specs Guide](guides/downloading-api-specs.md)
2. [Phase 1 Tasks](tasks/phase1-api-client-tasks.md)

---

### Phase 2: Terraform Framework Migration (11 weeks)
**Goal**: Migrate from Plugin SDK v2 to Plugin Framework

**Key Outcomes**:
- Modern, actively maintained framework
- Better type safety and error handling
- Access to new Terraform features
- Improved developer experience

**Start with**: [Phase 2 Tasks](tasks/phase2-framework-migration-tasks.md)

---

### Phase 3: E2E Testing Infrastructure (5 weeks)
**Goal**: Implement comprehensive testing against real Foreman

**Key Outcomes**:
- Tests against real Foreman instances
- CI/CD integration
- Multiple version support
- Confidence in releases

**Start with**: [Phase 3 Tasks](tasks/phase3-e2e-testing-tasks.md)

---

## 📊 Timeline Summary

```
Total Duration: 19 weeks (~4.5 months)

Week 1-4:   Phase 1 (API Client)
Week 5-15:  Phase 2 (Framework Migration)
Week 11-15: Phase 3 (E2E Testing) - Parallel with Phase 2
Week 16-19: Integration & Final Testing
```

## 🔗 Key Links

### Foreman Resources
- [Foreman API Documentation](https://apidocs.theforeman.org/)
- [Foreman GitHub](https://github.com/theforeman/foreman)
- [Foreman Community](https://community.theforeman.org/)

### Terraform Resources
- [Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Plugin Framework Migration Guide](https://developer.hashicorp.com/terraform/plugin/framework/migrating)
- [Plugin Testing](https://developer.hashicorp.com/terraform/plugin/framework/acctests)

### Tools
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) - Go API client generator
- [OpenAPI 3.0 Spec](https://swagger.io/specification/)
- [Docker Compose](https://docs.docker.com/compose/)

## 🎓 Learning Resources

### For API Client Work
- Understanding OpenAPI/Swagger specifications
- Go code generation patterns
- REST API best practices
- Foreman API structure

### For Framework Migration
- Terraform Plugin Framework documentation
- Protocol v6 provider architecture
- Resource lifecycle management
- State management in Terraform

### For E2E Testing
- Docker and containerization
- terraform-plugin-testing framework
- Integration testing patterns
- CI/CD with GitHub Actions

## 📝 Document Maintenance

These documents should be updated:
- **After each phase completes**: Update with lessons learned
- **When issues arise**: Document problems and solutions
- **When estimates change**: Adjust timelines
- **When scope changes**: Update task lists

## 🤝 Contributing

To contribute to the modernization effort:

1. **Review the documentation**
   - Understand the overall plan
   - Read the relevant phase documentation

2. **Check GitHub issues**
   - Look for open issues created from the roadmap
   - Comment on issues you're interested in

3. **Pick a task**
   - Start with clearly defined tasks
   - Ask questions if anything is unclear

4. **Submit PRs**
   - Follow the task checklist
   - Include tests
   - Update documentation

5. **Provide feedback**
   - Report blockers or issues
   - Suggest improvements
   - Share lessons learned

## 📞 Getting Help

- **Questions about the plan**: Open a discussion in GitHub Discussions
- **Technical issues**: Create an issue with `question` label
- **Contributions**: Follow the task guides and submit PRs
- **Feedback**: Open an issue with `feedback` label

## ✅ Next Steps

### Immediate Actions (Week 1)

1. **For Maintainers**:
   - [ ] Review and approve the modernization plan
   - [ ] Create GitHub issues from the roadmap
   - [ ] Set up project board for tracking
   - [ ] Allocate development resources
   - [ ] Announce initiative to community

2. **For Developers**:
   - [ ] Read Phase 1 documentation
   - [ ] Set up development environment
   - [ ] Test Docker setup locally
   - [ ] Prepare for API spec extraction

3. **For Community**:
   - [ ] Review the plan
   - [ ] Provide feedback
   - [ ] Volunteer for specific tasks
   - [ ] Join the discussion

### First Milestone: Phase 1 Proof of Concept (Week 2)

- Extract API specifications
- Build converter tool
- Generate first OpenAPI spec
- Demonstrate concept works

## 📅 Tracking Progress

Progress will be tracked through:
- GitHub Issues (one per task)
- GitHub Project Board (visual tracking)
- Regular status updates (weekly or bi-weekly)
- This documentation (updated with lessons learned)

## 🏆 Success Criteria

The modernization is successful when:

- [ ] API client is auto-generated
- [ ] All resources use Plugin Framework
- [ ] E2E tests run in CI
- [ ] Test coverage >80%
- [ ] No breaking changes for users
- [ ] Documentation complete
- [ ] Community satisfied

---

## Document History

- **2024-02-16**: Initial documentation created
  - Comprehensive plan covering all three phases
  - Detailed task breakdowns
  - Implementation roadmap with issue templates

---

**Questions?** Open an issue or start a discussion!

**Ready to contribute?** Check the [Implementation Roadmap](IMPLEMENTATION_ROADMAP.md) for tasks!

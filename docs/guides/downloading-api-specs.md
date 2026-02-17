# Downloading Foreman API Specifications from GitHub Actions

This guide explains how to download Foreman's API documentation (Apipie format) from GitHub Actions artifacts instead of running a Foreman instance.

## Why Use GitHub Actions Artifacts?

**Benefits:**
- ✅ No need to run a Foreman instance
- ✅ Version-specific and guaranteed accurate
- ✅ Easy to get multiple versions for compatibility testing
- ✅ Matches the exact released version

**vs. Running Instance:**
- ❌ Requires Docker or VM setup
- ❌ May not match production versions exactly
- ❌ Requires credentials and network access

## Quick Start

### For Foreman Core

1. **Visit the GitHub Actions page:**
   - Go to: https://github.com/theforeman/foreman/actions/workflows/foreman.yml
   
2. **Filter by version branch:**
   - For Foreman 3.18: https://github.com/theforeman/foreman/actions/workflows/foreman.yml?query=branch%3A3.18-stable
   - For Foreman 3.9: https://github.com/theforeman/foreman/actions/workflows/foreman.yml?query=branch%3A3.9-stable
   - For development: https://github.com/theforeman/foreman/actions/workflows/foreman.yml?query=branch%3Adevelop

3. **Download the artifact:**
   - Click on the latest successful (✓ green checkmark) workflow run
   - Scroll down to the "Artifacts" section
   - Download the `apidoc-*` artifact (there may be multiple, any one works)

4. **Extract and organize:**
   ```bash
   # Extract the downloaded artifact
   unzip ~/Downloads/apidoc-3-18-stable.zip -d /tmp/
   
   # Create api-specs directory if needed
   mkdir -p api-specs
   
   # Move to organized location
   mv /tmp/apidoc*.json api-specs/foreman-core-3.18-apipie.json
   ```

### For Katello Plugin

1. **Visit Katello GitHub Actions:**
   - Go to: https://github.com/Katello/katello/actions
   
2. **Find the appropriate workflow:**
   - Look for workflows that generate API documentation
   - Filter by branch: `KATELLO-4.x` (matches Foreman versions)

3. **Download and extract:**
   ```bash
   unzip ~/Downloads/katello-apidoc-*.zip -d /tmp/
   mv /tmp/katello-apidoc*.json api-specs/katello-4.13-apipie.json
   ```

## Version Mapping

| Foreman Version | GitHub Branch | Katello Version |
|-----------------|---------------|-----------------|
| 3.18            | 3.18-stable   | 4.13            |
| 3.17            | 3.17-stable   | 4.12            |
| 3.16            | 3.16-stable   | 4.11            |
| 3.9             | 3.9-stable    | 4.9             |
| Development     | develop       | master          |

## Using the Helper Script

We provide a helper script to guide you through the process:

```bash
./scripts/download-api-specs.sh
```

## Advanced: Using GitHub CLI

If you have the GitHub CLI (`gh`) installed and authenticated:

```bash
# Install GitHub CLI: https://cli.github.com/
gh auth login

# List recent workflow runs
gh run list --repo theforeman/foreman --workflow=foreman.yml --branch=3.18-stable --limit 5

# Download specific run artifacts
gh run download <run-id> --repo theforeman/foreman --dir api-specs/
```

## Alternative: Extract from Running Instance

If GitHub Actions artifacts aren't available:

```bash
# Using Docker
docker run -d --name foreman -p 3000:3000 theforeman/foreman:3.18

# Extract API docs
curl -u admin:changeme \
  http://localhost:3000/apidoc/api.json \
  -o api-specs/foreman-core-3.18-apipie.json
```

## Next Steps

After downloading specifications, proceed to [Task 1.2: Apipie to OpenAPI Converter](../tasks/phase1-api-client-tasks.md).

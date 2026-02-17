#!/bin/bash
# Script to help download Foreman API specifications from GitHub Actions artifacts
#
# Usage:
#   ./scripts/download-api-specs.sh
#
# This script provides guidance on downloading API specifications from GitHub Actions.
# Due to GitHub's authentication requirements for downloading artifacts, manual steps are needed.

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Foreman API Specification Download Guide${NC}"
echo "=========================================="
echo ""

# Create api-specs directory if it doesn't exist
mkdir -p api-specs

echo "This script helps you download Foreman API specifications from GitHub Actions."
echo ""
echo -e "${YELLOW}Note: GitHub Actions artifacts require authentication to download.${NC}"
echo "You'll need to manually download them through the GitHub web interface."
echo ""

# Function to display instructions for a component
show_instructions() {
    local component=$1
    local repo=$2
    local workflow=$3
    local branch=$4
    local output_file=$5

    echo -e "${GREEN}=== $component ===${NC}"
    echo ""
    echo "1. Visit: https://github.com/$repo/actions/workflows/$workflow?query=branch%3A$branch"
    echo ""
    echo "2. Click on the latest successful (green checkmark) workflow run"
    echo ""
    echo "3. Scroll down to 'Artifacts' section"
    echo ""
    echo "4. Download the 'apidoc-*' artifact"
    echo ""
    echo "5. Extract the downloaded .zip file"
    echo ""
    echo "6. Move the JSON file to: ./api-specs/$output_file"
    echo ""
    echo "Example commands after download:"
    echo "  unzip ~/Downloads/apidoc-*.zip -d /tmp/"
    echo "  mv /tmp/apidoc*.json ./api-specs/$output_file"
    echo ""
}

# Get version from user
read -p "Enter Foreman version (e.g., 3.18, 3.9, latest): " VERSION

if [ "$VERSION" = "latest" ]; then
    BRANCH="develop"
    VERSION_SUFFIX="latest"
else
    BRANCH="${VERSION}-stable"
    VERSION_SUFFIX="$VERSION"
fi

echo ""
echo "Downloading specifications for Foreman version: $VERSION (branch: $BRANCH)"
echo ""

# Foreman Core
show_instructions \
    "Foreman Core" \
    "theforeman/foreman" \
    "foreman.yml" \
    "$BRANCH" \
    "foreman-core-${VERSION_SUFFIX}-apipie.json"

# Katello (if applicable)
if [ "$VERSION" != "latest" ] && [ "${VERSION%%.*}" -ge "3" ]; then
    echo ""
    show_instructions \
        "Katello Plugin" \
        "Katello/katello" \
        "katello.yml" \
        "KATELLO-${VERSION}" \
        "katello-${VERSION_SUFFIX}-apipie.json"
fi

echo ""
echo -e "${GREEN}After downloading all specifications:${NC}"
echo ""
echo "1. Verify the files are in the api-specs/ directory:"
echo "   ls -lh api-specs/"
echo ""
echo "2. Proceed to Task 1.2: Apipie to OpenAPI Converter"
echo ""
echo -e "${YELLOW}Tip:${NC} You can download multiple versions for compatibility testing."
echo ""

# Alternative: Using curl with GitHub token (for advanced users)
echo ""
echo -e "${GREEN}Advanced: Using GitHub CLI (gh)${NC}"
echo "If you have GitHub CLI installed and authenticated:"
echo ""
echo "# Install gh: https://cli.github.com/"
echo "# Authenticate: gh auth login"
echo ""
echo "# Then you can download artifacts programmatically:"
echo "gh run list --repo theforeman/foreman --workflow=foreman.yml --branch=$BRANCH --limit 1"
echo "# Get the run ID from above, then:"
echo "# gh run download <run-id> --repo theforeman/foreman --name apidoc-3-18-stable"
echo ""

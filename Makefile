# ------------------------------------------------------------------------------
# Makefile definition for the Foreman Terraform provider project
# ------------------------------------------------------------------------------

# ------------------------------------------------------------------------------
# Makefile Macros
# ------------------------------------------------------------------------------

# GOPATH and GOBIN environment variable capture and overrides for use in the
# Makefile
GO_BIN := $(GOPATH)/bin
VERSION:=$(CI_COMMIT_TAG)

# List of files for use in the verious go commands. GOFMT_FILES is used when
# running format checks and formatting the codebase with "go fmt", GOVET_FILES
# contains the package names to issue "go vet" against, and GOTEST_FILES lists
# the packages used for testing with "go test".
#
# Files that are part of the vendor directory are not included as part of the
# format check, vetting, and testing.
GOFMT_FILES := $(shell find . -name '*.go' | grep -v vendor)
GOVET_FILES := $(shell go list ./... | grep -v vendor)
GOTEST_FILES := $(GOVET_FILES)
GODOC_FILES := $(GOVET_FILES)

# Root directory for (auto)generated project documentation
#
# NOTE(ALL): DOCS_DIR should be kept in sync with docs_dir in mkdocs.yml
DOCS_DIR := docs
# Directory to output static HTML generated from the `godoc` tool
GODOC_OUT_DIR := $(DOCS_DIR)/godoc
# wget options.  wget is used in the 'doc' target to generate static site
# documentation for the project.
#
# -r, --recursive
#   Turn on recusrive retrieving. The default maximum depth is 15
# -np, --no-parent
#   Disallow the retrieval of the links that refer to the hierarchy above the
#   beginning directory
# -nH, --no-host-directories
#   Disable generation of host-prefixed directories. By default, with the "-r"
#   option, will create a structure of directories beginning with the hostname.
#   This disables this behavior
# -nv, --no-verbose
#   Turn off verbose without being completely quiet, error messages and basic
#   information get printed
# -N, --timestamping
#   Turn on timestamping (only download files that do not already exist
#   locally or the remote has a newer version)
# -E, --adjust-extension
#   Appends the correct file suffix to the local filename if the downloaded
#   file does not already have it
# -p, --page-requisites
#   Download all the files that are necessary to properly display a given
#   HTML page. This includes things like inlined images, sounds, and
#   referenced stylesheets
# -k, --convert-links
#   After the download is complete, convert the links in the document to make
#   them suitable for local viewing
# -e, --execute
#   Execute a command as if it were part of .wgetrc. The commands are AFTER
#   comamnds in wgetrc, thus taking precedence
# -P, --directory-prefix
#   Set the directory prefix. All files and sub-directories will be saved to
#   this location to form the top of the retreival tree
WGET_OPTIONS := -r -np -nH -nv -N -E -p -k -e 'robots=off' -P "$(GODOC_OUT_DIR)"

# Options to pass to the autogodoc tool
AUTODOC_TARGET := autodoc$(TARGET_EXT)
AUTODOC_OPTIONS := -docs-dir="$(DOCS_DIR)"

ifndef VERSION
	VERSION:=$(shell git describe --always 2>/dev/null)
endif

# Path to the root of the terraform configuration directory and the path to the
# root of the third-party plugins directory.  From the Terraform documentation:
#
#   Third-party providers can be manually installed by placing their plugin
#   executables in one of the following locations depending on the host
#   operating system:
#
#     * On Windows, in the sub-path terraform.d/plugins beneath your user's
#       "Application Data" directory
#     * On all other systems, in the sub-path .terraform.d/plugins in your
#       user's home directory
HOST_OS := $(shell go env GOHOSTOS)
ifeq ($(strip $(HOST_OS)),windows)
	TERRAFORM_D := ~/AppData/Roaming/terraform.d
else
	TERRAFORM_D := ~/.terraform.d
endif
TERRAFORM_PLUGINS := $(TERRAFORM_D)/plugins

# Target binary name.  The target is the name of the repository + the binary
# file extension.  The file extension is retrieved from the "go env" command.
TARGET_NAME := $(shell basename "${PWD}")
TARGET_EXT := $(shell go env GOEXE)
TARGET := $(TARGET_NAME)_$(VERSION)_x4$(TARGET_EXT)

# Output directory - the binary will be placed in this location if the user
# invokes the 'build' target to put the binary on the local machine
OUT_DIR := build/$(GOOS)_$(GOARCH)/$(subst terraform-provider-,,$(TARGET_NAME))

# ------------------------------------------------------------------------------
# Makefile Targets
# ------------------------------------------------------------------------------

# All of the Makefile targets are not the names of files and therefore are
# phony targets
.PHONY: all build build-autodoc clean clean-godoc clean-mkdoc default ensure format formatcheck godoc install mkdocs test vet

# Default target - build the project
# Use the special built-in target name and human conventions
.DEFAULT: build
default: build
all: build

# Compiles the codebase into the target binary.  The binary will be in the
# output directory
build: formatcheck
	@echo "Compiling codebase to $(TARGET) Platform:$(GOOS) Arch:$(GOARCH)"
	@go build -v -o $(OUT_DIR)/$(TARGET)

# Compiles the autodoc into an executabvle. The executable will be in the
# output directory and can be invoked from the command line to generate
# mkdocs documentation.
build-autodoc: formatcheck
	@echo "Compiling codebase to $(OUT_DIR)/$(AUTODOC_TARGET) Platform:$(GOOS) Arch:$(GOARCH)"
	@go build -v -o $(OUT_DIR)/$(AUTODOC_TARGET) $$(go list ./cmd/autodoc)

# Removes the compiled binaries (if they exist), log files, and documentation
clean: clean-godoc clean-mkdoc
	@echo 'Cleaning binaries...'
	@rm -rf "$(OUT_DIR)" 2>/dev/null || true
	@rm "$(GO_BIN)/$(TARGET)" 2>/dev/null || true
	@rm "$(GO_BIN)/$(AUTODOC_TARGET)" 2>/dev/null || true
	@rm "$(TERRAFORM_PLUGINS)/$(TARGET_NAME)" 2>/dev/null || true
	@echo 'Cleaning log files...'
	@find . -type f -name '*.log' -delete 2>/dev/null || true

# Removes all godoc files
clean-godoc:
	@echo 'Cleaning godoc files...'
	@rm -rf $(DOCS_DIR)/godoc 2>/dev/null || true

# Removes all mkdocs files
clean-mkdoc:
	@echo 'Cleaning mkdocs files...'
	@rm mkdocs.yml 2>/dev/null || true
	@find "$(DOCS_DIR)" -type f -name '*.md' -delete 2>/dev/null || true

# Ensure the project dependencies are in sync and up-to-date.  This will read
# the dependencies and constraints in the Gopkg.toml file and update the
# /vendor directory and Gopkg.lock file to reflect the constraints.
ensure:
	@echo 'Ensuring project dependencies are up to date...'
	@dep ensure

# Runs "go fmt" on the codebase and writes the output back to the source files
format:
	@echo 'Formatting codebase...'
	@gofmt -w $(GOFMT_FILES)

# Runs "go fmt" on the codebase, but unlike the "format" target it does not
# write the results back to the source files.  It captures the output of the
# files that violate the formatting and displays them to the console.
formatcheck:
	@echo 'Validating format of codebase...'
	@badFiles=$$(gofmt -l $(GOFMT_FILES)); \
	if [ -n "$$badFiles" ]; \
	then \
		echo 'The following files violate go formatting:'; \
		echo ''; \
		echo "$$badFiles"; \
		echo ''; \
		echo 'Run "make format" to reformat the code.'; \
		exit 1; \
	else \
		echo 'All files pass format check.'; \
		exit 0; \
	fi

# Generates godoc for the project and saves the static assets to GODOC_OUT_DIR
# through recursive downloads with wget.  The godoc can be read locally through
# a web viewport by browsing the filesystem.  The documentation is also used in
# conjunction with the documentation stage of the gitlab pipeline for creating
# project documentation.
godoc:
	@echo "Generating godoc to $(GODOC_OUT_DIR)..."
	@mkdir -p "$(GODOC_OUT_DIR)"
	@pkgRoot=$$(go list .); \
	godocAddr="127.0.0.1:8000"; \
	godoc -http="$${godocAddr}" & \
	godocPID="$$!"; \
	echo "godoc PID: [$${godocPID}]"; \
	echo "Sleeping while godoc initializes..."; \
	sleep 5; \
	echo "Downloading pages..."; \
	echo ''; \
	wget $(WGET_OPTIONS) "http://$${godocAddr}/pkg/$${pkgRoot}"; \
	echo ''; \
	echo 'done.'; \
	echo "Killing godoc process [$${godocPID}]"; \
	kill "$${godocPID}";

# Compiles the codebase and moves the target binary into the terraform plugins
# directory for use
install: $(OUT_DIR)/$(TARGET)
	@echo "Creating plugins directory $(TERRAFORM_PLUGINS)"
	@mkdir -p $(TERRAFORM_PLUGINS)
	@echo "Moving $(TARGET) to terraform.d/plugins..."
	@mv $(OUT_DIR)/$(TARGET) $(TERRAFORM_PLUGINS)/$(TARGET_NAME)$(TARGET_EXT)

# Uses the autodoc tool to generate project mkdocs documentation
mkdocs: $(OUT_DIR)/$(AUTODOC_TARGET)
	@echo "Generating mkdocs documentation..."
	@if [ ! -d "$(DOCS_DIR)" ]; then \
		echo "Creating $(DOCS_DIR)"; \
		mkdir -p "$(DOCS_DIR)"; \
	fi; \
	if [ ! -d "$(DOCS_DIR)/datasources" ]; then \
		echo "Creating $(DOCS_DIR)/datasources"; \
		mkdir -p "$(DOCS_DIR)/datasources"; \
	fi; \
	if [ ! -d "$(DOCS_DIR)/resources" ]; then \
		echo "Creating $(DOCS_DIR)/resources"; \
		mkdir -p "$(DOCS_DIR)/resources"; \
	fi; \
	./$(OUT_DIR)/$(AUTODOC_TARGET) $(AUTODOC_OPTIONS)

# Runs the go unit and integration tests on the codebase
test:
	@echo 'Running unit tests...'
	@go test $(GOTEST_FILES)

# Runs "go vet" on the codebase and writes any errors or suspicious program
# behavior to the console
vet:
	@echo "Vetting the codebase for suspicious constructs..."
	@vetOutput=$$(go vet $(GOVET_FILES) 2>&1); \
	exitStatus=$$?; \
	if [ "$$exitStatus" -eq 0 ]; \
	then \
		echo 'All files pass vet check'; \
		exit 0; \
	else \
		echo 'Codebase failed vet check:'; \
		echo ''; \
		echo "$$vetOutput"; \
		echo ''; \
		exit 1; \
	fi

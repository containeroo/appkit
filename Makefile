# Makefile

.DEFAULT_GOAL := help

## Tool Versions
# renovate: datasource=github-releases depName=gi8lino/dev-tools
DEV_TOOLS_VERSION ?= v0.9.0

# renovate: datasource=github-releases depName=kumbuka-me/cli
KUMBUKA_CLI_VERSION ?= v0.11.11

## Documentation
DOCS_DIR := docs
SITE_CONFIG ?= $(DOCS_DIR)/site.toml
SITE_OUTPUT := $(DOCS_DIR)/site
SITE_PORT ?= $(call dev-port,site)
# Use recursive expansion: dev-port is supplied by an included dev-tools module
# further down in this file.
SITE_URL := http://127.0.0.1:$(SITE_PORT)/
SITE_PLUGINS := $(DOCS_DIR)/.kumbukaplugins
PYTHON ?= python3

FAVICON_GENERATE := $(DEV_TOOLS_BIN)/favicon-generate
SVG_TO_PNG := $(DEV_TOOLS_BIN)/svg-to-png

## Project-local tools
# Keep the release version in the target path so a version bump cannot reuse
# an older downloaded binary merely because it already exists on disk.
KUMBUKA_CLI := bin/kumbuka-cli-$(KUMBUKA_CLI_VERSION)
KUMBUKA_CLI_ASSET ?= kumbuka-cli_{version}_{os}_{arch}.tar.gz

include bin/dev-tools.mk
include $(call dev-tools-module,tag)
include $(call dev-tools-module,port)
include $(call dev-tools-module,browser)
include $(call dev-tools-module,help)
include	$(call dev-tools,favicon-generate)
include	$(call dev-tools,svg-to-png)

##@ Documentation

.PHONY: site
site: kumbuka ## Build the documentation site.
	$(KUMBUKA_CLI) plugins sync --file "$(SITE_PLUGINS)"
	$(KUMBUKA_CLI) build --config "$(SITE_CONFIG)" --plugins "$(SITE_PLUGINS)"

.PHONY: site-serve
site-serve: kumbuka $(DEV_PORT) $(OPEN_BROWSER) ## Build, serve, and open the documentation site locally.
	$(KUMBUKA_CLI) plugins sync --file "$(SITE_PLUGINS)"
	$(KUMBUKA_CLI) build --config "$(SITE_CONFIG)" --plugins "$(SITE_PLUGINS)" --site-url "$(SITE_URL)"
	@echo "Serving appkit documentation at $(SITE_URL)"
	@$(OPEN_BROWSER) "$(SITE_URL)" & \
	$(PYTHON) -m http.server "$(SITE_PORT)" --bind 127.0.0.1 --directory "$(SITE_OUTPUT)"

.PHONY: clean-site
clean-site: ## Remove generated documentation output.
	rm -rf "$(SITE_OUTPUT)"

##@ Dependencies

.PHONY: kumbuka
kumbuka: $(KUMBUKA_CLI) ## Install the pinned Kumbuka static-site generator.

$(KUMBUKA_CLI): $(GITHUB_RELEASE_INSTALL)
	@$(GITHUB_RELEASE_INSTALL) \
		--repo kumbuka-me/cli \
		--tag "$(KUMBUKA_CLI_VERSION)" \
		--asset "$(KUMBUKA_CLI_ASSET)" \
		--binary kumbuka-cli \
		--target "$(KUMBUKA_CLI)"

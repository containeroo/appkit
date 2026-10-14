# Makefile

.DEFAULT_GOAL := help

LOCALBIN ?= bin

$(LOCALBIN):
	@mkdir -p "$@"

## Tool Versions
# renovate: datasource=github-releases depName=gi8lino/dev-tools
DEV_TOOLS_VERSION ?= v0.9.0

include bin/dev-tools.mk
include $(call dev-tools-module,tag)
include $(call dev-tools-module,help)

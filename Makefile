GITHUB_OS_LIST := ubuntu-latest

GO_COVERAGE_PKGS := $(shell go list ./... | grep -v -P '/(test|fake)' | tr '\n' ',' | sed 's/,$$//')

include ./.bootstrap/makesystem.mk

ifeq ($(MAKESYSTEM_FOUND),1)
include $(MAKESYSTEM_BASE_DIR)/go.mk
endif

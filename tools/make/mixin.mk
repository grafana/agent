# Makefile.mixin adds mixin-specific targets.

PARENT_MAKEFILE := $(firstword $(MAKEFILE_LIST))

MIXIN_PATH       ?= operations/agent-flow-mixin
MIXIN_OUT_PATH   ?= operations/agent-flow-mixin-compiled

format-mixin: ## Format the mixin files.
	@find $(MIXIN_PATH) -type f -name '*.libsonnet' | xargs jsonnetfmt -i

check-mixin: ## Build, format and check the mixin files.
check-mixin: build-mixin format-mixin check-mixin-jb check-mixin-mixtool
	@cd $(MIXIN_PATH) && \
	jb install && \
	mixtool lint mixin.libsonnet

check-mixin-jb:
	@cd $(MIXIN_PATH) && \
	jb install

check-mixin-mixtool: check-mixin-jb
	@cd $(MIXIN_PATH) && \
	mixtool lint mixin.libsonnet

build-mixin: ## Generates the agent-flow-mixin zip file.
build-mixin: check-mixin-jb
		mkdir -p "$(MIXIN_OUT_PATH)"; \
		mixtool generate all --output-alerts "$(MIXIN_OUT_PATH)/alerts.yaml" --directory "$(MIXIN_OUT_PATH)/dashboards" "${MIXIN_PATH}/mixin.libsonnet"; \
		cd "$(MIXIN_OUT_PATH)/.." && zip -q -r "agent-flow-mixin.zip" $$(basename "$(MIXIN_OUT_PATH)"); \
		cd -; \
		echo "The mixin has been compiled to $(MIXIN_OUT_PATH) and archived to $$(realpath --relative-to=$$(pwd) $(MIXIN_OUT_PATH)/../agent-flow-mixin.zip)"; \


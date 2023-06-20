PODMAN = $(shell if command -v podman >/dev/null 2>&1; then echo podman; else echo docker; fi)
IMAGE = grafana/docs-base:latest
CONTENT_PATH = /hugo/content/docs/agent/latest
PORT = 3002:3002

.PHONY: pull
pull:
	$(PODMAN) pull $(IMAGE)

.PHONY: docs
docs: check-cloudwatch-integration pull
	$(PODMAN) run --init -v $(shell pwd)/sources:$(CONTENT_PATH):Z -p $(PORT) --rm -it $(IMAGE)

check-cloudwatch-integration:
	$(PODMAN) run -v $(shell git rev-parse --show-toplevel):/repo -v $(shell pwd):/docs -w /repo golang:1.19.4-bullseye go run pkg/integrations/cloudwatch_exporter/docs/doc.go check /docs/sources/static/configuration/integrations/cloudwatch-exporter-config.md

generate-cloudwatch-integration:
	$(PODMAN) run -v $(shell git rev-parse --show-toplevel):/repo -v $(shell pwd):/docs -w /repo golang:1.19.4-bullseye go run pkg/integrations/cloudwatch_exporter/docs/doc.go generate

sources/assets/hierarchy.svg: sources/operator/hierarchy.dot
	cat $< | $(PODMAN) run --rm -i nshine/dot dot -Tsvg > $@

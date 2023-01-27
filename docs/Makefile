PODMAN = $(shell if command -v podman &>/dev/null; then echo podman; else echo docker; fi)
IMAGE = grafana/docs-base:latest
CONTENT_PATH = /hugo/content/docs/agent/latest
PORT = 3002:3002

.PHONY: pull
pull:
	$(PODMAN) pull $(IMAGE)

.PHONY: docs
docs: pull
	$(PODMAN) run --init -v $(shell pwd)/sources:$(CONTENT_PATH):Z -p $(PORT) --rm -it $(IMAGE)

sources/assets/hierarchy.svg: sources/operator/hierarchy.dot
	cat $< | $(PODMAN) run --rm -i nshine/dot dot -Tsvg > $@

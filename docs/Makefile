IMAGE = grafana/docs-base:latest
CONTENT_PATH = /hugo/content/docs/agent/latest
PORT = 3002:3002

.PHONY: pull
pull:
	docker pull $(IMAGE)

.PHONY: docs
docs: check-cloudwatch-integration pull
	docker run --init -v $(shell pwd)/sources:$(CONTENT_PATH):Z -p $(PORT) --rm -it $(IMAGE)

check-cloudwatch-integration:
	docker run -v $(shell git rev-parse --show-toplevel):/repo -v $(shell pwd):/docs -w /repo golang:1.19.4-bullseye go run pkg/integrations/cloudwatch_exporter/docs/doc.go check /docs/sources/configuration/integrations/cloudwatch-exporter-config.md

sources/assets/hierarchy.svg: sources/operator/hierarchy.dot
	cat $< | docker run --rm -i nshine/dot dot -Tsvg > $@

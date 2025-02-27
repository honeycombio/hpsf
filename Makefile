GOCMD = go
GOTESTCMD = $(if $(shell command -v gotestsum),gotestsum --junitfile ./test_results/$(1).xml --format testname --,go test)

.PHONY: test
#: run all tests
test: test_with_race test_all

.PHONY: test_with_race
#: run only tests tagged with potential race conditions
test_with_race: test_results
	@echo
	@echo "+++ testing - race conditions?"
	@echo
	$(call GOTESTCMD,$@) -tags race --race --timeout 60s -v ./...

.PHONY: test_all
#: run all tests, but with no race condition detection
test_all: test_results
	@echo
	@echo "+++ testing - all the tests"
	@echo
	$(call GOTESTCMD,$@) -tags all --timeout 60s -v ./...

test_results:
	@mkdir -p test_results

tidy:
	$(GOCMD) mod tidy

TEMPLATE ?= pkg/config/templates/emathroughput
.PHONY: test_template
#: generate config from template (usage: make test_template TEMPLATE=pkg/config/templates/proxy)
test_template:
	@echo
	@echo "+++ generating config from template $(TEMPLATE)"
	@echo
	./testTemplate.sh $(TEMPLATE)

CONFIG ?= examples/hpsf.yaml
.PHONY: validate
#: validate provided config (usage: make validate CONFIG=examples/hpsf2.yaml)
validate:
	@echo
	@echo "+++ validating config $(CONFIG)"
	@echo
	go run ./cmd/hpsf -i $(CONFIG) validate
	for format in rConfig rRules cConfig ; do \
		echo; \
		echo "+++ validating config generation for $${format} with config $(CONFIG)"; \
		echo; \
		go run ./cmd/hpsf -i $(CONFIG) $${format} || exit 1; \
	done

.PHONY: validate_all
validate_all: examples/*
	for file in $^ ; do \
		$(MAKE) validate CONFIG=$${file} || exit 1; \
	done

.PHONY: smoke
smoke:
	@echo
	@echo "+++ basic smoke test - does it start?"
	@echo "+++ if so, make unsmoke after this"
	@echo
	mkdir -p tmp
	go run ./cmd/hpsf -i ./examples/hpsfProxy.yaml -o tmp/hpsfProxy.cconfig.yaml cConfig
	docker run --rm -d \
		--name smoke-proxy \
    -p 4227-4228:4227-4228 \
		-v ./tmp/hpsfProxy.cconfig.yaml:/etc/otelcol/config.yaml \
		otel/opentelemetry-collector:latest

.PHONY: unsmoke
unsmoke:
	@echo
	@echo "+++ stopping smoke test"
	@echo
	docker stop smoke-proxy

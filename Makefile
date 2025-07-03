GOCMD = go
GOTESTCMD = $(if $(shell command -v gotestsum),gotestsum --junitfile ./test_results/$(1).xml --format testname --,go test)

.PHONY: test
#: run all tests
test: test_with_race test_all test_scenarios

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

.PHONY: test_scenarios
#: run all tests in tests/scenario_tests
test_scenarios: test_results
	@echo
	@echo "+++ testing - scenario tests"
	@echo
	cd tests/scenario_tests && $(call GOTESTCMD,$@) -tags all --timeout 60s -v ./...

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
validate_all: examples/hpsf* pkg/data/templates/*
	for file in $^ ; do \
		$(MAKE) validate CONFIG=$${file} || exit 1; \
	done

.PHONY: .smoke_refinery
#: run smoke test for refinery component
#: Do not use directly, use the smoke target instead
.smoke_refinery:
	if [ -z "$(FILE)" ]; then \
		echo "+++ no component file provided, use smoke instead -- exiting"; \
		exit 1; \
	fi

	@echo generating refinery configs for component $(FILE)
	mkdir -p tmp

	# generate the configs from the provided file
	go run ./cmd/hpsf -i ${FILE} -o tmp/refinery-rules.yaml rRules || exit 1
	go run ./cmd/hpsf -i ${FILE} -o tmp/refinery-config.yaml rConfig || exit 1

	# run refinery with the generated configs
	docker run -d --name smoke-refinery \
		-v ./tmp/refinery-config.yaml:/etc/refinery/refinery.yaml \
		-v ./tmp/refinery-rules.yaml:/etc/refinery/rules.yaml \
		-e HTP_EXPORTER_APIKEY=hccik_01jj2jj42424jjjjjjj2jjjjjj424jjj2jjjjjjjjjjjjjjj4jjjjj24jj \
		honeycombio/refinery:latest || exit 1
	sleep 1

	# check if the container is running
	if [ "$$(docker inspect -f '{{.State.Running}}' 'smoke-refinery')" != "true" ]; then \
		echo "+++ container not running"; \
		docker logs 'smoke-refinery'; \
		docker rm 'smoke-refinery'; \
		exit 1; \
	else \
		echo "+++ container is running"; \
		docker kill 'smoke-refinery'; \
		docker rm 'smoke-refinery'; \
		echo "+++ refinery successfully started up for $(FILE)"; \
	fi

.PHONY: .smoke_collector
#: run smoke test for collector components
#: Do not use directly, use the smoke target instead
.smoke_collector:
	if [ -z "$(FILE)" ]; then \
		echo "+++ no component file provided, use smoke instead -- exiting"; \
		exit 1; \
	fi

	@echo generating collector configs for component $(FILE)
	mkdir -p tmp

	# generate the configs from the provided file
	go run ./cmd/hpsf -i ${FILE} -o tmp/collector-config.yaml cConfig || exit 1

	# check yq is installed and at least version 4.0.0
	if ! command -v yq &> /dev/null; then \
		echo "+++ yq could not be found, please install it"; \
		exit 1; \
	fi
	if [ "$$(yq --version | cut -d' ' -f2 | cut -d'.' -f1)" -lt 4 ]; then \
		echo "+++ yq version is less than 4.0.0, please update it"; \
		exit 1; \
	fi

	# use yq to remove the usage processor and honeycomb extension from collector config
	yq -i e \
		'del(.processors.usage) | \
		 del(.extensions.honeycomb) | \
		 del(.service.extensions[] | select(. == "honeycomb")) | \
		 del(.service.pipelines.traces.processors[] | select(. == "usage")) | \
		 del(.service.pipelines.metrics.processors[] | select(. == "usage")) | \
		 del(.service.pipelines.logs.processors[] | select(. == "usage"))' \
		tmp/collector-config.yaml || exit 1

	# run collector with the generated config
	docker run -d --name smoke-collector \
		--entrypoint /otelcol-contrib \
		-v ./tmp/collector-config.yaml:/etc/otelcol-contrib/config.yaml \
		-e HTP_COLLECTOR_POD_IP=localhost \
		-e HTP_REFINERY_POD_IP=localhost \
		honeycombio/supervised-collector:latest \
		--config /etc/otelcol-contrib/config.yaml || exit 1
	sleep 1

	# check if the container is running
	if [ "$$(docker inspect -f '{{.State.Running}}' 'smoke-collector')" != "true" ]; then \
		echo "+++ container not running"; \
		docker logs 'smoke-collector'; \
		docker rm 'smoke-collector'; \
		exit 1; \
	else \
		echo "+++ container is running"; \
		docker kill 'smoke-collector'; \
		docker rm 'smoke-collector'; \
		echo "+++ collector successfully started up for $(FILE)"; \
	fi

.PHONY: smoke
#: run smoke tests for HPSF templates
smoke: pkg/data/templates/*.yaml
	for file in $^ ; do \
		$(MAKE) .smoke_refinery FILE=$${file} || exit 1; \
		$(MAKE) .smoke_collector FILE=$${file} || exit 1; \
	done

.PHONY: unsmoke
unsmoke:
	@echo
	@echo "+++ stopping smoke test"
	@echo
	docker stop smoke-proxy

.PHONY: regenerate_translator_testdata
regenerate_translator_testdata:
	@echo
	@echo "+++ regenerating translator testdata"
	@echo
	OVERWRITE_TESTDATA=1 go test ./pkg/translator/

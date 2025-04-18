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
	go run ./cmd/hpsf -i ${FILE} -o tmp/refinery-rules.yaml rRules
	go run ./cmd/hpsf -i ${FILE} -o tmp/refinery-config.yaml rConfig
	
	# run refinery with the generated configs
	docker run -d --rm --name smoke-refinery \
		-v ./tmp/refinery-config.yaml:/etc/refinery/refinery.yaml \
		-v ./tmp/refinery-rules.yaml:/etc/refinery/rules.yaml \
		honeycombio/refinery:latest
	sleep 1

	# check if the container is running
	if [ "$$(docker inspect -f '{{.State.Running}}' 'smoke-refinery')" != "true" ]; then \
		echo "+++ container not running"; \
		exit 1; \
	else \
		echo "+++ container is running"; \
		docker kill 'smoke-refinery' > /dev/null; \
	fi

.PHONY: smoke
#: run smoke tests for HPSF components
smoke: pkg/data/components/*.yaml
	for file in $^ ; do \
		if [ "$$(yq '.templates[] | select(.kind | contains("refinery_config","refinery_rules"))' $${file})" != "" ]; then \
			$(MAKE) .smoke_refinery FILE=$${file}; \
		fi; \
	done

.PHONY: unsmoke
unsmoke:
	@echo
	@echo "+++ stopping smoke test"
	@echo
	docker stop smoke-proxy

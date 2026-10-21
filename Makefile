GO ?= go
GOFLAGS ?= -buildvcs=false
.PHONY: build test vet verify gates oracle rejection explain clean demo oracle-compact

build:
	mkdir -p bin
	$(GO) build $(GOFLAGS) -o bin/dlc ./src/cmd/dlc
	$(GO) build $(GOFLAGS) -o bin/conecheck ./tools/conecheck

vet:
	$(GO) vet ./...

test:
	$(GO) test -count=1 ./...
	python3 -m unittest discover -s harness -p 'test_*.py'

verify: build vet test

gates: verify rejection

oracle: build
	python3 harness/verify_pass1.py --require-oracle $(if $(ORACLE_OUTPUT),--output $(ORACLE_OUTPUT),)
	$(GO) test -count=1 -run TestSouffleDifferential ./tests/integration

rejection: build
	@for f in experiments/explain-samples/rejection_*.dl; do \
		if bin/dlc explain "$$f"; then echo "ERROR: accepted $$f"; exit 1; fi; \
	done

explain: build
	bin/dlc explain "$(FILE)"

clean:
	rm -rf bin/

demo: build
	$(GO) build $(GOFLAGS) -o bin/dlc-demo ./src/cmd/dlc-demo
	bin/dlc-demo $(DEMO_ARGS)

oracle-compact: build
	python3 harness/verify_compact.py --require-oracle $(if $(COMPACT_OUTPUT),--output $(COMPACT_OUTPUT),)

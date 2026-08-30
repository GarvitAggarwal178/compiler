.PHONY: build test vet gates rejection explain clean

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

# The four rejection grounds, demonstrated live (dlc explain), one sample
# program per ground -- see experiments/explain-samples/.
rejection:
	@for f in experiments/explain-samples/rejection_*.dl; do \
		echo "=== $$f ==="; \
		go run ./src/cmd/dlc explain "$$f" || true; \
	done

explain:
	go run ./src/cmd/dlc explain $(FILE)

# Everything Phase 7 of docs/restructure-notes.md checks before calling the
# repository state clean.
gates: build vet test rejection

clean:
	rm -rf bin/

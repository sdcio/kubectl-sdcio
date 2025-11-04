.PHONY: build
build:
	CGO_ENABLED=0 go build -o kubectl-sdcio ./cmd/kubectl-sdcio.go 

.PHONY: go-tests
go-tests:
	go test ./...

.PHONY: test
test: go-tests



MOCKDIR = ./mocks
.PHONY: mocks-gen
mocks-gen: mocks-rm ## Generate mocks for all the defined interfaces.
	mkdir -p $(MOCKDIR)
	go install go.uber.org/mock/mockgen@latest
# 	mockgen -package=mocknetconf -source=pkg/datastore/target/netconf/driver.go -destination=$(MOCKDIR)/mocknetconf/driver.go

.PHONY: mocks-rm
mocks-rm: ## remove generated mocks
	rm -rf $(MOCKDIR)/*

.PHONY: unit-tests
unit-tests: mocks-gen
	rm -rf /tmp/sdcio/dataserver-tests/coverage
	mkdir -p /tmp/sdcio/dataserver-tests/coverage
	CGO_ENABLED=0 go test -cover -race ./... -v -covermode atomic -args -test.gocoverdir="/tmp/sdcio/dataserver-tests/coverage"

.PHONY: build
build:
	CGO_ENABLED=0 go build -o kubectl-sdc ./cmd/kubectl-sdc.go 

.PHONY: go-tests
go-tests:
	go test ./...

.PHONY: test
test: go-tests



MOCKDIR = ./mocks
.PHONY: mocks-gen
mocks-gen: mocks-rm ## Generate mocks for all the defined interfaces.
	mkdir -p $(MOCKDIR)
	mkdir -p $(MOCKDIR)/blame $(MOCKDIR)/apply $(MOCKDIR)/deviations
	go install go.uber.org/mock/mockgen@latest
	mockgen -package=mockblame -source=pkg/commands/blame/blame.go -destination=$(MOCKDIR)/blame/blame.go
	mockgen -package=mockapply -source=pkg/commands/apply/apply.go -destination=$(MOCKDIR)/apply/apply.go
	mockgen -package=mockdeviations -source=pkg/commands/deviations/deviations.go -destination=$(MOCKDIR)/deviations/deviations.go

.PHONY: mocks-rm
mocks-rm: ## remove generated mocks
	rm -rf $(MOCKDIR)/*

.PHONY: unit-tests
unit-tests: mocks-gen
	rm -rf /tmp/sdcio/dataserver-tests/coverage
	mkdir -p /tmp/sdcio/dataserver-tests/coverage
	CGO_ENABLED=0 go test -cover -race ./... -v -covermode atomic -args -test.gocoverdir="/tmp/sdcio/dataserver-tests/coverage"

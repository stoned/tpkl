GO?=	go
GOLANGCI_LINT?= golangci-lint
GORELEASER?= goreleaser

PKL?= pkl

COVER_OUTPUT= cover.out

PKL_GENERATED_CODE_DIR= modules/tpkl

.PHONY: build
build tpkl: generate
	env CGO_ENABLED=0 $(GO) build

.PHONY: generate  # XXX
generate:
	$(GO) generate ./...

.PHONY: gen-pkl-tests
gen-pkl-tests: tpkl
	$(PKL) test --external-module-reader='tpkl=./tpkl readers' modules/testdata/pkl-test/test*.pkl --overwrite || [ "$$?" -eq 10 ] && $(MAKE) generate

.PHONY: test stest cover cover/html
GO_TEST= $(GO) test ./... -coverprofile $(COVER_OUTPUT)
test: generate
	$(GO_TEST) -race -p=1 $(TEST_FLAGS)

stest: generate
	$(GO_TEST) $(TEST_FLAGS)

cover cover/html: test
	$(GO) tool cover -func=$(COVER_OUTPUT)
	$(if $(subst .,,$(@D)),$(GO) tool cover -html=$(COVER_OUTPUT))

.PHONY: lint fmt fmt/diff
lint:
	$(GOLANGCI_LINT) run
fmt fmt/diff:
	$(GOLANGCI_LINT) fmt $(if $(subst .,,$(@D)),-d,)

.PHONY: build/dev
build/dev:
	$(GORELEASER) build --snapshot --clean

.PHONY: release/dev
release/dev:
	$(GORELEASER) release --snapshot

.PHONY: clean clobber
clean:
	$(GO) clean || rm -f ./tpkl
	-rm -f $(COVER_OUTPUT)
	-rm -rf $(PKL_GENERATED_CODE_DIR)
	-find . -name "*.txtar" -exec rm {} \;
clobber: clean
	rm -rf dist tpkl

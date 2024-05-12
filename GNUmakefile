default: testacc

install:
	go build

# Run acceptance tests
.PHONY: testacc
testacc: install
	TF_ACC_PROVIDER_NAMESPACE="rrey" TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m


.PHONY: gendoc
gendoc:
	GOOS=darwin GOARCH=amd64 go generate ./...

.PHONY: fixture
fixture:
	./spire-1.9.3/bin/spire-server entry create \
		-parentID spiffe://example.org/some/path \
		-spiffeID spiffe://example.org/some/datasource-test \
	    -selector "unix:uid:501"

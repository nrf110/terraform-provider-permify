default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_LOG=DEBUG TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

.PHONY: docs
docs:
	tfplugindocs generate --provider-name permify
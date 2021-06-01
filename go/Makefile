include scripts/common.mk

test.unit:
	@$(INFO) "testing"
	@go test ./... -v || $(FAIL)
	@$(OK) "done"

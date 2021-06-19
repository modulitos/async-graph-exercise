include scripts/common.mk

test.unit:
	@$(INFO) "testing"
	@go test ./... -v || $(FAIL)
	@$(OK) "done"
test.unit.one_node:
	@$(INFO) "testing"
	@go test ./... -run TestCalculateReward/one_node_100 -v || $(FAIL)
	@$(OK) "done"
build:
	@$(INFO) "building"
	@go build cmd/app.go
	@$(OK) "executable created at 'app'"

run:
	@$(INFO) "running"
	@go run cmd/app.go
	@$(OK) "done"
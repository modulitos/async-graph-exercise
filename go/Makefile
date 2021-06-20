include scripts/common.mk

test.unit:
	@$(INFO) "testing"
	@go test ./... -v || $(FAIL)
	@$(OK) "done"
test.unit.single:
	@$(INFO) "testing"
	@go test ./... -run TestCalculateReward/caching_many_calls -v || $(FAIL)
	@$(OK) "done"
build:
	@$(INFO) "building"
	@go build cmd/app.go
	@$(OK) "executable created at 'app'"

run:
	@$(INFO) "running"
	@go run cmd/app.go
	@$(OK) "done"
.PHONY: build
build:
	powershell -Command "Start-Process 'go' -ArgumentList 'build', 'C:/Users/gibel/GolandProjects/S3_project/auth/cmd/auth/main.go'"
	powershell -Command "Start-Process 'go' -ArgumentList 'build', 'C:/Users/gibel/GolandProjects/S3_project/S3/cmd/S3/main.go'"
	powershell -Command "Start-Process 'go' -ArgumentList 'build', 'C:/Users/gibel/GolandProjects/S3_project/APIGateway/cmd/gateway/main.go'"

.PHONY: test
test: test-go test-api test-frontend

.PHONY: test-go
test-go:
	set CGO_ENABLED=1
	go test -v -timeout 30s ./...

.PHONY: test-api
test-api:
	@echo "Running API tests from api_tests.go"
	@echo "Please ensure all services are running with 'make run' before testing"
	@echo "You can also use api_tests.http to run the tests in an IDE with HTTP client support"
	go test tests/api_test.go tests/model.go

.PHONY: test-frontend
test-frontend:
	@echo "Running Frontend tests from frontend_tests.go"
	@echo "Please ensure all services are running with 'make run' before testing"
	@echo "You can also use frontend_tests.http to run the tests in an IDE with HTTP client support"
	go test tests/frontend_test.go tests/model.go

.PHONY: run
run:
	powershell -Command "Start-Process 'go' -ArgumentList 'run', 'C:/Users/gibel/GolandProjects/S3_project/auth/cmd/auth/main.go'"
	powershell -Command "Start-Process 'go' -ArgumentList 'run', 'C:/Users/gibel/GolandProjects/S3_project/S3/cmd/S3/main.go'"
	powershell -Command "Start-Process 'go' -ArgumentList 'run', 'C:/Users/gibel/GolandProjects/S3_project/APIGateway/cmd/gateway/main.go'"

.DEFAULT_GOAL := build

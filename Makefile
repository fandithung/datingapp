APP_NAME = dating-app
MAIN_PATH = cmd/api/main.go
BUILD_DIR = build
POSTGRES_URL = postgresql://dating_app:dating_app_password@localhost:5432/dating_app_db?sslmode=disable

GOBASE = $(shell pwd)
GOBIN = $(GOBASE)/$(BUILD_DIR)

COLOR_RESET = \033[0m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m

.PHONY: all build clean run test migrate-* install-tools mock seed

all: clean build

build:
	@echo "$(COLOR_GREEN)Building $(APP_NAME)...$(COLOR_RESET)"
	@go build -o $(GOBIN)/$(APP_NAME) $(MAIN_PATH)

clean:
	@echo "$(COLOR_GREEN)Cleaning...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@go clean

run:
	@echo "$(COLOR_GREEN)Running $(APP_NAME)...$(COLOR_RESET)"
	@go run $(MAIN_PATH)

run-dev:
	@echo "$(COLOR_GREEN)Running $(APP_NAME) in development mode...$(COLOR_RESET)"
	@air

docker-up:
	@echo "$(COLOR_GREEN)Starting docker containers...$(COLOR_RESET)"
	@docker-compose up -d

docker-down:
	@echo "$(COLOR_GREEN)Stopping docker containers...$(COLOR_RESET)"
	@docker-compose down

migrate-up:
	@echo "$(COLOR_GREEN)Running migrations up...$(COLOR_RESET)"
	@migrate -path migrations -database "$(POSTGRES_URL)" up

migrate-down:
	@echo "$(COLOR_GREEN)Running migrations down...$(COLOR_RESET)"
	@migrate -path migrations -database "$(POSTGRES_URL)" down


migrate-version:
	@echo "$(COLOR_GREEN)Current migration version:$(COLOR_RESET)"
	@migrate -path migrations -database "$(POSTGRES_URL)" version

seed:
	@read -p "Enter number of users to seed (default: 1000): " count; \
	if [ -z "$$count" ]; then \
		count=1000; \
	fi; \
	echo "$(COLOR_GREEN)Seeding database with $$count users...$(COLOR_RESET)"; \
	go run cmd/seeder/main.go -count=$$count

install-tools:
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

mock:
	@echo "$(COLOR_GREEN)Generating mocks...$(COLOR_RESET)"
	@mockgen -source=internal/datingapp.go -destination=internal/service/mock/mock_service.go
	@mockgen -source=internal/repository/repository.go -destination=internal/repository/mock/mock_repository.go

lint:
	@echo "$(COLOR_GREEN)Running linter...$(COLOR_RESET)"
	@golangci-lint run

test: mock
	@echo "$(COLOR_GREEN)Running tests...$(COLOR_RESET)"
	@go test -v ./...
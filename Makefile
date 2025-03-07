include .env.dev
MIGRATION_PATH = ./migrations

APP_NAME = mygoapi
BUILD_DIR = bin
GO_FILES = ./cmd/api

# Detect OS
UNAME_S := $(shell uname -s)

# Set default GOOS and GOARCH based on OS
ifeq ($(UNAME_S), Darwin) # macOS
    TARGET_OS := darwin
    TARGET_ARCH := amd64
else ifeq ($(UNAME_S), Linux) # Linux
    TARGET_OS := linux
    TARGET_ARCH := amd64
else # Assume Windows (Git Bash, Cygwin, or WSL)
    TARGET_OS := windows
    TARGET_ARCH := amd64
endif

# Default Build (for the current OS)
build:
	@echo "Building for $(TARGET_OS)/$(TARGET_ARCH)..."
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -ldflags "-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(GO_FILES)

clean:
	rm -rf $(BUILD_DIR)/*

run-dev:
	GO_ENV=development air -c .air.dev.toml

run-prod:
	GO_ENV=production $(BUILD_DIR)/$(APP_NAME)

db-migration:
	@migrate create -seq -ext sql -dir $(MIGRATION_PATH) $(filter-out $@,$(MAKECMDGOALS))

db-migrate-up:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_URL) up

db-migrate-down:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_URL) down $(filter-out $@,$(MAKECMDGOALS))

db-migrate-force:
	@migrate -path=$(MIGRATION_PATH) -database=$(DB_URL) force $(filter-out $@,$(MAKECMDGOALS))
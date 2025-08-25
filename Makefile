.PHONY: lint api-generate ycf-zip ycf-create-function-or-ignore ycf-create-function-version ycf-timer ycf-clear ycf-deploy

ifneq (,$(wildcard .env))
include .env
export
endif

LOCAL_BIN := ./bin

# Lint
lint:
	golangci-lint run ./...

# Generate API
api-generate:
	go mod tidy
	go mod download
	go generate ./...

# Yandex Cloud Function
ycf-zip:
	zip -r '$(APP_NAME).zip' internal pkg handler.go go.mod go.sum -x "*/*_test.go" -x ".DS_Store"

ycf-create-function-or-ignore:
	@if yc serverless function get --name '$(APP_NAME)' > /dev/null 2>&1; then \
		echo "Function '$(APP_NAME)' already exists, skipping creation."; \
	else \
		yc serverless function create --name "$(APP_NAME)"; \
	fi

ycf-create-function-version:
	yc serverless function version create \
	--function-name '$(APP_NAME)' \
	--service-account-id '$(YCF_SA_ID)' \
	--runtime golang123 \
	--entrypoint handler.Handler \
	--execution-timeout $(YCF_TIMEOUT) \
	--memory 128m \
	--environment APP_MODE=prod \
	--environment YDB_BASE_URL=$(YDB_BASE_URL) \
    --environment YDB_PATH=$(YDB_PATH) \
    --environment YDB_CREDS_FILE=$(YDB_CREDS_FILE) \
    --environment YDB_DATABASE_NAME=$(YDB_DATABASE_NAME) \
    --environment YDB_TABLES_MAP=$(YDB_TABLES_MAP) \
    --environment YDB_BATCH_SIZE=$(YDB_BATCH_SIZE) \
    --environment YDB_TIMEOUT=$(YDB_TIMEOUT) \
    --environment TG_TOKEN=$(TG_TOKEN) \
    --environment TG_CHAT_ID=$(TG_CHAT_ID) \
    --environment ESB_BASE_URL=$(ESB_BASE_URL) \
    --environment ESB_API_KEY=$(ESB_API_KEY) \
    --environment ESB_TIMEOUT=$(ESB_TIMEOUT) \
    --environment ESB_LIMIT_PAGE_SIZE=$(ESB_LIMIT_PAGE_SIZE) \
    --environment APP_NAME=$(APP_NAME) \
    --environment APP_VERSION=$(APP_VERSION) \
	--source-path "./$(APP_NAME).zip"

ycf-timer:
	@if yc serverless trigger get --name "run-$(APP_NAME)" > /dev/null 2>&1; then \
  		echo "Trigger 'run-$(APP_NAME)' already exists, skipping creation."; \
	else \
		yc serverless trigger create timer \
		--cron-expression $(YCF_CRON) \
		--invoke-function-name "$(APP_NAME)" \
		--invoke-function-service-account-id "$(YCF_SA_ID)" \
		--name "run-$(APP_NAME)"; \
	fi

ycf-clear:
	rm "./$(APP_NAME).zip"

ycf-deploy: api-generate ycf-zip ycf-create-function-or-ignore ycf-create-function-version ycf-timer ycf-clear
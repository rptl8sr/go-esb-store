# Go ESB Store (Store State Observer)
A small Go service that fetches retail store data from an external ESB API and upserts it into YDB.
## Features
- ESB integration for:
    - retrieving total count of stores
    - fetching paginated store data (filterable)
- Persistence to YDB with batched upsert
- Dev mode: creates tables if they do not exist
- Prod mode: uses instance metadata credentials from the attached service account
- Deployable as a Yandex Cloud Function with a CRON timer trigger

## Requirements
- Go 1.23 (Go SDK 1.23.x)
- YDB Go SDK v3
- Yandex Cloud credentials helpers (metadata/dev key-file)
- OpenAPI spec (api.yaml) for client generation
- Yandex Cloud CLI (yc) for deployment
- make, zip

## Configuration
Provide configuration via environment variables (or your config file), see the dumb.env for examples

## Make targets
- `make lint` — run golangci-lint (if installed)
- `make api-generate` — `go generate ./...` plus `go mod tidy`
- `make ycf-zip` — build a ZIP bundle for Cloud Function
- `make ycf-create-function-or-ignore` — create function if it does not exist
- `make ycf-create-function-version` — create a new function version (uses env vars)
- `make ycf-timer` — create timer trigger if not exists
- `make ycf-clear` — remove build ZIP
- `make ycf-deploy` — full pipeline: lint → generate → zip → create function (if needed) → create version → ensure timer → cleanup

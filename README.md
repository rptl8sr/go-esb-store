# Go ESB Store (Store State Observer)
A small Go service that fetches retail store data from an external ESB API and upserts it into YDB.
## Features
- Pulls stores with paging from the upstream API (see api.yaml).
- Batch UPSERT into YDB.
- Optional schema bootstrap in dev mode (creates table and index if missing).

## Requirements
- Go 1.23+
- YDB database (Serverless or dedicated)
- Credentials for YDB access

## Configuration
Provide configuration via environment variables (or your config file), for example:
- MODE: dev or prod
- YDB_ENDPOINT: YDB endpoint URL (e.g., grpc://<ydb-host>:2135)
- YDB_DATABASE: YDB database path (e.g., /../../path  )
- YDB_PATH: optional additional path part if needed by your setup
- YDB_CREDS_FILE: path to service account key file (required in dev)
- YDB_TIMEOUT_SEC: connection timeout in seconds (e.g., 10)
- YDB_BATCH_SIZE: batch size for UPSERT (default 500)
- STORES_TABLE: table name override (default stores)

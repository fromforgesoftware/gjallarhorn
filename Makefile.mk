GJALLARHORN_DIR ?= .

.PHONY: gjallarhorn-proto gjallarhorn-build gjallarhorn-test gjallarhorn-vet gjallarhorn-lint gjallarhorn-migrate gjallarhorn-run

gjallarhorn-proto:          ## Regenerate Go code from the Gjallarhorn protos
	cd $(GJALLARHORN_DIR) && buf generate

gjallarhorn-lint:           ## Lint the Gjallarhorn protos
	cd $(GJALLARHORN_DIR) && buf lint

gjallarhorn-build:          ## Build all Gjallarhorn packages
	cd $(GJALLARHORN_DIR) && go build ./...

gjallarhorn-vet:            ## Vet the Gjallarhorn module
	cd $(GJALLARHORN_DIR) && go vet ./...

gjallarhorn-test:           ## Run Gjallarhorn unit tests
	cd $(GJALLARHORN_DIR) && go test ./...

GJALLARHORN_LOCAL_ENV := SVC_NAME=gjallarhorn REST_ADDRESS=:8080 HTTP_ADDRESS=:8080 GRPC_ADDRESS=:9090

gjallarhorn-migrate:        ## Apply Gjallarhorn migrations (reads DB_* env)
	cd $(GJALLARHORN_DIR) && go run ./cmd/migrator

gjallarhorn-run:            ## Run the Gjallarhorn server locally
	cd $(GJALLARHORN_DIR) && $(GJALLARHORN_LOCAL_ENV) go run ./cmd/server

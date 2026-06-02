HERALD_DIR ?= services/herald

.PHONY: herald-proto herald-build herald-test herald-vet herald-lint herald-migrate herald-run

herald-proto:          ## Regenerate Go code from the Herald protos
	cd $(HERALD_DIR) && buf generate

herald-lint:           ## Lint the Herald protos
	cd $(HERALD_DIR) && buf lint

herald-build:          ## Build all Herald packages
	cd $(HERALD_DIR) && go build ./...

herald-vet:            ## Vet the Herald module
	cd $(HERALD_DIR) && go vet ./...

herald-test:           ## Run Herald unit tests
	cd $(HERALD_DIR) && go test ./...

HERALD_LOCAL_ENV := SVC_NAME=herald REST_ADDRESS=:8080 HTTP_ADDRESS=:8080 GRPC_ADDRESS=:9090

herald-migrate:        ## Apply Herald migrations (reads DB_* env)
	cd $(HERALD_DIR) && go run ./cmd/migrator

herald-run:            ## Run the Herald server locally
	cd $(HERALD_DIR) && $(HERALD_LOCAL_ENV) go run ./cmd/server

install-deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

validate-openapi:
	redocly lint openapi/openapi.yaml

bundle-openapi: validate-openapi
	redocly bundle openapi/openapi.yaml -o openapi/bundled.yaml

gen-openapi-server: bundle-openapi
	go generate pkg/api/generate.go


gen: gen-openapi-server

test-data: up
	go run scripts/generate_test_data.go

run:
	go run ./cmd/service/main.go --envFile=configs/.env.service.local

up-debug:
	docker compose -f docker-compose.dev.yaml up -d

up:
	docker compose up -d

clean:
	docker compose down --volumes
	docker compose -f docker-compose.dev.yaml down --volumes

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/bufbuild/buf/cmd/buf@latest

gen:
	buf generate

run:
	go run ./cmd/service/main.go --envFile=configs/.env.service.local

up-debug:
	docker compose -f docker-compose.dev.yaml up -d

up:
	docker compose up -d

clean:
	docker compose down --volumes
	docker compose -f docker-compose.dev.yaml down --volumes
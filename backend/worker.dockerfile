FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o worker cmd/worker/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates
RUN addgroup -g 65532 nonroot &&\
    adduser -S -u 65532 nonroot -G nonroot

WORKDIR /app

COPY --from=builder /workspace/worker .

USER nonroot

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ps aux | grep '[w]orker' || exit 1

CMD ["./worker"]
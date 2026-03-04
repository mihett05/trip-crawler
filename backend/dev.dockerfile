FROM golang:1.25-trixie

RUN apt-get update \
 && apt-get install -y --no-install-recommends  \
    ca-certificates \
    git \
    curl \
    wget \
    vim \
    htop \
    net-tools \
    dnsutils \
    iputils-ping

RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install github.com/air-verse/air@latest

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080 40000

CMD ["air", "-c", ".air.toml"]

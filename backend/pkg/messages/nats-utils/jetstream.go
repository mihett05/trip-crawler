package natsutils

import "github.com/nats-io/nats.go/jetstream"

//go:generate bash -c "mockgen -source=$GOFILE -destination=$(basename ${GOFILE} .go)_mock.go -package=$GOPACKAGE"
type JetStream interface {
	jetstream.JetStream
}

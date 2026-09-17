package natsx

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type Bus struct {
	Conn   *nats.Conn
	JS     jetstream.JetStream
	Stream string
}

func Connect(url, stream string, subjects []string, log *zap.Logger) (*Bus, error) {
	conn, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second),
		nats.Name("golangFoodService"),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("jetstream: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     stream,
		Subjects: subjects,
		Storage:  jetstream.FileStorage,
	})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure stream %s: %w", stream, err)
	}

	log.Info("nats connected", zap.String("stream", stream))
	return &Bus{Conn: conn, JS: js, Stream: stream}, nil
}

func (b *Bus) Publish(ctx context.Context, subject string, payload []byte) error {
	_, err := b.JS.Publish(ctx, subject, payload)
	return err
}

func (b *Bus) Close() {
	if b != nil && b.Conn != nil {
		b.Conn.Close()
	}
}

func (b *Bus) Ping() error {
	if b == nil || b.Conn == nil || !b.Conn.IsConnected() {
		return fmt.Errorf("nats disconnected")
	}
	return nil
}

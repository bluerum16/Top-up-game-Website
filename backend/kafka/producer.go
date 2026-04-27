package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/irham/topup-backend/config"
	kafkago "github.com/segmentio/kafka-go"
)

const (
	TopicOrders    = "topup.orders"
	TopicPayments  = "topup.payments"
	TopicUsers     = "topup.users"
	TopicPageviews = "topup.pageviews"
	TopicSearch    = "topup.search"
	TopicContacts  = "topup.contacts"
)

type Producer struct {
	brokers []string
	mu      sync.RWMutex
	writers map[string]*kafkago.Writer
}

func NewProducer(cfg config.KafkaConfig) *Producer {
	return &Producer{
		brokers: cfg.Brokers,
		writers: make(map[string]*kafkago.Writer),
	}
}

func (p *Producer) writer(topic string) *kafkago.Writer {
	p.mu.RLock()
	w, ok := p.writers[topic]
	p.mu.RUnlock()
	if ok {
		return w
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if w, ok := p.writers[topic]; ok {
		return w
	}
	w = &kafkago.Writer{
		Addr:         kafkago.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafkago.Hash{},
		BatchTimeout: 50 * time.Millisecond,
		RequiredAcks: kafkago.RequireOne,
		Async:        false,
	}
	p.writers[topic] = w
	return w
}

// Publish marshals payload as JSON and writes a single message.
// key is used for partitioning (e.g. order_id, user_id) — pass "" for round-robin.
func (p *Producer) Publish(ctx context.Context, topic, key string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("kafka marshal %s: %w", topic, err)
	}
	msg := kafkago.Message{Value: body}
	if key != "" {
		msg.Key = []byte(key)
	}
	if err := p.writer(topic).WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka publish %s: %w", topic, err)
	}
	return nil
}

func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	var firstErr error
	for _, w := range p.writers {
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	p.writers = nil
	return firstErr
}

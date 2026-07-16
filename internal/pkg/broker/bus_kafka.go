package broker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog/log"

	"github.com/AlexBond702/order-service/internal/pkg/broker/codec"
)

var _ Bus[any] = (*kafkaBus[any])(nil)

const (
	consumeBackoffInitial = 500 * time.Millisecond
	consumeBackoffMax     = 30 * time.Second
	consumeMaxRetries     = 5
)

type EventIdGetter interface {
	EventId() string
}

type kafkaBus[T any] struct {
	client        *KafkaClient
	codec         codec.Codec[T]
	topic         string
	consumerGroup string
	consumer      sarama.ConsumerGroup
}

func NewBus[T any](client *KafkaClient, codec codec.Codec[T], topic, consumerGroup string) (Bus[T], error) {
	if client == nil || topic == "" {
		return nil, fmt.Errorf("broker: empty client or topic")
	}
	group := Coalesce(consumerGroup, client.defaultConsumerGroup)
	if group == "" {
		return nil, fmt.Errorf("newbus: empty group")
	}
	return &kafkaBus[T]{
		client:        client,
		codec:         codec,
		topic:         topic,
		consumerGroup: group,
	}, nil
}

func MustKafkaBus[T any](client *KafkaClient, c codec.Codec[T], topic, consumerGroup string) Bus[T] {
	bus, err := NewBus(client, c, topic, consumerGroup)
	if err != nil {
		log.Fatal().Err(err).Str("topic", topic).Msg("error to create newbus") ///
	}
	return bus
}

func getEventId[T any](v *T) string {
	get, ok := any(v).(EventIdGetter)
	if ok {
		key := get.EventId()
		if key != "" {
			return key
		}
	}
	return uuid.Must(uuid.NewV4()).String()
}

func (b *kafkaBus[T]) Send(_ context.Context, msg *T, headers ...Header) error {
	data, err := b.codec.Encode(msg)
	if err != nil {
		return fmt.Errorf("broker: codec encode: %w", err)
	}
	key := getEventId(msg)
	headerSl := make([]sarama.RecordHeader, 0, len(headers))
	if len(headers) > 0 {
		for _, header := range headers {
			headerSl = append(headerSl, sarama.RecordHeader{
				Key:   []byte(header.Key),
				Value: []byte(header.Value),
			})
		}
	}
	message := &sarama.ProducerMessage{
		Topic:   b.topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(data),
		Headers: headerSl,
	}
	_, _, err = b.client.Producer().SendMessage(message)
	if err != nil {
		return fmt.Errorf("broker: send message: %w", err)
	}
	return nil
}

func (b *kafkaBus[T]) QueueName() string {
	return b.topic
}

func (b *kafkaBus[T]) Subscribe(ctx context.Context, wg *sync.WaitGroup, handler MessageHandler[T]) error {
	consumer, err := b.client.NewConsumerGroup(b.consumerGroup)
	if err != nil {
		return fmt.Errorf("broker: create consumer group %q: %w", b.consumerGroup, err)
	}
	b.consumer = consumer

	h := &consumerGroupHandler[T]{codec: b.codec, handler: handler, topic: b.topic}

	if wg != nil {
		wg.Add(1)
	}
	go func() {
		for err := range consumer.Errors() {
			log.Warn().Err(err).Str("topic", b.topic).Msg("consumer group error")
		}
	}()

	go func() {
		if wg != nil {
			defer wg.Done()
		}

		defer func() {
			if err := consumer.Close(); err != nil && !errors.Is(err, sarama.ErrClosedClient) {
				log.Error().Err(err).Str("topic", b.topic).Msg("failed to close consumer group")
			}
		}()

		backoff := consumeBackoffInitial
		fails := 0

		for {
			err := consumer.Consume(ctx, []string{b.topic}, h)
			if ctx.Err() != nil {
				return
			}
			if err == nil {
				backoff = consumeBackoffInitial
				fails = 0
				continue
			}
			fails++
			if fails >= consumeMaxRetries {
				log.Fatal().Err(err).Msg("broker is not available")
			}
			log.Warn().Err(err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
			backoff = min(backoff*2, consumeBackoffMax)
		}
	}()
	return nil
}

type consumerGroupHandler[T any] struct {
	codec   codec.Codec[T]
	handler MessageHandler[T]
	topic   string
}

func (h *consumerGroupHandler[T]) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler[T]) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler[T]) ConsumeClaim(
	session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim,
) error {
	ctx := session.Context()
	defer session.Commit()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			if err := h.process(ctx, session, msg); err != nil {
				return err
			}
		}
	}
}

func (h *consumerGroupHandler[T]) process(ctx context.Context,
	session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage,
) error {
	decoded, err := h.codec.Decode(msg.Value)
	if err != nil {
		log.Error().Err(err)
		session.MarkMessage(msg, "")
		return nil
	}
	headers := make([]Header, 0, len(msg.Headers))
	for _, h := range msg.Headers {
		headers = append(headers, Header{
			Value: string(h.Value),
			Key:   string(h.Key),
		})
	}
	err = h.handler(ctx, decoded, headers)
	if err != nil && IsNotCriticalError(err) {
		log.Warn().Err(err)
		session.MarkMessage(msg, "")
		return nil
	}
	if err != nil {
		log.Error().Err(err)
		return err
	}
	session.MarkMessage(msg, "")
	return nil
}

func (b *kafkaBus[T]) Close() error {
	if b.consumer != nil {
		err := b.consumer.Close()
		return err
	}
	return nil
}

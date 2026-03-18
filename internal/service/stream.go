package service

import (
	"SneakerFlash/internal/infra/redis"
	"context"
	"fmt"
)

type StreamService struct{}

func NewStreamService() *StreamService {
	return &StreamService{}
}

func (s *StreamService) SubscribeOrder(userID, orderID uint) (<-chan []byte, func(), error) {
	return subscribeTopic(orderStreamTopic(userID, orderID))
}

func (s *StreamService) SubscribeProduct(productID uint) (<-chan []byte, func(), error) {
	return subscribeTopic(productStreamTopic(productID))
}

func subscribeTopic(topic string) (<-chan []byte, func(), error) {
	if redis.RDB == nil {
		ch, unsubscribe := broker.Subscribe(topic)
		return ch, unsubscribe, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	pubsub := redis.RDB.Subscribe(ctx, topic)
	if _, err := pubsub.Receive(ctx); err != nil {
		cancel()
		_ = pubsub.Close()
		return nil, func() {}, fmt.Errorf("subscribe %s: %w", topic, err)
	}

	out := make(chan []byte, 8)
	go func() {
		defer close(out)
		defer pubsub.Close()
		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				select {
				case out <- []byte(msg.Payload):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, func() {
		cancel()
	}, nil
}

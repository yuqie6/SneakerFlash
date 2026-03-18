package service

import (
	"SneakerFlash/internal/infra/redis"
	"SneakerFlash/internal/model"
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"sync"
)

type StreamEvent struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type streamBroker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan []byte]struct{}
}

func newStreamBroker() *streamBroker {
	return &streamBroker{
		subscribers: make(map[string]map[chan []byte]struct{}),
	}
}

func (b *streamBroker) Subscribe(topic string) (<-chan []byte, func()) {
	ch := make(chan []byte, 8)
	b.mu.Lock()
	if _, ok := b.subscribers[topic]; !ok {
		b.subscribers[topic] = make(map[chan []byte]struct{})
	}
	b.subscribers[topic][ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		if subs, ok := b.subscribers[topic]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(b.subscribers, topic)
			}
		}
		b.mu.Unlock()
		close(ch)
	}
}

func (b *streamBroker) Publish(topic string, event StreamEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	b.publishPayload(topic, data)
}

func (b *streamBroker) publishPayload(topic string, data []byte) {
	b.mu.RLock()
	subs := b.subscribers[topic]
	for ch := range subs {
		select {
		case ch <- data:
		default:
		}
	}
	b.mu.RUnlock()
}

var broker = newStreamBroker()

func publishStreamEvent(topic string, event StreamEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	if redis.RDB != nil {
		if publishErr := redis.RDB.Publish(context.Background(), topic, data).Err(); publishErr == nil {
			return
		} else {
			slog.Warn("发布实时事件到 Redis 失败，回退到进程内广播", slog.String("topic", topic), slog.Any("err", publishErr))
		}
	}

	broker.publishPayload(topic, data)
}

func publishOrderEvent(userID, orderID uint, status model.OrderStatus, paymentStatus model.PaymentStatus) {
	publishStreamEvent(orderStreamTopic(userID, orderID), StreamEvent{
		Event: "order_update",
		Data: map[string]any{
			"order_id":       orderID,
			"status":         status,
			"payment_status": paymentStatus,
		},
	})
}

func publishProductStockEvent(productID uint, stock int) {
	publishStreamEvent(productStreamTopic(productID), StreamEvent{
		Event: "stock_update",
		Data: map[string]any{
			"product_id": productID,
			"stock":      stock,
		},
	})
}

func orderStreamTopic(userID, orderID uint) string {
	return "order:" + itoa(userID) + ":" + itoa(orderID)
}

func productStreamTopic(productID uint) string {
	return "product:" + itoa(productID)
}

func itoa(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

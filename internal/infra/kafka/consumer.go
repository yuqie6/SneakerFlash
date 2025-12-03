package kafka

import (
	"SneakerFlash/internal/config"
	"context"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// BatchMessageHandler 批量消息处理函数
type BatchMessageHandler func(msgs [][]byte) error

// StartBatchConsumer 启动批量消费模式的 Kafka Consumer
func StartBatchConsumer(cfg config.KafkaConfig, handler BatchMessageHandler) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	groupID := "sneaker-group"
	group, err := sarama.NewConsumerGroup(cfg.Brokers, groupID, config)
	if err != nil {
		log.Fatalf("[ERROR] 创建消费组失败: %v", err)
	}
	defer group.Close()

	// 设置默认值
	batchSize := cfg.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	flushInterval := cfg.FlushInterval
	if flushInterval <= 0 {
		flushInterval = 200
	}

	ctx := context.Background()
	consumer := &BatchConsumerHandler{
		callback:      handler,
		batchSize:     batchSize,
		flushInterval: time.Duration(flushInterval) * time.Millisecond,
		buffer:        make([]msgWithSession, 0, batchSize),
	}

	log.Printf("[INFO] Worker 正在监听 kafka topic: %s (batch_size=%d, flush_interval=%dms)",
		cfg.Topic, batchSize, flushInterval)

	for {
		topics := []string{cfg.Topic}
		err := group.Consume(ctx, topics, consumer)
		if err != nil {
			log.Printf("[ERROR] 消费异常: %v", err)
		}
	}
}

// msgWithSession 记录消息和对应的 session，用于批量确认
type msgWithSession struct {
	msg  *sarama.ConsumerMessage
	sess sarama.ConsumerGroupSession
}

// BatchConsumerHandler 批量消费处理器
type BatchConsumerHandler struct {
	callback      BatchMessageHandler
	batchSize     int
	flushInterval time.Duration
	buffer        []msgWithSession
	mu            sync.Mutex
}

func (h *BatchConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *BatchConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *BatchConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ticker := time.NewTicker(h.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				// channel 关闭，刷盘剩余消息
				h.flush()
				return nil
			}

			h.mu.Lock()
			h.buffer = append(h.buffer, msgWithSession{msg: msg, sess: sess})

			if len(h.buffer) >= h.batchSize {
				h.flushLocked()
			}
			h.mu.Unlock()

		case <-ticker.C:
			// 定时刷盘，避免消息延迟
			h.mu.Lock()
			if len(h.buffer) > 0 {
				h.flushLocked()
			}
			h.mu.Unlock()

		case <-sess.Context().Done():
			// session 结束
			h.flush()
			return nil
		}
	}
}

// flush 加锁后刷盘
func (h *BatchConsumerHandler) flush() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.flushLocked()
}

// flushLocked 刷盘（调用前需持有锁）
func (h *BatchConsumerHandler) flushLocked() {
	if len(h.buffer) == 0 {
		return
	}

	// 提取消息体
	msgBodies := make([][]byte, len(h.buffer))
	for i, item := range h.buffer {
		msgBodies[i] = item.msg.Value
	}

	// 批量处理
	startTime := time.Now()
	err := h.callback(msgBodies)
	elapsed := time.Since(startTime)

	if err != nil {
		log.Printf("[ERROR] 批量处理失败 (count=%d, elapsed=%v): %v", len(h.buffer), elapsed, err)
		// 批量失败时不确认 offset，等待重试
		// 注意：这里可以考虑部分成功的场景，但为简化先全量重试
		h.buffer = h.buffer[:0]
		return
	}

	// 批量确认 offset
	for _, item := range h.buffer {
		item.sess.MarkMessage(item.msg, "")
	}

	log.Printf("[INFO] 批量处理成功 (count=%d, elapsed=%v, tps=%.0f)",
		len(msgBodies), elapsed, float64(len(msgBodies))/elapsed.Seconds())

	// 清空 buffer
	h.buffer = h.buffer[:0]
}

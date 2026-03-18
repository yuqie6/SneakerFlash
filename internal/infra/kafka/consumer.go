package kafka

import (
	"SneakerFlash/internal/config"
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// BatchMessageHandler 批量消息处理函数，返回处理失败的消息索引列表
type BatchMessageHandler func(msgs [][]byte) (failedIndexes []int, err error)

// StartBatchConsumer 启动批量消费模式的 Kafka Consumer
func StartBatchConsumer(cfg config.KafkaConfig, handler BatchMessageHandler) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Return.Errors = true
	saramaCfg.Consumer.Offsets.Initial = resolveInitialOffset(cfg.InitialOffset)

	groupID := resolveConsumerGroup(cfg.ConsumerGroup)
	group, err := sarama.NewConsumerGroup(cfg.Brokers, groupID, saramaCfg)
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
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	ctx := context.Background()
	consumer := &BatchConsumerHandler{
		callback:      handler,
		batchSize:     batchSize,
		flushInterval: time.Duration(flushInterval) * time.Millisecond,
		buffer:        make([]msgWithSession, 0, batchSize),
		maxRetries:    maxRetries,
		dlqTopic:      cfg.DLQTopic,
		topic:         cfg.Topic,
		retryStore:    newRetryStore(),
		sendToDLQ:     SendToDLQ,
	}

	log.Printf("[INFO] Worker 正在监听 kafka topic: %s (group_id=%s, initial_offset=%s, batch_size=%d, flush_interval=%dms, max_retries=%d, dlq_topic=%s)",
		cfg.Topic, groupID, normalizeInitialOffset(cfg.InitialOffset), batchSize, flushInterval, maxRetries, cfg.DLQTopic)

	for {
		topics := []string{cfg.Topic}
		err := group.Consume(ctx, topics, consumer)
		if err != nil {
			log.Printf("[ERROR] 消费异常: %v", err)
		}
	}
}

func resolveConsumerGroup(group string) string {
	group = strings.TrimSpace(group)
	if group == "" {
		return "sneaker-group"
	}
	return group
}

func resolveInitialOffset(offset string) int64 {
	switch normalizeInitialOffset(offset) {
	case "newest":
		return sarama.OffsetNewest
	default:
		return sarama.OffsetOldest
	}
}

func normalizeInitialOffset(offset string) string {
	offset = strings.TrimSpace(strings.ToLower(offset))
	if offset == "newest" {
		return "newest"
	}
	return "oldest"
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
	maxRetries    int
	dlqTopic      string
	topic         string
	retryStore    retryStore
	sendToDLQ     func(string, DLQMessage) error
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

// msgKey 生成消息唯一键用于重试计数
func msgKey(msg *sarama.ConsumerMessage) string {
	return fmt.Sprintf("%s:%d:%d", msg.Topic, msg.Partition, msg.Offset)
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
	failedIndexes, err := h.callback(msgBodies)
	elapsed := time.Since(startTime)

	if err != nil {
		log.Printf("[ERROR] 批量处理失败 (count=%d, elapsed=%v): %v", len(h.buffer), elapsed, err)
		// 全部失败，处理重试逻辑
		h.handleBatchFailure(err)
		return
	}

	// 部分失败情况
	if len(failedIndexes) > 0 {
		log.Printf("[WARN] 批量处理部分失败 (total=%d, failed=%d, elapsed=%v)",
			len(h.buffer), len(failedIndexes), elapsed)
		h.handlePartialFailure(failedIndexes)
		return
	}

	// 全部成功，批量确认 offset
	for _, item := range h.buffer {
		h.ackMessage(item)
	}

	log.Printf("[INFO] 批量处理成功 (count=%d, elapsed=%v, tps=%.0f)",
		len(msgBodies), elapsed, float64(len(msgBodies))/elapsed.Seconds())

	// 清空 buffer
	h.buffer = h.buffer[:0]
}

// handleBatchFailure 处理批量全部失败的情况
func (h *BatchConsumerHandler) handleBatchFailure(lastErr error) {
	errStr := "unknown error"
	if lastErr != nil {
		errStr = lastErr.Error()
	}
	for _, item := range h.buffer {
		h.handleFailedMessage(item, errStr)
	}
	h.buffer = h.buffer[:0]
}

// handlePartialFailure 处理部分失败的情况
func (h *BatchConsumerHandler) handlePartialFailure(failedIndexes []int) {
	failedSet := make(map[int]bool)
	for _, idx := range failedIndexes {
		failedSet[idx] = true
	}

	type partitionKey struct {
		topic     string
		partition int32
	}
	type bufferedResult struct {
		item   msgWithSession
		failed bool
	}

	grouped := make(map[partitionKey][]bufferedResult)
	for i, item := range h.buffer {
		key := partitionKey{topic: item.msg.Topic, partition: item.msg.Partition}
		grouped[key] = append(grouped[key], bufferedResult{
			item:   item,
			failed: failedSet[i],
		})
	}

	for _, results := range grouped {
		sort.Slice(results, func(i, j int) bool {
			return results[i].item.msg.Offset < results[j].item.msg.Offset
		})

		blocked := false
		for _, result := range results {
			if result.failed {
				if !h.handleFailedMessage(result.item, "message processing failed") {
					blocked = true
				}
				continue
			}

			if blocked {
				log.Printf("[INFO] 跳过确认成功消息，等待前序失败消息完成重试: topic=%s, partition=%d, offset=%d, current_retry_key=%s, dlq_topic=%s",
					result.item.msg.Topic, result.item.msg.Partition, result.item.msg.Offset, msgKey(result.item.msg), h.dlqTopic)
				continue
			}

			h.ackMessage(result.item)
		}
	}
	h.buffer = h.buffer[:0]
}

// handleFailedMessage 处理单条失败消息
func (h *BatchConsumerHandler) handleFailedMessage(item msgWithSession, errStr string) bool {
	key := msgKey(item.msg)
	retryCount := h.getAndIncrRetryCount(item.sess.Context(), key)

	if errStr == "" {
		errStr = "unknown error"
	}

	if retryCount >= h.maxRetries {
		// 达到最大重试次数，投递到 DLQ
		log.Printf("[WARN] 消息达到最大重试次数，投递到 DLQ: retry_key=%s, retry_count=%d, max_retries=%d, topic=%s, partition=%d, offset=%d, dlq_topic=%s",
			key, retryCount, h.maxRetries, item.msg.Topic, item.msg.Partition, item.msg.Offset, h.dlqTopic)

		dlqMsg := NewDLQMessage(item.msg.Topic, item.msg.Value, retryCount, errStr)
		if err := h.sendDLQ(h.dlqTopic, dlqMsg); err != nil {
			log.Printf("[ERROR] 投递 DLQ 失败: retry_key=%s, retry_count=%d, max_retries=%d, topic=%s, partition=%d, offset=%d, dlq_topic=%s, err=%v",
				key, retryCount, h.maxRetries, item.msg.Topic, item.msg.Partition, item.msg.Offset, h.dlqTopic, err)
		}

		// 标记消息已处理（即使 DLQ 失败也要 ack，避免无限重试）
		h.ackMessage(item)
		return true
	} else {
		// 未达到最大重试次数，不 MarkMessage，让 Kafka 重新投递
		log.Printf("[INFO] 消息处理失败，等待重试: retry_key=%s, retry_count=%d, max_retries=%d, topic=%s, partition=%d, offset=%d, dlq_topic=%s",
			key, retryCount, h.maxRetries, item.msg.Topic, item.msg.Partition, item.msg.Offset, h.dlqTopic)
		// 不 MarkMessage，Kafka 会在 rebalance 或 session 超时后重新投递
		// 注意：这种方式可能导致重复消费，业务层需要保证幂等
		return false
	}
}

func (h *BatchConsumerHandler) ackMessage(item msgWithSession) {
	item.sess.MarkMessage(item.msg, "")
	h.removeRetryCount(item.sess.Context(), item.msg)
}

func (h *BatchConsumerHandler) sendDLQ(dlqTopic string, msg DLQMessage) error {
	if h.sendToDLQ != nil {
		return h.sendToDLQ(dlqTopic, msg)
	}
	return SendToDLQ(dlqTopic, msg)
}

// getAndIncrRetryCount 获取并增加重试次数
func (h *BatchConsumerHandler) getAndIncrRetryCount(ctx context.Context, key string) int {
	if h.retryStore == nil {
		return 1
	}
	count, err := h.retryStore.Incr(ctx, key)
	if err != nil {
		if count > 0 {
			log.Printf("[WARN] 记录消费重试次数时设置过期失败，继续使用当前计数: key=%s, count=%d, err=%v", key, count, err)
			return count
		}
		log.Printf("[WARN] 记录消费重试次数失败，降级为单次重试: key=%s, err=%v", key, err)
		return 1
	}
	return count
}

// removeRetryCount 移除重试计数
func (h *BatchConsumerHandler) removeRetryCount(ctx context.Context, msg *sarama.ConsumerMessage) {
	if h.retryStore == nil {
		return
	}
	key := msgKey(msg)
	if err := h.retryStore.Delete(ctx, key); err != nil {
		log.Printf("[WARN] 清理消费重试次数失败: key=%s, err=%v", key, err)
	}
}

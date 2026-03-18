//go:build integration

package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	redisinfra "SneakerFlash/internal/infra/redis"
	goredis "github.com/redis/go-redis/v9"

	"github.com/IBM/sarama"
)

func TestBatchConsumerHandler_ConsumeToDLQWithRealKafka(t *testing.T) {
	brokers := kafkaIntegrationBrokers()
	redisClient := newIntegrationRedisClient()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("skip kafka dlq integration test: redis unavailable at %s: %v", redisClient.Options().Addr, err)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll

	client, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		t.Skipf("skip kafka dlq integration test: kafka unavailable at %v: %v", brokers, err)
	}
	defer func() {
		_ = client.Close()
	}()

	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		t.Fatalf("create cluster admin: %v", err)
	}
	defer func() {
		_ = admin.Close()
	}()

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		t.Fatalf("create producer: %v", err)
	}
	defer func() {
		_ = producer.Close()
	}()

	originalRedis := redisinfra.RDB
	originalProducer := Producer
	originalClient := Client
	redisinfra.RDB = redisClient
	Producer = producer
	Client = client
	defer func() {
		redisinfra.RDB = originalRedis
		Producer = originalProducer
		Client = originalClient
	}()

	suffix := time.Now().UnixNano()
	topic := fmt.Sprintf("it-seckill-orders-%d", suffix)
	dlqTopic := fmt.Sprintf("it-seckill-orders-dlq-%d", suffix)
	groupID := fmt.Sprintf("it-seckill-group-%d", suffix)

	createKafkaTopic(t, admin, topic, 1)
	createKafkaTopic(t, admin, dlqTopic, 1)
	defer func() {
		_ = admin.DeleteTopic(topic)
		_ = admin.DeleteTopic(dlqTopic)
	}()

	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(`{"order_num":"ORD-DLQ-IT-1"}`),
	})
	if err != nil {
		t.Fatalf("send source message: %v", err)
	}

	retryKey := fmt.Sprintf("%s:%d:%d", topic, partition, offset)
	retryCacheKey := retryCounterKey(retryKey)

	var callbackCalls int32
	for attempt := 1; attempt <= 3; attempt++ {
		processed := make(chan struct{}, 1)
		handler := &BatchConsumerHandler{
			callback: func(msgs [][]byte) ([]int, error) {
				atomic.AddInt32(&callbackCalls, 1)
				select {
				case processed <- struct{}{}:
				default:
				}
				failed := make([]int, len(msgs))
				for i := range msgs {
					failed[i] = i
				}
				return failed, nil
			},
			batchSize:     1,
			flushInterval: 10 * time.Millisecond,
			buffer:        make([]msgWithSession, 0, 1),
			maxRetries:    3,
			dlqTopic:      dlqTopic,
			topic:         topic,
			retryStore:    newRetryStore(),
			sendToDLQ:     SendToDLQ,
		}

		consumeOnce(t, brokers, groupID, topic, cfg, handler, processed)

		if attempt < 3 {
			waitForCondition(t, 5*time.Second, func() bool {
				val, err := redisClient.Get(context.Background(), retryCacheKey).Int64()
				return err == nil && val == int64(attempt)
			}, fmt.Sprintf("retry count reaches %d", attempt))
			assertNoDLQMessage(t, brokers, dlqTopic)
			continue
		}

		waitForCondition(t, 5*time.Second, func() bool {
			exists, err := redisClient.Exists(context.Background(), retryCacheKey).Result()
			return err == nil && exists == 0
		}, "retry counter cleared after dlq")
	}

	if got := atomic.LoadInt32(&callbackCalls); got != 3 {
		t.Fatalf("callback calls = %d, want 3", got)
	}

	dlqMsg := consumeOneKafkaMessage(t, client, dlqTopic)
	var payload DLQMessage
	if err := json.Unmarshal(dlqMsg.Value, &payload); err != nil {
		t.Fatalf("decode dlq message: %v", err)
	}
	if payload.OriginalTopic != topic {
		t.Fatalf("dlq original topic = %s, want %s", payload.OriginalTopic, topic)
	}
	if payload.RetryCount != 3 {
		t.Fatalf("dlq retry count = %d, want 3", payload.RetryCount)
	}
	if string(payload.OriginalValue) != `{"order_num":"ORD-DLQ-IT-1"}` {
		t.Fatalf("dlq original value = %s", string(payload.OriginalValue))
	}
}

func kafkaIntegrationBrokers() []string {
	raw := strings.TrimSpace(os.Getenv("SNEAKERFLASH_KAFKA_IT_BROKERS"))
	if raw == "" {
		raw = "127.0.0.1:19092"
	}
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			brokers = append(brokers, part)
		}
	}
	return brokers
}

func newIntegrationRedisClient() *goredis.Client {
	addr := strings.TrimSpace(os.Getenv("SNEAKERFLASH_REDIS_IT_ADDR"))
	if addr == "" {
		addr = "127.0.0.1:16379"
	}
	password := os.Getenv("SNEAKERFLASH_REDIS_IT_PASSWORD")
	if password == "" {
		password = "123456"
	}

	return goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
}

func createKafkaTopic(t *testing.T, admin sarama.ClusterAdmin, topic string, partitions int32) {
	t.Helper()

	err := admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: 1,
	}, false)
	if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
		t.Fatalf("create topic %s: %v", topic, err)
	}
}

func consumeOnce(t *testing.T, brokers []string, groupID string, topic string, cfg *sarama.Config, handler *BatchConsumerHandler, processed <-chan struct{}) {
	t.Helper()

	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		t.Fatalf("create consumer group: %v", err)
	}
	defer func() {
		_ = group.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- group.Consume(ctx, []string{topic}, handler)
	}()

	select {
	case <-processed:
	case <-time.After(5 * time.Second):
		t.Fatalf("consume topic %s timeout", topic)
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("consume topic %s: %v", topic, err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("wait consume loop exit timeout")
	}
}

func assertNoDLQMessage(t *testing.T, brokers []string, topic string) {
	t.Helper()

	cfg := sarama.NewConfig()
	client, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		t.Fatalf("create client for dlq check: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		t.Fatalf("create consumer for dlq check: %v", err)
	}
	defer func() {
		_ = consumer.Close()
	}()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		t.Fatalf("consume dlq partition: %v", err)
	}
	defer func() {
		_ = partitionConsumer.Close()
	}()

	select {
	case msg := <-partitionConsumer.Messages():
		t.Fatalf("unexpected dlq message before max retries: offset=%d value=%s", msg.Offset, string(msg.Value))
	case <-time.After(150 * time.Millisecond):
	}
}

func consumeOneKafkaMessage(t *testing.T, client sarama.Client, topic string) *sarama.ConsumerMessage {
	t.Helper()

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		t.Fatalf("create consumer: %v", err)
	}
	defer func() {
		_ = consumer.Close()
	}()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		t.Fatalf("consume topic %s: %v", topic, err)
	}
	defer func() {
		_ = partitionConsumer.Close()
	}()

	select {
	case msg := <-partitionConsumer.Messages():
		return msg
	case <-time.After(5 * time.Second):
		t.Fatalf("wait dlq message timeout")
		return nil
	}
}

func waitForCondition(t *testing.T, timeout time.Duration, fn func() bool, desc string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("wait condition timeout: %s", desc)
}

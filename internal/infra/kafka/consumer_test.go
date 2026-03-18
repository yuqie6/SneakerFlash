package kafka

import (
	"context"
	"reflect"
	"testing"

	"github.com/IBM/sarama"
)

type markedMessage struct {
	topic     string
	partition int32
	offset    int64
}

type mockConsumerGroupSession struct {
	ctx    context.Context
	marked []markedMessage
}

func (m *mockConsumerGroupSession) Claims() map[string][]int32 {
	return nil
}

func (m *mockConsumerGroupSession) MemberID() string {
	return ""
}

func (m *mockConsumerGroupSession) GenerationID() int32 {
	return 0
}

func (m *mockConsumerGroupSession) MarkOffset(topic string, partition int32, offset int64, metadata string) {
	m.marked = append(m.marked, markedMessage{topic: topic, partition: partition, offset: offset - 1})
}

func (m *mockConsumerGroupSession) Commit() {}

func (m *mockConsumerGroupSession) ResetOffset(string, int32, int64, string) {}

func (m *mockConsumerGroupSession) MarkMessage(msg *sarama.ConsumerMessage, _ string) {
	m.marked = append(m.marked, markedMessage{
		topic:     msg.Topic,
		partition: msg.Partition,
		offset:    msg.Offset,
	})
}

func (m *mockConsumerGroupSession) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

func TestBatchConsumerHandler_HandlePartialFailure_DoesNotAckPastFailedOffsetSamePartition(t *testing.T) {
	session := &mockConsumerGroupSession{}
	store := newMemoryRetryStore()
	handler := &BatchConsumerHandler{
		maxRetries: 3,
		retryStore: store,
		sendToDLQ: func(string, DLQMessage) error {
			t.Fatalf("unexpected dlq send")
			return nil
		},
		buffer: []msgWithSession{
			{msg: &sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 10}, sess: session},
			{msg: &sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 11}, sess: session},
			{msg: &sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 12}, sess: session},
		},
	}

	handler.handlePartialFailure([]int{1})

	wantMarked := []markedMessage{{topic: "orders", partition: 0, offset: 10}}
	if !reflect.DeepEqual(session.marked, wantMarked) {
		t.Fatalf("marked = %+v, want %+v", session.marked, wantMarked)
	}

	if got := store.Count(msgKey(&sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 11})); got != 1 {
		t.Fatalf("retry count = %d, want 1", got)
	}
}

func TestBatchConsumerHandler_HandlePartialFailure_AcksOtherPartitions(t *testing.T) {
	failedSession := &mockConsumerGroupSession{}
	successSession := &mockConsumerGroupSession{}
	handler := &BatchConsumerHandler{
		maxRetries: 3,
		retryStore: newMemoryRetryStore(),
		sendToDLQ: func(string, DLQMessage) error {
			t.Fatalf("unexpected dlq send")
			return nil
		},
		buffer: []msgWithSession{
			{msg: &sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 20}, sess: failedSession},
			{msg: &sarama.ConsumerMessage{Topic: "orders", Partition: 1, Offset: 30}, sess: successSession},
		},
	}

	handler.handlePartialFailure([]int{0})

	if len(failedSession.marked) != 0 {
		t.Fatalf("failed partition should not be acked, got %+v", failedSession.marked)
	}

	wantMarked := []markedMessage{{topic: "orders", partition: 1, offset: 30}}
	if !reflect.DeepEqual(successSession.marked, wantMarked) {
		t.Fatalf("success partition marked = %+v, want %+v", successSession.marked, wantMarked)
	}
}

func TestBatchConsumerHandler_HandlePartialFailure_AcksAfterDLQ(t *testing.T) {
	session := &mockConsumerGroupSession{}
	dlqCalls := 0
	store := newMemoryRetryStore()
	store.counts[msgKey(&sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 40})] = 2
	handler := &BatchConsumerHandler{
		maxRetries: 3,
		dlqTopic:   "orders-dlq",
		retryStore: store,
		sendToDLQ: func(topic string, msg DLQMessage) error {
			dlqCalls++
			if topic != "orders-dlq" {
				t.Fatalf("topic = %s, want orders-dlq", topic)
			}
			if msg.OriginalTopic != "orders" || msg.RetryCount != 3 {
				t.Fatalf("dlq msg = %+v", msg)
			}
			return nil
		},
		buffer: []msgWithSession{
			{msg: &sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 40, Value: []byte("bad")}, sess: session},
			{msg: &sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 41, Value: []byte("good")}, sess: session},
		},
	}

	handler.handlePartialFailure([]int{0})

	if dlqCalls != 1 {
		t.Fatalf("dlq calls = %d, want 1", dlqCalls)
	}

	wantMarked := []markedMessage{
		{topic: "orders", partition: 0, offset: 40},
		{topic: "orders", partition: 0, offset: 41},
	}
	if !reflect.DeepEqual(session.marked, wantMarked) {
		t.Fatalf("marked = %+v, want %+v", session.marked, wantMarked)
	}

	if got := store.Count(msgKey(&sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 40})); got != 0 {
		t.Fatalf("retry count for dlq message should be cleared")
	}
}

func TestBatchConsumerHandler_HandlePartialFailure_RetryCountSharedAcrossHandlers(t *testing.T) {
	store := newMemoryRetryStore()
	message := &sarama.ConsumerMessage{Topic: "orders", Partition: 0, Offset: 50, Value: []byte("bad")}

	for i := 0; i < 2; i++ {
		session := &mockConsumerGroupSession{}
		handler := &BatchConsumerHandler{
			maxRetries: 3,
			retryStore: store,
			sendToDLQ: func(string, DLQMessage) error {
				t.Fatalf("unexpected dlq send before max retries")
				return nil
			},
			buffer: []msgWithSession{{msg: message, sess: session}},
		}

		handler.handlePartialFailure([]int{0})

		if len(session.marked) != 0 {
			t.Fatalf("attempt %d should not ack before max retries", i+1)
		}
	}

	dlqCalls := 0
	session := &mockConsumerGroupSession{}
	handler := &BatchConsumerHandler{
		maxRetries: 3,
		dlqTopic:   "orders-dlq",
		retryStore: store,
		sendToDLQ: func(string, DLQMessage) error {
			dlqCalls++
			return nil
		},
		buffer: []msgWithSession{{msg: message, sess: session}},
	}

	handler.handlePartialFailure([]int{0})

	if dlqCalls != 1 {
		t.Fatalf("dlq calls = %d, want 1", dlqCalls)
	}
	if len(session.marked) != 1 || session.marked[0].offset != 50 {
		t.Fatalf("marked = %+v, want offset 50 acked after dlq", session.marked)
	}
}

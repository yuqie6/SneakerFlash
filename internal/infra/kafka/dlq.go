package kafka

import (
	"encoding/json"
	"log"
	"time"
)

// DLQMessage 死信消息结构，包含原始消息和元数据
type DLQMessage struct {
	OriginalTopic string `json:"original_topic"` // 原始主题
	OriginalValue []byte `json:"original_value"` // 原始消息体
	RetryCount    int    `json:"retry_count"`    // 重试次数
	LastError     string `json:"last_error"`     // 最后一次错误
	FailedAt      int64  `json:"failed_at"`      // 失败时间戳 (Unix ms)
}

// SendToDLQ 将失败消息投递到死信主题
func SendToDLQ(dlqTopic string, msg DLQMessage) error {
	if dlqTopic == "" {
		log.Printf("[WARN] DLQ topic not configured, skipping DLQ send")
		return nil
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal DLQ message: %v", err)
		return err
	}

	err = Send(dlqTopic, string(msgBytes))
	if err != nil {
		log.Printf("[ERROR] Failed to send message to DLQ topic %s: %v", dlqTopic, err)
		return err
	}

	log.Printf("[INFO] Message sent to DLQ topic %s. Original topic: %s, Retry count: %d",
		dlqTopic, msg.OriginalTopic, msg.RetryCount)
	return nil
}

// NewDLQMessage 创建死信消息
func NewDLQMessage(originalTopic string, originalValue []byte, retryCount int, lastError string) DLQMessage {
	return DLQMessage{
		OriginalTopic: originalTopic,
		OriginalValue: originalValue,
		RetryCount:    retryCount,
		LastError:     lastError,
		FailedAt:      time.Now().UnixMilli(),
	}
}

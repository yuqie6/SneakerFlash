package kafka

import (
	"SneakerFlash/internal/config"
	"context"
	"log"

	"github.com/IBM/sarama"
)

type MessageHandler func(msg []byte) error

func StartConsumer(cfg config.KafkaConfig, handler MessageHandler) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // 从最新的消息开始消费

	// 创建消费组
	groupID := "sneaker-group"
	group, err := sarama.NewConsumerGroup(cfg.Brokers, groupID, config)
	if err != nil {
		log.Fatalf("[ERROR] 创建消费组失败: %v", err)
	}

	defer group.Close()

	// 循环消费
	ctx := context.Background()
	comsumer := &ConsumerHandler{
		callback: handler,
	}

	log.Println("[INFO] Worker 正在监听 kafka topic:", cfg.Topic)

	for {
		topics := []string{cfg.Topic}
		err := group.Consume(ctx, topics, comsumer)
		if err != nil {
			log.Printf("[ERROR] 消费异常: %v", err)
		}
	}
}

type ConsumerHandler struct {
	callback MessageHandler
}

func (h *ConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *ConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *ConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		err := h.callback(msg.Value)

		if err != nil {
			log.Printf("[ERROR] 消息处理失败: %v", err)
		}

		// 标记已消费
		sess.MarkMessage(msg, "")
	}
	return nil
}

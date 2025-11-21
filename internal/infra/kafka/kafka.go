package kafka

import (
	"SneakerFlash/internal/config"
	"log"

	"github.com/IBM/sarama"
)

var Producer sarama.SyncProducer

func InitProducer(cfg config.KafkaConfig) {
	config := sarama.NewConfig()

	// 等待所有leader 和follower 确认才算发送成功
	config.Producer.RequiredAcks = sarama.WaitForAll

	// 将消息随机打到各个partition, 防止热点问题
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	// 成功的消息必须返回 success
	config.Producer.Return.Successes = true

	// 建立kafka连接
	client, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		log.Fatalf("kafka producer 连接失败: %s", err)
	}

	Producer = client
	log.Println("kafka producer 连接成功")
}

func Send(topic, message string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := Producer.SendMessage(msg)
	if err != nil {
		return err
	}
	log.Println("消息发送成功", topic, partition, offset)
	return nil
}

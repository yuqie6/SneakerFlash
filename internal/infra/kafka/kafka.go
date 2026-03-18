package kafka

import (
	"SneakerFlash/internal/config"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

var Producer sarama.SyncProducer
var Client sarama.Client

func InitProducer(cfg config.KafkaConfig) {
	saramaConfig := sarama.NewConfig()

	// 等待所有leader 和follower 确认才算发送成功
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll

	// 将消息随机打到各个partition, 防止热点问题
	saramaConfig.Producer.Partitioner = sarama.NewRandomPartitioner

	// 成功的消息必须返回 success
	saramaConfig.Producer.Return.Successes = true

	client, err := sarama.NewClient(cfg.Brokers, saramaConfig)
	if err != nil {
		log.Fatalf("kafka client 连接失败: %s", err)
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		_ = client.Close()
		log.Fatalf("kafka producer 连接失败: %s", err)
	}

	Client = client
	Producer = producer
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

func Ping(topic string) error {
	if Client == nil {
		return fmt.Errorf("kafka client not initialized")
	}

	if topic == "" {
		return fmt.Errorf("kafka topic is empty")
	}

	if err := Client.RefreshMetadata(topic); err != nil {
		return fmt.Errorf("refresh metadata failed: %w", err)
	}

	partitions, err := Client.Partitions(topic)
	if err != nil {
		return fmt.Errorf("load partitions failed: %w", err)
	}
	if len(partitions) == 0 {
		return fmt.Errorf("topic has no partitions")
	}

	return nil
}

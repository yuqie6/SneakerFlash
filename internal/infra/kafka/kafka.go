package kafka

import (
	"SneakerFlash/internal/config"
	"log"

	"github.com/IBM/sarama"
)

var Producer sarama.SyncProducer

func InitProducer(cfg config.KafkaConfig) {
	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.WaitForAll

	config.Producer.Partitioner = sarama.NewRandomPartitioner

	config.Producer.Return.Successes = true

	client, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		log.Fatalf("kafka producer 连接失败: %s", err)
	}

	Producer = client
	log.Println("kafka producer 连接成功")
}

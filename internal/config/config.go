package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstrcture:"server"`
	Data   DataConfig   `mapstrcture:"data"`
}

type ServerConfig struct {
	Port string `mapstrcture:"port"`
}

type DataConfig struct {
	Database DatabaseConfig `mapstrcture:"database"`
	Redis    RedisConfig    `mapstrcture:"redis"`
	Kafka    KafkaConfig    `mapstrcture:"kafka"`
}

type DatabaseConfig struct {
	Port        int       `mapstrcture:"port"`
	User        string    `mapstrcture:"user"`
	Password    string    `mapstrcture:"password"`
	DBname      string    `mapstrcture:"dbname"`
	Host        string    `mapstrcture:"host"`
	LogLever    int       `mapstrcture:"log_lever"`
	MaxIdle     int       `mapstrcture:"max_idle"`
	MaxOpen     int       `mapstrcture:"max_open"`
	MaxLiftTime time.Time `mapstrcture:"max_lift_time"`
}

type RedisConfig struct {
	Addr     string `mapstrcture:"addr"`
	Password string `mapstrcture:"password"`
	DB       int    `mapstrcture:"db"`
}

type KafkaConfig struct {
	Brokers []string `mapstrcture:"brokers"`
	Topic   string   `mapstrcture:"topic"`
}

var Conf Config

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("找不到配置文件")
		} else {
			log.Fatalf("配置文件加载失败: %v", err)
		}
	}

	if err := viper.Unmarshal(&Conf); err != nil {
		log.Fatal("序列化配置文件失败")
	}

	log.Println("配置文件加载成功!")
}

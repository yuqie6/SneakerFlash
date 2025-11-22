package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Data   DataConfig   `mapstructure:"data"`
	JWT    JWTConfig    `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port      string `mapstructure:"port"`
	MachineID int    `mapstructure:"machineid"`
}

type DataConfig struct {
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
}

type JWTConfig struct {
	Secret  string `mapstructure:"secret"`
	Expried int    `mapstructure:"expried"`
}

type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBname      string `mapstructure:"dbname"`
	LogLever    int    `mapstructure:"log_lever"`
	MaxIdle     int    `mapstructure:"max_idle"`
	MaxOpen     int    `mapstructure:"max_open"`
	MaxLifetime int    `mapstructure:"max_lifetime"`
}

type RedisConfig struct {
	Addr        string `mapstructure:"addr"`
	Password    string `mapstructure:"password"`
	DB          int    `mapstructure:"db"`
	PoolSize    int    `mapstructure:"pool_size"`
	MinIdle     int    `mapstructure:"min_idle"`
	ConnTimeout int    `mapstructure:"conn_timeout"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

var Conf Config

func Init() {
	configFile := os.Getenv("SNEAKERFLASH_CONFIG")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yml")
		viper.AddConfigPath("./")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("SNEAKERFLASH")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

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

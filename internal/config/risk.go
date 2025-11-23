package config

type RiskConfig struct {
	// 接口级限流
	LoginRate    RateLimitConfig `mapstructure:"login_rate"`
	SeckillRate  RateLimitConfig `mapstructure:"seckill_rate"`
	PayRate      RateLimitConfig `mapstructure:"pay_rate"`
	ProductRate  RateLimitConfig `mapstructure:"product_rate"`
	Enable       bool            `mapstructure:"enable"`
	HotspotBurst int             `mapstructure:"hotspot_burst"` // 默认热点 burst
}

type RateLimitConfig struct {
	Rate  int `mapstructure:"rate"`  // 每秒令牌
	Burst int `mapstructure:"burst"` // 桶大小
}

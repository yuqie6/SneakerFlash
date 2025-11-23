package middlerware

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/gjson"
)

// RedisTokenBucket 使用 Lua 保证原子性
const redisLua = `
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local ttl = tonumber(ARGV[5])

-- tokens key
local key = KEYS[1]
local last_tokens = tonumber(redis.call("HGET", key, "tokens")) or burst
local last_time = tonumber(redis.call("HGET", key, "time")) or now

local delta = math.max(0, now - last_time)
local filled = math.min(burst, last_tokens + delta * rate)
local allowed = filled >= requested
local new_tokens = filled
if allowed then
  new_tokens = filled - requested
end

redis.call("HSET", key, "tokens", new_tokens)
redis.call("HSET", key, "time", now)
redis.call("EXPIRE", key, ttl)

return allowed
`

type LimitConfig struct {
	KeyPrefix string
	Rate      int
	Burst     int
	TTL       int // 过期时间秒
}

func rateLimited(c *gin.Context, msg string) {
	appG := app.Gin{C: c}
	appG.ErrorMsg(http.StatusTooManyRequests, e.RISK_LIMITED, msg)
	c.Abort()
}

func buildKey(prefix, base string, c *gin.Context) string {
	if uidAny, ok := c.Get("userID"); ok {
		return fmt.Sprintf("%s:uid:%v:%s", prefix, uidAny, base)
	}
	if ip := clientIP(c); ip != "" {
		return fmt.Sprintf("%s:ip:%s:%s", prefix, ip, base)
	}
	return fmt.Sprintf("%s:%s", prefix, base)
}

// InterfaceLimiter 针对固定 key 的限流
func InterfaceLimiter(rdb *redis.Client, cfg LimitConfig, msg string) gin.HandlerFunc {
	if cfg.Rate <= 0 || cfg.Burst <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		key := buildKey(cfg.KeyPrefix, c.FullPath(), c)
		if ok := allow(rdb, key, cfg); !ok {
			rateLimited(c, msg)
			return
		}
		c.Next()
	}
}

// ParamLimiter 针对参数值限流，如 product_id
func ParamLimiter(rdb *redis.Client, cfg LimitConfig, param string, msg string) gin.HandlerFunc {
	if cfg.Rate <= 0 || cfg.Burst <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		val := extractParam(c, param)
		if val == "" {
			c.Next()
			return
		}
		key := buildKey(cfg.KeyPrefix, val, c)
		if ok := allow(rdb, key, cfg); !ok {
			rateLimited(c, msg)
			return
		}
		c.Next()
	}
}

func allow(rdb *redis.Client, key string, cfg LimitConfig) bool {
	now := time.Now().Unix()
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 120
	}
	status, err := rdb.Eval(context.Background(), redisLua, []string{key}, cfg.Rate, cfg.Burst, now, 1, ttl).Bool()
	if err != nil {
		// 失败时不阻断请求，避免误伤
		return true
	}
	return status
}

// BlackListMiddleware 简单黑名单校验（IP/UserID）
func BlackListMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		appG := app.Gin{C: c}
		ip := clientIP(c)
		if ip != "" {
			in, _ := rdb.SIsMember(context.Background(), "risk:ip:black", ip).Result()
			if in {
				appG.ErrorMsg(http.StatusTooManyRequests, e.RISK_BLACK, "IP 被限制")
				c.Abort()
				return
			}
		}
		if uidAny, ok := c.Get("userID"); ok {
			uid := fmt.Sprintf("%v", uidAny)
			in, _ := rdb.SIsMember(context.Background(), "risk:user:black", uid).Result()
			if in {
				appG.ErrorMsg(http.StatusTooManyRequests, e.RISK_BLACK, "账号被限制")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func clientIP(c *gin.Context) string {
	h := c.GetHeader("X-Forwarded-For")
	if h != "" {
		parts := strings.Split(h, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	return c.ClientIP()
}

func extractParam(c *gin.Context, name string) string {
	val := c.PostForm(name)
	if val != "" {
		return strings.TrimSpace(val)
	}
	val = c.Query(name)
	if val != "" {
		return strings.TrimSpace(val)
	}
	val = c.Param(name)
	if val != "" {
		return strings.TrimSpace(val)
	}

	body := readBodyOnce(c)
	if body != "" {
		res := gjson.Get(body, name)
		if res.Exists() {
			return strings.TrimSpace(res.String())
		}
	}
	return ""
}

// readBodyOnce 读取 body 并复位，避免影响后续 handler
func readBodyOnce(c *gin.Context) string {
	const key = "_cached_body"
	if cached, ok := c.Get(key); ok {
		if s, ok := cached.(string); ok {
			return s
		}
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	s := string(body)
	c.Request.Body = io.NopCloser(strings.NewReader(s))
	c.Set(key, s)
	return s
}

// Helper 根据配置生成接口限流配置
func BuildLimit(rateCfg config.RateLimitConfig, prefix string, ttl int) LimitConfig {
	return LimitConfig{
		KeyPrefix: prefix,
		Rate:      rateCfg.Rate,
		Burst:     rateCfg.Burst,
		TTL:       ttl,
	}
}

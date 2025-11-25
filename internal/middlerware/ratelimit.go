package middlerware

import (
	"SneakerFlash/internal/config"
	"SneakerFlash/internal/pkg/app"
	"SneakerFlash/internal/pkg/e"
	"context"
	"fmt"
	"io"
	"log/slog"
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

-- Redis 不支持直接返回 boolean，false 会变成 nil，故转为数字 1/0
if allowed then
  return 1
else
  return 0
end
`

// LimitConfig 定义令牌桶限流参数。
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

func buildKeys(prefix, base string, c *gin.Context) []string {
	keys := []string{fmt.Sprintf("%s:%s", prefix, base)} // 全局维度优先防爆款
	if uidAny, ok := c.Get("userID"); ok {
		keys = append(keys, fmt.Sprintf("%s:uid:%v:%s", prefix, uidAny, base))
	}
	if dev := deviceID(c); dev != "" {
		keys = append(keys, fmt.Sprintf("%s:dev:%s:%s", prefix, dev, base))
	}
	if ip := clientIP(c); ip != "" {
		keys = append(keys, fmt.Sprintf("%s:ip:%s:%s", prefix, ip, base))
	}
	if len(keys) == 0 {
		keys = append(keys, fmt.Sprintf("%s:%s", prefix, base))
	}
	return keys
}

// InterfaceLimiter 针对接口路径做桶限流，优先按 userID，其次按 IP。
func InterfaceLimiter(rdb *redis.Client, cfg LimitConfig, msg string) gin.HandlerFunc {
	if cfg.Rate <= 0 || cfg.Burst <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		for _, key := range buildKeys(cfg.KeyPrefix, c.FullPath(), c) {
			if ok := allow(rdb, key, cfg); !ok {
				rateLimited(c, msg)
				return
			}
		}
		c.Next()
	}
}

// ParamLimiter 针对参数值限流（如 product_id），配合 buildKey 兼容用户/IP 维度。
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
		for _, key := range buildKeys(cfg.KeyPrefix, val, c) {
			if ok := allow(rdb, key, cfg); !ok {
				rateLimited(c, msg)
				return
			}
		}
		c.Next()
	}
}

// allow 执行 Lua 令牌桶，失败时放行以避免误杀。
func allow(rdb *redis.Client, key string, cfg LimitConfig) bool {
	now := time.Now().Unix()
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 120
	}
	result, err := rdb.Eval(context.Background(), redisLua, []string{key}, cfg.Rate, cfg.Burst, now, 1, ttl).Int()
	if err != nil {
		// Redis 异常时记录告警，默认放行避免误杀
		slog.Warn("限流脚本执行失败", slog.Any("err", err), slog.String("key", key))
		return true
	}
	return result == 1
}

// BlackListMiddleware 简单黑名单校验（IP/UserID），命中则直接拒绝。
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

// GrayListMiddleware 灰名单命中直接返回限流响应，可按需插拔。
func GrayListMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		appG := app.Gin{C: c}
		ip := clientIP(c)
		if ip != "" {
			in, _ := rdb.SIsMember(context.Background(), "risk:ip:gray", ip).Result()
			if in {
				appG.ErrorMsg(http.StatusTooManyRequests, e.RISK_LIMITED, "灰名单限制")
				c.Abort()
				return
			}
		}
		if uidAny, ok := c.Get("userID"); ok {
			uid := fmt.Sprintf("%v", uidAny)
			in, _ := rdb.SIsMember(context.Background(), "risk:user:gray", uid).Result()
			if in {
				appG.ErrorMsg(http.StatusTooManyRequests, e.RISK_LIMITED, "灰名单限制")
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

// deviceID 从请求头/查询获取设备标识，作为额外限流维度。
func deviceID(c *gin.Context) string {
	if v := strings.TrimSpace(c.GetHeader("X-Device-ID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.Query("device_id")); v != "" {
		return v
	}
	return ""
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

// readBodyOnce 读取 body 并复位，避免影响后续 handler。
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

// BuildLimit 将配置转换为限流器的内部配置。
func BuildLimit(rateCfg config.RateLimitConfig, prefix string, ttl int) LimitConfig {
	return LimitConfig{
		KeyPrefix: prefix,
		Rate:      rateCfg.Rate,
		Burst:     rateCfg.Burst,
		TTL:       ttl,
	}
}

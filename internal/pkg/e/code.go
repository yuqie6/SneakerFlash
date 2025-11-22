// 定义业务错误码
package e

// 错误码
const (
	SUCCESS        = 200
	ERROR          = 500
	ERROR_SYSTEM   = 501
	INVALID_PARAMS = 400
	UNAUTHORIZED   = 401
	RATE_LIMIT     = 429

	// 风控错误 7xx
	RISK_BLOCKED = 700

	// 用户模块错误 100xx
	ERROR_EXIST_USER            = 10001
	ERROR_NOT_EXIST_USER        = 10002
	ERROR_AUTH_CHECK_TOKEN_FAIL = 10003
	ERROR_AUTH_TOKEN            = 10004

	// 商品错误 200xx
	ERROR_NOT_EXIST_PRODUCT = 20001

	// 秒杀错误 300xx
	ERROR_SECKILL_FULL     = 30001
	ERROR_REPEAT_BUY       = 30002
	ERROR_TOO_MANY_REQUEST = 30003
)

var Msglags = map[int]string{
	SUCCESS:        "ok",
	ERROR:          "系统开小差了，请稍后重试",
	ERROR_SYSTEM:   "系统繁忙，请稍后重试",
	INVALID_PARAMS: "请求参数错误",
	UNAUTHORIZED:   "未登录或token已失效",
	RATE_LIMIT:     "请求过于频繁，请稍后再试",
	RISK_BLOCKED:   "触发风控，暂时无法操作",

	ERROR_EXIST_USER:            "用户已存在",
	ERROR_NOT_EXIST_USER:        "用户不存在",
	ERROR_AUTH_CHECK_TOKEN_FAIL: "token 校验失败",
	ERROR_AUTH_TOKEN:            "token 生成失败",

	ERROR_NOT_EXIST_PRODUCT: "商品不存在",

	ERROR_SECKILL_FULL:     "手慢无，商品已售罄",
	ERROR_REPEAT_BUY:       "您已经抢购过该商品",
	ERROR_TOO_MANY_REQUEST: "请求过于频繁，请稍后再试",
}

func GetMsg(code int) string {
	msg, ok := Msglags[code]
	if ok {
		return msg
	}
	return Msglags[ERROR]
}

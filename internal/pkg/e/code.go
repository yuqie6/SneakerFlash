// 定义业务错误码
package e

// 错误码
const (
	SUCCESS        = 200
	ERROR          = 500
	INVALID_PARAMS = 400

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
	INVALID_PARAMS: "请求参数错误",

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

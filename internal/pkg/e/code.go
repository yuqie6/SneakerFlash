// 定义业务错误码
package e

// 错误码
const (
	SUCCESS         = 200
	ERROR           = 500
	INVAILID_PARAMS = 400

	// 用户模块错误 100xx
	ERROR_EXIST_USER            = 10001
	ERROR_NOT_EXIST_USER        = 10002
	ERROR_AUTH_CHECK_TOKEN_FAIL = 10003
	ERROR_AUTH_TOKEN            = 10004

	// 商品错误 200xx
	ERROR_NOT_EXIST_PTODUCT = 20001

	// 秒杀错误 300xx
	ERROR_SECKILL_FULL     = 30001
	ERROR_REPEAT_BUY       = 30002
	ERROR_TOO_MANY_REQUEST = 30003
)

var Msglags = map[int]string{
	SUCCESS:         "ok",
	ERROR:           "fail",
	INVAILID_PARAMS: "请求参数错误",

	ERROR_EXIST_USER:     "用户已存在",
	ERROR_NOT_EXIST_USER: "用户不存在",

	ERROR_SECKILL_FULL: "手慢无",
	ERROR_REPEAT_BUY:   "您已经抢购过该商品",
}

func GetMsg(code int) string {
	msg, ok := Msglags[code]
	if ok {
		return msg
	}
	return Msglags[ERROR]
}

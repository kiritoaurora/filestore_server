package common

// 错误码
type ErrorCode int32

const (
	_ int32 = iota + 9999
	// 正常
	StatusOK
	// 请求参数无效
	StatusParamInvalid
	// 服务出错
	StatusServerError
	// 注册失败
	StatusRegisterFailed
	// 登陆失败
	StatusLoginFailed
	// token无效
	StatusTokenInvalid
	// 用户不存在
	StatusUserNotExists
)
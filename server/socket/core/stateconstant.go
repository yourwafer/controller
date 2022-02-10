package core

const (
	// 请求指令不存在
	COMMAND_NOT_FOUND = 17

	// 解码异常
	DECODE_EXCEPTION = 18

	// 编码异常
	ENCODE_EXCEPTION = 19

	// 会话身份异常
	IDENTITY_EXCEPTION = 22

	// 业务异常(此异常，将会把错误码通过Result结构返回给客户端)
	MANAGED_EXCEPTION = 23

	// session中缓存值异常
	SESSION_EXCEPTION = 24

	// 未知异常
	UNKNOWN_EXCEPTION = 25

	// 网络相关异常
	SOCKET_EXCEPTION = 26

	// 需要管理后台IP
	MANAGE_IP_EXCEPTION = 27
)

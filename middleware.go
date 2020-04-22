package driver

var (
	hid          string           //硬件id
	path         string           //链接地址
	sendTopic    string           //发送主题（mqtt）
	version      string           //当前版本
	authUserName string           //中间件认证用户名
	authPassword string           //中间件认证密码
	handleMsg    func(msg string) //消息处理方法
)

type IMessageMiddleWare interface {
	Init(url, id, send, ver, username, password string, callback func(msg string)) error
	Response(msg string) error
}

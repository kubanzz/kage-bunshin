package routers

import (
	"kage-bunshin/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	// beego.Router("/", &controllers.MainController{})
	beego.Router("/", &controllers.WeChatController{}, "*:ServerWechat")
	beego.AutoRouter(&controllers.WeChatController{})
	beego.AutoRouter(&controllers.CollectorController{})
}

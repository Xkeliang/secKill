package router

import (
	"Seckill/SecProxy/controller"
	"github.com/astaxie/beego/logs"
	"github.com/beego"
)

func init() {
	logs.Debug("enter router init")
	beego.Router("/seckill", &controller.SkillController{}, "*:SecKill")
	beego.Router("/secinfo", &controller.SkillController{}, "*:SecInfo")
}

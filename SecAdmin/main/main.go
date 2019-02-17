package main

import (
	_"Seckill/SecAdmin/router"
	"github.com/astaxie/beego"
	"fmt"
	"github.com/astaxie/beego/logs"
)

func main (){
	fmt.Println(beego.BConfig.WebConfig.ViewsPath)
	//app :=beego.SetViewsPath("D:\\blockchain\\Golang\\src\\Seckill\\SecAdmin\\views")
	//fmt.Println(beego.BConfig.WebConfig.ViewsPath)
	//fmt.Println(*app)
	//fmt.Println(beego.AppPath)
	//init 配置，mysql、etcd、logs
	err := initAll()
	if err != nil{
		panic(fmt.Sprintf("init database failed,err:%v",err))
		return

	}
	logs.Debug("initAll success")
	//fmt.Println(beego.BConfig.AppName)
	beego.Run()
}


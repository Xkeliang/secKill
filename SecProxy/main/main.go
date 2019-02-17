package main

import "github.com/beego"

func main() {
	//读取配置信息
	err := initConfig()
	if err != nil {
		panic(err)
		return
	}
	//初始化系统
	err = initSec()
	if err != nil {
		panic(err)
		return
	}
	//开始监控etcd服务
	initSetProductWatcher()
	beego.Run()
}

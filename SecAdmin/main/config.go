package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type EtcdConf struct {
	Addr string
	EtcdKeyPrefix string
	ProductKey string
	Timeout int
}

type MysqlConfig struct {
	UserName string
	Passwd string
	Port int
	Database string
	Host string
}

var (
	AppConf Config
)

type Config struct {
	mysqlConf MysqlConfig
	etcdConf EtcdConf
}

func initConfig()(err error)  {
	username :=beego.AppConfig.String("mysqlUserName")
	if len(username) == 0{
		logs.Error("load config of mysqlUserName failed,is null")
		return
	}
	mysqlPasswd :=beego.AppConfig.String("mysqlPasswd")
	if len(username) == 0{
		logs.Error("load config of mysqlPasswd failed,is null")
		return
	}
	mysqlHost :=beego.AppConfig.String("mysqlHost")
	if len(username) == 0{
		logs.Error("load config of mysqlHost failed,is null")
		return
	}
	mysqlDatabase :=beego.AppConfig.String("mysqlDatabase")
	if len(username) == 0{
		logs.Error("load config of mysqlDatabase failed,is null")
		return
	}
	mysqlPort,err := beego.AppConfig.Int("mysqlPort")
	if err != nil {
		logs.Error("load config of mysqlPort failed,err :%v",err)
	}
	AppConf.mysqlConf.UserName=username
	AppConf.mysqlConf.Passwd=mysqlPasswd
	AppConf.mysqlConf.Host=mysqlHost
	AppConf.mysqlConf.Database=mysqlDatabase
	AppConf.mysqlConf.Port=mysqlPort


	addr :=beego.AppConfig.String("etcdAddr")
	if len(addr) == 0{
		logs.Error("load config of etcdAddr failed,is null")
		return
	}
	prefix :=beego.AppConfig.String("etcdSecKeyPrefix")
	if len(prefix) == 0{
		logs.Error("load config of etcdSecKeyPrefix failed,is null")
		return
	}
	productKey :=beego.AppConfig.String("productSecKey")
	if len(productKey) == 0{
		logs.Error("load config of productSecKey failed,is null")
		return
	}
	timeout,err := beego.AppConfig.Int("etcdTimeout")
	if err != nil {
		logs.Error("load config of etcdTimeout failed,err :%v",err)
	}

	AppConf.etcdConf.Addr=addr
	AppConf.etcdConf.EtcdKeyPrefix=prefix
	AppConf.etcdConf.ProductKey=productKey
	AppConf.etcdConf.Timeout=timeout
	return
}


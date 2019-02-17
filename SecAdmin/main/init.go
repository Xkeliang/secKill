package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/astaxie/beego/logs"
	"fmt"
	"Seckill/SecAdmin/model"
	"time"
	etcd_Client "go.etcd.io/etcd/clientv3"
)

var Db *sqlx.DB
var EtcdClient *etcd_Client.Client

//
func initDb()(err error)  {
	dns := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",AppConf.mysqlConf.UserName,
		AppConf.mysqlConf.Passwd,AppConf.mysqlConf.Host,AppConf.mysqlConf.Port,AppConf.mysqlConf.Database)
	database,err := sqlx.Open("mysql",dns)
	if err != nil {
		logs.Error("open mysql failed,err ：%v",err)
		return
	}
	Db = database
	logs.Debug("success mysql success,Db:%v",Db)
	return
}

func initEtcd()(err error){
	cli,err:= etcd_Client.New(etcd_Client.Config{
		Endpoints : []string{AppConf.etcdConf.Addr},
		DialTimeout : time.Duration(AppConf.etcdConf.Timeout)*time.Second,
	})
	if err != nil {
		logs.Error("connect etcd error : /%v",err)
		return
	}
	EtcdClient = cli
	logs.Debug("init etcd succ")
	return
}


func initAll()(err error)  {
	err = initLogs()
	if err != nil {
		logs.Warn("init logs failed,err:",err)
		return
	}
	err = initConfig()
	if err != nil {
		logs.Warn("init config failed,err :%v",err)
		return
	}
	logs.Debug("load Config sucess,appconf:%v",AppConf)

	err = initDb()
	if err != nil {
		logs.Warn("init Db failed,err :%v",err)
		return
	}
	err = initEtcd()
	if err != nil {
		logs.Warn("init etcd failed,err: %v",err)
		return
	}
	//初始化的etcd、mysql传到model包
	err = model.InitDB(Db,EtcdClient,AppConf.etcdConf.EtcdKeyPrefix,AppConf.etcdConf.ProductKey)
	if err != nil {
		logs.Warn("init model failed,err :%v",err)
		return
	}
	return
}

//配置log
func initLogs()(err error) {
	err =logs.SetLogger(logs.AdapterFile,`{"filename":"logs/project.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":10,"color":true}`)
	logs.EnableFuncCallDepth(true)
	return
}
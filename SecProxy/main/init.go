package main

import (
	"Seckill/SecProxy/service"
	"context"
	"encoding/json"
	"fmt"
	"github.com/beego/logs"
	"github.com/garyburd/redigo/redis"
	etcd_Client "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

var redisPool *redis.Pool
var etcdClient *etcd_Client.Client

//初始化redis
func initRadis() (err error) {
	redisPool = &redis.Pool{
		MaxIdle:     secKillConf.RedisBlackConf.RedisMaxIdle,
		MaxActive:   secKillConf.RedisBlackConf.RedisMaxActive,
		IdleTimeout: time.Duration(secKillConf.RedisBlackConf.RedisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.RedisBlackConf.RedisAddr)
		},
	}
	//获取连接检测连接是否成功
	conn := redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("ping")
	if err != nil {
		logs.Error("ping redis err %v", err)
		return
	}
	return
}

//初始化日志文件
func initLog() (err error) {
	config := make(map[string]interface{})
	config["filename"] = secKillConf.LogPath
	config["level"] = converLogLevel(secKillConf.LogLevel)

	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println("marshar failed,err:", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	logs.EnableFuncCallDepth(true)
	logs.SetLogFuncCallDepth(3)
	return
}

//初始化etcd
func initEtcd() (err error) {
	cli, err := etcd_Client.New(etcd_Client.Config{
		Endpoints:   []string{secKillConf.EtcdConf.EtcdAddr},
		DialTimeout: time.Duration(secKillConf.EtcdConf.EtcdTimeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd error : /%v", err)
		return
	}
	etcdClient = cli
	return
}

//监控更新
func initSetProductWatcher() {
	go watchSecProductKey(secKillConf.EtcdConf.EtcdProductSecKey)
}

func watchSecProductKey(key string) {
	cli, err := etcd_Client.New(etcd_Client.Config{
		Endpoints:   []string{secKillConf.EtcdConf.EtcdAddr, "localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: time.Duration(secKillConf.EtcdConf.EtcdTimeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd error : /%v", err)
		return
	}
	logs.Debug("begin watch key:%s", key)
	for {
		rch := cli.Watch(context.Background(), key)
		var secProductInfo []service.SecProductInfoConf
		var getConfSucc = true
		for wresp := range rch {
			//var ev *mvccpb.Event
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					logs.Warn("key[%s] 's config delete", key)
					continue
				}
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == key {
					err = json.Unmarshal(ev.Kv.Value, &secProductInfo)
					if err != nil {
						logs.Error("key[%s],Unmarshal %s,err  %v", key, ev.Kv.Key, err)
						getConfSucc = false
						continue
					}
					logs.Debug("get config from etcd %s %q:%q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				}
				if getConfSucc {
					logs.Debug("get config from etcd succ,%v", secProductInfo)
					updateProductInfo(secProductInfo)
				}
			}
		}
	}
}

//加载商品信息到MAP
func updateProductInfo(secProductInfo []service.SecProductInfoConf) {
	var tem map[int]*service.SecProductInfoConf = make(map[int]*service.SecProductInfoConf, 1024)
	for _, v := range secProductInfo {
		product := v
		fmt.Println(v)
		tem[v.ProductId] = &product
	}

	secKillConf.RwSecKillConfLock.Lock()
	secKillConf.SecProductInfoMap = tem
	secKillConf.RwSecKillConfLock.Unlock()
}

//初始化秒杀系统
func initSec() (err error) {

	err = initLog()
	if err != nil {
		logs.Error("init log failed,err:%v", err)
		return
	}

	err = initRadis()
	if err != nil {
		logs.Error("init radis failed,err : %v", err)
		return
	}
	err = initEtcd()
	if err != nil {
		logs.Error("init etcd failed,err : %v", err)
		return
	}
	err = loadSecConf()
	if err != nil {
		logs.Error("loadSecConf failed,err : %v", err)
		return
	}

	service.InitService(secKillConf)

	logs.Info("init sec success")
	return
}

//日志级别转换为beego的数字形式
func converLogLevel(level string) int {
	switch level {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}

	return logs.LevelDebug
}

/*
	加载秒杀物品
	从etcd 到 结构体
*/
func loadSecConf() (err error) {
	key := secKillConf.EtcdConf.EtcdProductSecKey
	resp, err := etcdClient.Get(context.Background(), key)
	if err != nil {
		logs.Error("get [%s] from etcd key error :%v", key, err)
		return
	}
	var secProductInfo []service.SecProductInfoConf
	for k, y := range resp.Kvs {
		logs.Debug("key[%d] value[%s]", k, y)
		err = json.Unmarshal(y.Value, &secProductInfo)
		if err != nil {
			logs.Error("Unmarshal failed:%v", err)
			return
		}
		logs.Debug("sec info succ is [%v]", secProductInfo)
	}
	logs.Debug("load secProductInfo", secProductInfo)
	updateProductInfo(secProductInfo)
	return
}

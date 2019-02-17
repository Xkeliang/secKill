package main

import (
	"Seckill/SecLayer/service"
	"fmt"
	"github.com/beego/config"
	"github.com/beego/logs"
	"strings"
)

var (
	appConfig *service.SecLayerConf
)

func initConfig(confType, filename string) (err error) {
	conf, err := config.NewConfig(confType, filename)
	if err != nil {
		logs.Error("new config failed,err:", err)
		return
	}

	appConfig = &service.SecLayerConf{}
	appConfig.LogLevel = conf.String("logs::logLevel")
	if len(appConfig.LogLevel) == 0 {
		appConfig.LogLevel = "debug"
	}
	appConfig.LogPath = conf.String("logs::logPath")
	if len(appConfig.LogPath) == 0 {
		appConfig.LogPath = "./logs"
	}
	//读取redis相关配置
	appConfig.Proxy2LayerRadis.RedisAddr = conf.String("redis::redisProxy2LayerAddr")
	if len(appConfig.Proxy2LayerRadis.RedisAddr) == 0 {
		logs.Error("read redisProxy2LayerAddr failed")
		err = fmt.Errorf("read redisProxy2LayerAddr failed")
		return
	}

	appConfig.Proxy2LayerRadis.RedisQueueName = conf.String("redis::redisProxy2LayerQueueName")
	if len(appConfig.Proxy2LayerRadis.RedisQueueName) == 0 {
		logs.Error("read redisProxy2LayerQueueName failed")
		err = fmt.Errorf("read redisProxy2LayerQueueName failed")
		return
	}

	appConfig.Proxy2LayerRadis.RedisMaxIdle, err = conf.Int("redis::redisProxy2LayerIdle")
	if err != nil {
		logs.Error("read redisProxy2LayerIdle failed,err:%v", err)
		return
	}

	appConfig.Proxy2LayerRadis.RedisIdleTimeout, err = conf.Int("redis::redisProxy2LayerIdleTimeout")
	if err != nil {
		logs.Error("read redisProxy2LayerIdleTimeout failed,err:%v", err)
		return
	}

	appConfig.Proxy2LayerRadis.RedisMaxActive, err = conf.Int("redis::redisProxy2LayerActive")
	if err != nil {
		logs.Error("read redisProxy2LayerActive failed,err:%v", err)
		return
	}

	//layer2Proxy相关配置
	appConfig.Layer2ProxyRadis.RedisAddr = conf.String("redis::redisLayer2ProxyAddr")
	if len(appConfig.Proxy2LayerRadis.RedisAddr) == 0 {
		logs.Error("read redisLayer2ProxyAddr failed")
		err = fmt.Errorf("read redisLayer2ProxyAddr failed")
		return
	}

	appConfig.Layer2ProxyRadis.RedisQueueName = conf.String("redis::redisLayer2ProxyQueueName")
	if len(appConfig.Proxy2LayerRadis.RedisQueueName) == 0 {
		logs.Error("read redisLayer2ProxyQueueName failed")
		err = fmt.Errorf("read redisLayer2ProxyQueueName failed")
		return
	}

	appConfig.Layer2ProxyRadis.RedisMaxIdle, err = conf.Int("redis::redisLayer2ProxyIdle")
	if err != nil {
		logs.Error("read redisLayer2ProxyIdle failed,err:%v", err)
		return
	}

	appConfig.Layer2ProxyRadis.RedisIdleTimeout, err = conf.Int("redis::redisLayer2ProxyIdleTimeout")
	if err != nil {
		logs.Error("read redisLayer2ProxyIdleTimeout failed,err:%v", err)
		return
	}

	appConfig.Layer2ProxyRadis.RedisMaxActive, err = conf.Int("redis::redisLayer2ProxyActive")
	if err != nil {
		logs.Error("read redisLayer2ProxyActive failed,err:%v", err)
		return
	}
	appConfig.WriteGoroutineNum, err = conf.Int("service::writeProxy2LayerGoroutineNum")
	if err != nil {
		logs.Error("read writeProxy2LayerGoroutineNum failed,err:%v", err)
		return
	}

	appConfig.ReadGoroutineNum, err = conf.Int("service::readProxy2LayerGoroutineNum")
	if err != nil {
		logs.Error("read readProxy2LayerGoroutineNum failed,err:%v", err)
		return
	}
	appConfig.HandleUserGoroutineNum, err = conf.Int("service::handleUserGoroutineNum")
	if err != nil {
		logs.Error("read handleUserGoroutineNum failed,err:%v", err)
		return
	}

	appConfig.Read2handleChanSize, err = conf.Int("service::read2handleChanSize")
	if err != nil {
		logs.Error("read Read2handleChanSize failed,err:%v", err)
		return
	}

	appConfig.MaxrequestWaitTimeout, err = conf.Int("service::maxrequestWaitTimeout")
	if err != nil {
		logs.Error("read maxrequestWaitTimeout failed,err:%v", err)
		return
	}

	appConfig.Handle2writeChanSize, err = conf.Int("service::handle2writeChanSize")
	if err != nil {
		logs.Error("read handle2writeChanSize failed,err:%v", err)
		return
	}

	appConfig.SendToWriteChanTimeout, err = conf.Int("service::sendToWriteChanTimeout")
	if err != nil {
		logs.Error("read sendToWriteChanTimeout failed,err:%v", err)
		return
	}
	appConfig.SendToHandleChanTimeout, err = conf.Int("service::sendToHandleChanTimeout")
	if err != nil {
		logs.Error("read sendToHandleChanTimeout failed,err:%v", err)
		return
	}

	appConfig.TokenPasswd = conf.String("service::seckillTokenPasswd")
	if len(appConfig.TokenPasswd) == 0 {
		logs.Error("read seckillTokenPasswd failed")
		err = fmt.Errorf("read seckillTokenPasswd failed")
		return
	}

	//etcd
	etcdAddr := conf.String("etcd::etcdAddr")
	if etcdAddr == "" {
		err = fmt.Errorf("init config faild etcd[%s]", etcdAddr)
		return
	}
	logs.Debug("read config succ, etcd addr:%v", etcdAddr)
	etcdTimeout, err := conf.Int("etcd::etcdTimeout")
	if err != nil {
		err = fmt.Errorf("init config faild,read etcdTimeout error :%v", err)
		return
	}
	etcdSecKeyPrefix := conf.String("etcd::etcdSecKeyPrefix")
	if etcdSecKeyPrefix == "" {
		err = fmt.Errorf("init config faild,read etcdSecKeyPrefix error")
		return
	}
	productSecKey := conf.String("etcd::productSecKey")
	if productSecKey == "" {
		err = fmt.Errorf("init config faild,read productSecKey error")
		return
	}
	if strings.HasSuffix(etcdSecKeyPrefix, "/") == false {
		etcdSecKeyPrefix += "/"
	}
	logs.Debug("read config success ecdtaddr :%v", etcdAddr)
	appConfig.EtcdConfig.EtcdAddr = etcdAddr
	appConfig.EtcdConfig.EtcdTimeout = etcdTimeout
	appConfig.EtcdConfig.EtcdSecKeyPrefix = etcdSecKeyPrefix
	appConfig.EtcdConfig.EtcdProductSecKey = fmt.Sprintf("%s%s", etcdSecKeyPrefix, productSecKey)
	fmt.Println(appConfig.EtcdConfig.EtcdProductSecKey)
	return
}

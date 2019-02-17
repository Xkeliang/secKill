package main

import (
	_ "Seckill/SecProxy/router"
	"Seckill/SecProxy/service"
	"fmt"
	"github.com/beego"
	"github.com/beego/logs"
	"strings"
)

//存储配置信息
var secKillConf = &service.SecKillConf{
	SecProductInfoMap: make(map[int]*service.SecProductInfoConf, 1024),
}

//读取配置信息
func initConfig() (err error) {
	///*
	//调试
	path := beego.AppPath
	fmt.Println(path)
	//*/
	/*
		redis白名单配置
	*/
	redisBlackAddr := beego.AppConfig.String("redisBlackAddr")
	if redisBlackAddr == "" {
		err = fmt.Errorf("init config faild redis[%s]", redisBlackAddr)
		return
	}
	logs.Debug("read config succ, redis addr:%v", redisBlackAddr)
	maxBlackIdle, err := beego.AppConfig.Int("redisBlackMaxIdle")
	if err != nil {
		err = fmt.Errorf("init config faild,read redisBlackMaxIdle error :%v", err)
	}
	maxBlackActive, err := beego.AppConfig.Int("redisBlackMaxActive")
	if err != nil {
		err = fmt.Errorf("init config faild,read redisBlackMaxActive error :%v", err)
	}
	idleBlackTimeout, err := beego.AppConfig.Int("redisBlackIdleTimeout")
	if err != nil {
		err = fmt.Errorf("init config faild,read redisIdleTimeout error :%v", err)
	}
	logs.Debug("read config success redisaddr :%v", redisBlackAddr)
	secKillConf.RedisBlackConf.RedisAddr = redisBlackAddr
	secKillConf.RedisBlackConf.RedisMaxIdle = maxBlackIdle
	secKillConf.RedisBlackConf.RedisMaxActive = maxBlackActive
	secKillConf.RedisBlackConf.RedisIdleTimeout = idleBlackTimeout

	/*
		etcd配置
	*/
	etcdAddr := beego.AppConfig.String("etcdAddr")
	if etcdAddr == "" {
		err = fmt.Errorf("init config faild etcd[%s]", etcdAddr)
		return
	}
	logs.Debug("read config succ, etcd addr:%v", etcdAddr)
	etcdTimeout, err := beego.AppConfig.Int("etcdTimeout")
	if err != nil {
		err = fmt.Errorf("init config faild,read etcdTimeout error :%v", err)
		return
	}
	etcdSecKeyPrefix := beego.AppConfig.String("etcdSecKeyPrefix")
	if etcdSecKeyPrefix == "" {
		err = fmt.Errorf("init config faild,read etcdSecKeyPrefix error")
		return
	}
	productSecKey := beego.AppConfig.String("productSecKey")
	if productSecKey == "" {
		err = fmt.Errorf("init config faild,read productSecKey error")
		return
	}
	if strings.HasSuffix(etcdSecKeyPrefix, "/") == false {
		etcdSecKeyPrefix += "/"
	}
	logs.Debug("read config success ecdtaddr :%v", etcdAddr)
	secKillConf.EtcdConf.EtcdAddr = etcdAddr
	secKillConf.EtcdConf.EtcdTimeout = etcdTimeout
	secKillConf.EtcdConf.EtcdSecKeyPrefix = etcdSecKeyPrefix
	secKillConf.EtcdConf.EtcdProductSecKey = fmt.Sprintf("%s%s", etcdSecKeyPrefix, productSecKey)
	/*
		日志配置
	*/
	logPath := beego.AppConfig.String("logPath")
	logLevel := beego.AppConfig.String("logLevel")
	secKillConf.LogPath = logPath
	secKillConf.LogLevel = logLevel
	/*
		cookie密钥配置
	*/
	secKillConf.CookieSecretKey = beego.AppConfig.String("cookieSecretKey")
	/*
		限制配置
	*/
	secLimit, err := beego.AppConfig.Int("userSecAccessLimit")
	if err != nil {
		err = fmt.Errorf("init config failed, read userSecAccessLimit error:%v", err)
		return
	}
	secKillConf.AccessLimitConf.UserSecAccessLimit = secLimit

	referWhitelist := beego.AppConfig.String("referWhitelist")
	if len(referWhitelist) > 0 {
		secKillConf.ReferWhiteList = strings.Split(referWhitelist, ",")
	}
	ipLimit, err := beego.AppConfig.Int("ipSecAccessLimit")
	if err != nil {
		err = fmt.Errorf("init config faild,read ipSecAccessLimit  error", err)
		return
	}
	secKillConf.AccessLimitConf.IPSecAccessLimit = ipLimit

	minIdLimit, err := beego.AppConfig.Int("userMinAccessLimit")
	if err != nil {
		err = fmt.Errorf("init config failed, read userMinAccessLimit error:%v", err)
		return
	}
	secKillConf.AccessLimitConf.UserMinAccessLimit = minIdLimit

	minIpLimit, err := beego.AppConfig.Int("ipMinAccessLimit")
	if err != nil {
		err = fmt.Errorf("init config failed, read ipMinAccessLimit error:%v", err)
		return
	}
	secKillConf.AccessLimitConf.IPMinAccessLimit = minIpLimit
	/*
		redis 接入层到逻辑层配置
	*/
	redisProxy2LayerAddr := beego.AppConfig.String("redisProxy2LayerAddr")
	if redisProxy2LayerAddr == "" {
		err = fmt.Errorf("init config faild redis[%s] config is null", redisProxy2LayerAddr)
		return
	}
	redisProxy2LayerIdle, err := beego.AppConfig.Int("redisProxy2LayerIdle")
	if err != nil {
		err = fmt.Errorf("init config faild,read redisProxy2LayerIdle error :%v", err)
	}
	redisProxy2LayerActive, err := beego.AppConfig.Int("redisProxy2LayerActive")
	if err != nil {
		err = fmt.Errorf("init config faild,read redisProxy2LayerActive error :%v", err)
	}
	idleProxy2LayerTimeout, err := beego.AppConfig.Int("redisProxy2LayerIdleTimeout")
	if err != nil {
		err = fmt.Errorf("init config faild,read redisIdleTimeout error :%v", err)
	}

	logs.Debug("read config success redisaddr :%v", redisProxy2LayerAddr)
	secKillConf.RedisProxy2LayerConf.RedisAddr = redisProxy2LayerAddr
	secKillConf.RedisProxy2LayerConf.RedisMaxIdle = redisProxy2LayerIdle
	secKillConf.RedisProxy2LayerConf.RedisMaxActive = redisProxy2LayerActive
	secKillConf.RedisProxy2LayerConf.RedisIdleTimeout = idleProxy2LayerTimeout
	/*
		redis 逻辑层到接入层配置
	*/
	writeProxy2LayerGoroutineNum, err := beego.AppConfig.Int("writeProxy2LayerGoroutineNum")
	if err != nil {
		err = fmt.Errorf("writeProxy2LayerGoroutineNum error :%v", err)
	}
	secKillConf.WriteProxy2LayerGoroutineNum = writeProxy2LayerGoroutineNum

	readLayer2ProxyGoroutineNum, err := beego.AppConfig.Int("readLayer2ProxyGoroutineNum")
	if err != nil {
		err = fmt.Errorf("readProxy2LayerGoroutineNum error :%v", err)
	}
	secKillConf.ReadLayer2ProxyGoroutineNum = readLayer2ProxyGoroutineNum

	redisLayer2ProxyAddr := beego.AppConfig.String("redisLayer2ProxyAddr")
	if redisLayer2ProxyAddr == "" {
		err = fmt.Errorf("init config faild redis[%s] config is null", redisProxy2LayerAddr)
		return
	}
	redisLayer2ProxyIdle, err := beego.AppConfig.Int("redisLayer2ProxyIdle")
	if err != nil {
		err = fmt.Errorf("init config faild,read redisLayer2ProxyIdle error :%v", err)
	}
	redisLayer2ProxyActive, err := beego.AppConfig.Int("redisLayer2ProxyActive")
	if err != nil {
		err = fmt.Errorf("init config faild,read redisLayer2ProxyActive error :%v", err)
	}
	idleLayer2ProxyTimeout, err := beego.AppConfig.Int("redisLayer2ProxyIdleTimeout")
	if err != nil {
		err = fmt.Errorf("init config faild,read redisIdleTimeout error :%v", err)
	}

	logs.Debug("read config success redisaddr :%v", redisLayer2ProxyAddr)
	secKillConf.RedisLayer2ProxyConf.RedisAddr = redisLayer2ProxyAddr
	secKillConf.RedisLayer2ProxyConf.RedisMaxIdle = redisLayer2ProxyIdle
	secKillConf.RedisLayer2ProxyConf.RedisMaxActive = redisLayer2ProxyActive
	secKillConf.RedisLayer2ProxyConf.RedisIdleTimeout = idleLayer2ProxyTimeout
	return err
}

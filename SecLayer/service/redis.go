package service

import (
	"github.com/beego/logs"
	"github.com/garyburd/redigo/redis"
	"time"
)

func initRedisPool(redisConf RedisConf) (pool *redis.Pool, err error) {
	pool = &redis.Pool{
		MaxIdle:     redisConf.RedisMaxIdle,
		MaxActive:   redisConf.RedisMaxActive,
		IdleTimeout: time.Duration(redisConf.RedisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisConf.RedisAddr)
		},
	}
	conn := pool.Get()
	defer conn.Close()

	_, err = conn.Do("ping")
	if err != nil {
		logs.Error("ping redis err %v", err)
		return
	}
	return
}

func initRedis(conf *SecLayerConf) (err error) {
	secLayerContext.Proxy2LayerRedisPool, err = initRedisPool(conf.Proxy2LayerRadis)
	if err != nil {
		logs.Error("init proxy2layerRadid failed,err :%v", err)
		return
	}
	secLayerContext.Layer2ProxyRedisPool, err = initRedisPool(conf.Layer2ProxyRadis)
	if err != nil {
		logs.Error("init layer2proxyRadis failed,err :%v", err)
		return
	}
	return
}
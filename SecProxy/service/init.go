package service

import (
	"github.com/beego/logs"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
	"fmt"
)

var (
	secKillConf *SecKillConf
)

func InitService(serviceConf *SecKillConf) (err error) {
	secKillConf = serviceConf
	err = loadBlackList()
	if err != nil {
		logs.Error("load black list err:%v", err)
		return
	}
	logs.Debug("init service succ,config:%v\n", secKillConf)

	err = initProxy2LayerRedis()
	if err != nil {
		logs.Error("load procy2layer redis pool failed list err:%v", err)
		return
	}
	err = initLayer2ProxyRedis()
	if err != nil {
		logs.Error("load layer2Proxy redis pool failed list err:%v", err)
		return
	}
	secKillConf.secLimitMgr = &SecLimitMgr{
		UserLimitMap: make(map[int]*Limit, 1000),
		IpLimitMap:   make(map[string]*Limit, 1000),
	}

	secKillConf.SecReqChan = make(chan *SecRequest, secKillConf.SecReqChanSize)
	secKillConf.UserConMap = make(map[string]chan *SecResult, 1000)

	initRedisProcessFunc()

	return
}

//创建协程读写redis
func initRedisProcessFunc() {
	for i := 0; i < secKillConf.WriteProxy2LayerGoroutineNum; i++ {
		go WriteHandle()
	}
	for i := 0; i < secKillConf.ReadLayer2ProxyGoroutineNum; i++ {
		go ReadHandle()
	}
	return
}


//init Redis
func initProxy2LayerRedis() (err error) {
	secKillConf.proxy2LayerRedisPool = &redis.Pool{
		MaxIdle:     secKillConf.RedisProxy2LayerConf.RedisMaxIdle,
		MaxActive:   secKillConf.RedisProxy2LayerConf.RedisMaxActive,
		IdleTimeout: time.Duration(secKillConf.RedisProxy2LayerConf.RedisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.RedisProxy2LayerConf.RedisAddr)
		},
	}
	conn := secKillConf.proxy2LayerRedisPool.Get()
	defer conn.Close()

	_, err = conn.Do("ping")
	if err != nil {
		logs.Error("ping redis err %v", err)
		return
	}

	return
}

func initLayer2ProxyRedis() (err error) {
	secKillConf.layer2ProxyRedisPool = &redis.Pool{
		MaxIdle:     secKillConf.RedisLayer2ProxyConf.RedisMaxIdle,
		MaxActive:   secKillConf.RedisLayer2ProxyConf.RedisMaxActive,
		IdleTimeout: time.Duration(secKillConf.RedisLayer2ProxyConf.RedisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.RedisLayer2ProxyConf.RedisAddr)
		},
	}

	conn := secKillConf.layer2ProxyRedisPool.Get()
	defer conn.Close()

	_, err = conn.Do("ping")

	if err != nil {
		logs.Error("ping redis err %v", err)
		return
	}

	return
}

func initBlackRedis() (err error) {
	secKillConf.blackRedisPool = &redis.Pool{
		MaxIdle:     secKillConf.RedisBlackConf.RedisMaxIdle,
		MaxActive:   secKillConf.RedisBlackConf.RedisMaxActive,
		IdleTimeout: time.Duration(secKillConf.RedisBlackConf.RedisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConf.RedisBlackConf.RedisAddr)
		},
	}
	conn := secKillConf.blackRedisPool.Get()
	defer conn.Close()

	_, err = conn.Do("ping")
	if err != nil {
		logs.Error("ping redis err %v", err)
		return
	}

	return
}

//同步Ip黑名单
func SyncIpBlackList() {
	var ipList []string
	lastTime := time.Now().Unix()
	for {
		conn := secKillConf.blackRedisPool.Get()
		fmt.Println("syncIpBlacklist---------blpop")
		defer conn.Close()
		reply, err := conn.Do("BLPOP", "blackiplist", 0)
		ip, err := redis.String(reply, err)
		if err != nil {
			continue
		}
		curTime := time.Now().Unix()
		ipList = append(ipList, ip)
		if len(ipList) > 100 || curTime-lastTime > 5 {
			secKillConf.RwBlackLock.Lock()
			for _, v := range ipList {
				secKillConf.IpBlackMap[v] = true
			}
			secKillConf.RwBlackLock.Unlock()
			lastTime = curTime
			logs.Info("sync ip list from redia succ,ip[%v]", ipList)
		}

	}
}

//同步Id黑名单
func SyncIdBlackList() {
	for {
		conn := secKillConf.blackRedisPool.Get()
		fmt.Println("syncIdBlacklist---------blpop")
		defer conn.Close()
		reply, err := conn.Do("BLPOP", "blackidlist", 0)
		id, err := redis.Int(reply, err)
		if err != nil {
			continue
		}
		secKillConf.RwBlackLock.Lock()
		secKillConf.IdBlackMap[id] = true
		secKillConf.RwBlackLock.Unlock()
		logs.Info("sync id list from redis succ, ip[%v]", id)
	}
}

//加载黑名单到Map

func loadBlackList() (err error) {
	secKillConf.IpBlackMap = make(map[string]bool, 10000)
	secKillConf.IdBlackMap = make(map[int]bool, 10000)
	err = initBlackRedis()
	if err != nil {
		logs.Error("init black redis failed,err :%v", err)
		return
	}
	conn := secKillConf.blackRedisPool.Get()
	defer conn.Close()

	reply, err := conn.Do("hgetall", "idblacklist")
	idlist, err := redis.Strings(reply, err)
	if err != nil {
		logs.Warn("hget all failed,err:%v", err)
		return
	}

	for _, v := range idlist {
		id, err := strconv.Atoi(v)
		if err != nil {
			logs.Warn("invalid user id[%d]", id)
			continue
		}
		secKillConf.IdBlackMap[id] = true
	}

	reply, err = conn.Do("hgetall", "ipblacklist")
	iplist, err := redis.Strings(reply, err)
	if err != nil {
		logs.Error("hgetall redis err %v", err)
		return
	}
	for _, v := range iplist {

		secKillConf.IpBlackMap[v] = true
	}
	go SyncIpBlackList()
	go SyncIdBlackList()
	return
}

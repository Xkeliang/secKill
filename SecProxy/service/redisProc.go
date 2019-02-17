package service

import (
	"encoding/json"
	"fmt"
	"github.com/beego/logs"
	"github.com/garyburd/redigo/redis"
	"time"
)

func WriteHandle() {
	for {
		//阻塞接受请求
		//写入redis
		fmt.Println("one writeHandle start")
		req := <-secKillConf.SecReqChan
		conn := secKillConf.proxy2LayerRedisPool.Get()
		data, err := json.Marshal(req)
		if err == redis.ErrNil {
			time.Sleep(1 * time.Second)
			conn.Close()
			continue
		}
		if err != nil {
			logs.Error("json Marshal failed,err :%v,req :%v", err, req)
			conn.Close()
			continue
		}
		logs.Debug("LPUSH sec_queue start:%v",data)
		_, err = conn.Do("LPUSH", "sec_queue", data)
		logs.Debug("LPUSH sec_queue success:%v",data)
		if err != nil {
			logs.Error("lpush failed,err :%v,req :%v", err, req)
			conn.Close()
			continue
		}
		conn.Close()
		fmt.Println("one writeHandle end")
	}
	return
}

func ReadHandle() {
	for {
		conn := secKillConf.layer2ProxyRedisPool.Get()
		fmt.Println("one readHandle start")
		//从redis不停获取响应
		reply, err := conn.Do("BRPOP", "recv_queue",0)
		if err != nil {
			logs.Error("rpop failed, err:%v", err)
			conn.Close()
			continue
		}
		tmp, ok := reply.([]interface{})
		if !ok || len(tmp) != 2{
			logs.Error("pop from queue failed, err:%v", err)
			conn.Close()
			continue
		}

		data, ok := tmp[1].([]byte)
		if !ok {
			logs.Error("pop from queue failed, err:%v", err)
			conn.Close()
			continue
		}

		logs.Debug("pop from queue, data:%s", string(data))

		var result SecResult
		err = json.Unmarshal([]byte(data), &result)
		if err != nil {
			logs.Error("json.Unmarshal failed, err:%v", err)
			conn.Close()
			continue
		}
		logs.Debug("recv_queue data :%v",result)
		//检测Map中（key，value）UserConMap
		userKey := fmt.Sprintf("%v_%v", result.UserId, result.ProductId)

		secKillConf.UserConnMapLock.Lock()
		resultChan, ok := secKillConf.UserConMap[userKey]
		secKillConf.UserConnMapLock.Unlock()
		if !ok {
			logs.Warn("user not found:%v", userKey)
			conn.Close()
			continue
		}
		//将响应存入通道
		resultChan <- &result
		fmt.Println("one readHandle end")
		conn.Close()
	}
	return
}

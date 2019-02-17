package service

import (
	"fmt"
	"crypto/md5"
	"encoding/json"
	"time"
	"github.com/beego/logs"
	"math/rand"
)

func Run(conf *SecLayerConf) (err error) {
	//初始化处理线程
	err = RunProcess(conf)
	return
}
func RunProcess(conf *SecLayerConf) (err error) {
	for i := 0; i < secLayerContext.secLayerConf.ReadGoroutineNum; i++ {
		secLayerContext.waitGroup.Add(1)
		go HandleReader(conf)
	}
	for i := 0; i < secLayerContext.secLayerConf.WriteGoroutineNum; i++ {
		secLayerContext.waitGroup.Add(1)
		go HandleWriter()
	}
	for i := 0; i < secLayerContext.secLayerConf.HandleUserGoroutineNum; i++ {
		secLayerContext.waitGroup.Add(1)
		go HandleUser()
	}
	logs.Debug("all prcess goroutine started")
	secLayerContext.waitGroup.Wait()
	logs.Debug("wait all goroutine exited")
	return
}

func HandleReader(conf *SecLayerConf) {
	for {
		conn := secLayerContext.Proxy2LayerRedisPool.Get()
		for {
			ret,err :=conn.Do("brpop", secLayerContext.secLayerConf.Proxy2LayerRadis.RedisQueueName,0)//阻塞时间，0无限阻塞
			if err != nil {
				logs.Error("pop from queue failed err:%v", err)
				break

			}

			tmp, ok := ret.([]interface{})
			if !ok || len(tmp) != 2{
				logs.Error("pop from queue failed, err:%v", err)
				continue
			}

			data, ok := tmp[1].([]byte)
			if !ok {
				logs.Error("pop from queue failed, err:%v", err)
				continue
			}

			logs.Debug("pop from queue, data:%s", string(data))

			var req SecRequest
			err = json.Unmarshal([]byte(data), &req)
			if err != nil {
				logs.Error("unmarshal to secrequeset failed,err", err)
				continue
			}
			now := time.Now().Unix()
			if now-req.AccessTime.Unix() >= int64(conf.MaxrequestWaitTimeout) {
				logs.Warn("req[%v] is expire", req)
				continue
			}
			timer := time.NewTicker(time.Millisecond * time.Duration(secLayerContext.secLayerConf.SendToHandleChanTimeout))
			select {
			case secLayerContext.Read2HandleChan <- &req:
			case <-timer.C:
				logs.Warn("send to handle chan timeout,req:%v", req)
				break
			}
			//secLayerContext.Read2HandleChan <- &req
		}
		conn.Close()
	}
}

func HandleWriter() {
	logs.Debug("handle write running")

	for res := range secLayerContext.Handle2writeChan {
		err := sendToRedis(res)
		if err != nil {
			logs.Error("send to redis,err:%v,res:%v", err, res)
			continue
		}
	}
}

func sendToRedis(res *SecResponse) (err error) {
	data, err := json.Marshal(res)
	if err != nil {
		logs.Error("marshal failed,err", err)
		return
	}
	conn := secLayerContext.Layer2ProxyRedisPool.Get()
	defer conn.Close()
	_, err = conn.Do("rpush", secLayerContext.secLayerConf.Layer2ProxyRadis.RedisQueueName, string(data))
	if err != nil {
		logs.Error("rpush to redis failed,res:%v,err:%v", res, err)
		return
	}
	return
}

func HandleUser() {
	logs.Debug("handle user running")
	for req := range secLayerContext.Read2HandleChan {
		logs.Debug("begin process request:%v", req)
		res, err := HandleSeckill(req)
		if err != nil {
			logs.Warn("process request %v failed,err:%v", res, err)
			res = &SecResponse{
				Code: ErrServiceBusy,
			}
		}
		timer := time.NewTicker(time.Millisecond * time.Duration(secLayerContext.secLayerConf.SendToWriteChanTimeout))
		select {
		case secLayerContext.Handle2writeChan <- res:
		case <-timer.C:
			logs.Warn("send to response chan timeout,res:%v", res)
			break

		}
		//secLayerContext.Handle2writeChan <- res
	}
	return
}

func HandleSeckill(req *SecRequest) (res *SecResponse, err error) {

	secLayerContext.RWSecProductLock.RLock()
	defer secLayerContext.RWSecProductLock.RUnlock()

	res = &SecResponse{}
	res.UserId=req.UserId
	res.ProductId=req.ProductId
	product, ok := secLayerContext.secLayerConf.SecProductInfoMap[req.ProductId]
	if !ok {
		logs.Error("not found product:%v", req.ProductId)
		res.Code = ErrNotFoundProduct
		return
	}
	if product.Status == ProductStatusSoldout {
		res.Code = ErrSoldout
		return
	}
	now := time.Now().Unix()
	alreadySoldCount := product.secLimit.Check(now)
	if alreadySoldCount >= product.SoldMaxlimit {
		res.Code = ErrRetry
		return
	}
	secLayerContext.HistoryMapLock.Lock()
	userHistory, ok := secLayerContext.HistoryMap[req.UserId]

	if !ok {
		userHistory = &UserBuyHistory{
			history: make(map[int]int, 16),
		}
		secLayerContext.HistoryMap[req.UserId] = userHistory
	}

	historyCount := userHistory.GetProductBuyCount(req.ProductId)
	secLayerContext.HistoryMapLock.Unlock()

	if historyCount >= product.OnePersonBuyLimit {
		res.Code = ErrAlreadyBuy
		return
	}

	curSoldCount := secLayerContext.productCountMgr.Count(req.ProductId)
	if curSoldCount >= product.Total {
		res.Code = ErrSoldout
		return
	}
	curRate := rand.Float64()
	if curRate >= product.BuyRate {
		res.Code = ErrRetry
		return
	}
	userHistory.Add(req.ProductId, 1)
	secLayerContext.productCountMgr.Add(req.ProductId, 1)

	//用户Id商品id当前时间密钥
	tokenData := fmt.Sprintf("userId=%d&productId=%d&timestamp=%d&security=%s", req.UserId, req.ProductId, now, secLayerContext.secLayerConf.TokenPasswd)
	res.Token = fmt.Sprintf("%x", md5.Sum([]byte(tokenData)))
	res.TokenTime = now

	res.Code = ErrSecKillSucc

	return
}

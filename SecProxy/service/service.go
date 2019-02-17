package service

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/beego/logs"
	"time"
)

func NewSecRequest() (secRequest *SecRequest) {
	secRequest = &SecRequest{
		ResultChan: make(chan *SecResult, 1),
	}
	return
}

//从map中获取产品信息格式化到data类型
func SecInfo(productId int) (data []map[string]interface{}, code int, err error) {
	secKillConf.RwSecKillConfLock.RLock()
	defer secKillConf.RwSecKillConfLock.RUnlock()

	item, code, err := SecInfoById(productId)

	if err != nil {
		logs.Debug("SecInfoById,err", err)
		return
	}
	data = append(data, item)
	return
}

//获取活动产品
func SecInfoList() (data []map[string]interface{}, code int, err error) {
	secKillConf.RwSecKillConfLock.RLock()
	defer secKillConf.RwSecKillConfLock.RUnlock()
	for _, v := range secKillConf.SecProductInfoMap {
		item, _, err := SecInfoById(v.ProductId)
		if err != nil {
			logs.Error("get product_id[%d] failed err:%v ", v.ProductId, err)
			continue
		}
		logs.Debug("get product[%d]， result[%v], all[%v] v[%v]", v.ProductId, item, secKillConf.SecProductInfoMap, v)
		data = append(data, item)
	}
	return
}

//获取活动产品通过产品Id
func SecInfoById(productId int) (data map[string]interface{}, code int, err error) {
	secKillConf.RwSecKillConfLock.RLock()
	defer secKillConf.RwSecKillConfLock.RUnlock()
	logs.Debug("SecPofductInfoMap:%v", secKillConf.SecProductInfoMap)
	v, ok := secKillConf.SecProductInfoMap[productId]
	logs.Info("value[%d]=%v", productId, v)
	if !ok {
		code = ErrNotFoundProductId
		err = fmt.Errorf("not fount productId", err)
		return
	}
	start := false
	end := false
	status := "success"
	now := time.Now().Unix()
	if now-v.StartTime < 0 {
		start = false
		end = false
		status = "sec kil do not begin"
		code = ErrActiveNotStart
	}
	if now-v.StartTime > 0 {
		start = true
	}
	if now-v.EndTime > 0 {
		start = false
		end = true
		status = "sec kill is already end"
		code = ErrActiveAlreadyEnd
	}
	if v.Status == ProductStatusForceSaleOut || v.Status == ProductStatusSaleOut {
		start = false
		end = true
		status = "Product is sale out"
		code = ErrActiveSaleOut
	}
	data = make(map[string]interface{})
	data["product_id"] = productId
	data["start"] = start
	data["end"] = end
	data["statue"] = status
	return
}


//处理抢购请求
//获取请求进行过滤
//格式化请求传入radis
//从redis获取响应
func SecKill(req *SecRequest) (data map[string]interface{}, code int, err error) {
	secKillConf.RwSecKillConfLock.RLock()
	defer secKillConf.RwSecKillConfLock.RUnlock()

	//过滤用户
/*	err = userCheck(req)
	if err != nil {
		code = ErrUserCheckAuthFailed
		logs.Warn("userId[%d] invalid,check failed,req[%v]", req.UserId, req)
		return
	}*/

	//反作弊
	//即，检查用户、ip请求速度和数量
	/*err = antiSpam(req)
	if err != nil {
		code = ErrUserServiceBusy
		logs.Warn("userId[%d] invalid,check failed,req[%v]", req.UserId, req)
		return
	}*/

	//查询请求的产品状态
	data, code, err = SecInfoById(req.ProductId)
	if err != nil {
		logs.Warn("userId[%d] secInfoById failed,erq[%v]", req.UserId, req)
		return
	}

	if code != 0 {
		logs.Warn("userId[%d] secInfoById failed,code[%d] req[%v]", req.UserId, code, req)
		return
	}

	//建立用户请求Map（userid与productid）
	userKey := fmt.Sprintf("%v_%v", req.UserId, req.ProductId)
	secKillConf.UserConMap[userKey] = req.ResultChan

	//把请求放入通道加入redis
	secKillConf.SecReqChan <- req

	ticker := time.NewTicker(time.Second * 10)

	defer func() {
		ticker.Stop()
		secKillConf.UserConnMapLock.Lock()
		delete(secKillConf.UserConMap, userKey)
		secKillConf.UserConnMapLock.Unlock()
	}()

	//select
	//超时关闭
	//请求关闭
	//获取响应
	select {
	case <-ticker.C:
		code = ErrProcessTimeout
		err = fmt.Errorf("request timeout")
		return
	case <-req.CloseNotify:
		code = ErrClientClosed
		err = fmt.Errorf("client already closed")
		return
	case result := <-req.ResultChan:
		code = result.Code
		data["product_id"] = result.ProductId
		data["token"] = result.Token
		data["user_id"] = result.UserId

		return
	}
	return
}


//检查用户refe
//检查用户合法性
//数字签名加密验证id与cookie  md5
//与userCookieSign比较
func userCheck(req *SecRequest) (err error) {
	found := false
	for _, refer := range secKillConf.ReferWhiteList {
		if refer == req.ClientRefence {
			found = true
			break
		}
	}
	if !found {
		err = fmt.Errorf("invalid request")
		logs.Warn("user[%d],check failed,req[%v]", req.UserId, req)
	}
	authData := fmt.Sprintf("%d%s", req.UserId, secKillConf.CookieSecretKey)
	autoSign := fmt.Sprintf("%s", md5.Sum([]byte(authData)))
	if autoSign != req.UserCookieSign {
		err = errors.New("invalid user cookie")
		return
	}

	return
}

package model

import (
	"github.com/beego/logs"
	"time"
	"fmt"
	"encoding/json"
	"strings"
	"context"
)

const (
	ActivityStatusNormal = 0
	ActivityStatusDisable =1
	ActivityStatusExpire =2
)

//活动产品
type Activity struct {
	ActivityId int `db:"id"`
	ActivityName string `db:"name"`
	ProductId int `db:"product_id"`
	StartTime int64 `db:"start_time"`
	EndTime int64 `db:"end_time"`
	Total int `db:"total"`
	Status int `db:"status"`

	StartTimeStr string
	EndTimeStr string
	StatusStr string

	Speed int `db:"sec_speed"`
	BuyLimit int `db:"buy_limit"`
	BuyRate float64 `db:buy_rate`
}
//秒杀信息加载到Etcd
type SecProductInfoConf struct {
	ProductId         int   `json:"product_id"`
	StartTime         int64 `json:"start_time"`
	EndTime           int64 `json:"end_time"`
	Status            int   `json:"status"`
	Total             int   `json:"total"`
	Left              int   `json:"left"`
	OnePersonBuyLimit int	`json:"one_person_buy_limit"`
	BuyRate           float64 `json:"buy_rate"`

	//每秒最多能卖多少个
	SoldMaxlimit int	`json:"sold_maxlimit"`
}
type ActivityModel struct {
}

func NewActivityModel() *ActivityModel {
	return &ActivityModel{}
}

func (p *ActivityModel)GetActivityList()(activitList []*Activity,err error)  {
	sql := "select id,name,product_id,start_time,end_time,total,status,buy_limit,sec_speed from activity order by id desc"
	err = Db.Select(&activitList,sql)
	if err != nil {
		logs.Error("select activity from database failed,err:%v",err)
		return
	}
	for _,v := range activitList{
		t := time.Unix(v.StartTime,0)
		v.StartTimeStr=t.Format("2006-01-02 15:04:05")
		t = time.Unix(v.EndTime,0)
		v.EndTimeStr=t.Format("2006-01-02 15:04:05")

		now := time.Now().Unix()
		if now > v.EndTime {
			v.StatusStr= "已结束"
			continue
		}
		if v.Status == ActivityStatusNormal {
			v.StatusStr="正常"
		}else if v.Status == ActivityStatusDisable {
			v.StatusStr="已禁用"
		}
	}
	return
}

func (p *ActivityModel)ProductValid(productId int,total int)(valid bool,err error)  {
	sql := "select id,name,total,status from product where id=?"
	var productList  []*Product
	err = Db.Select(&productList,sql,productId)
	if err != nil {
		logs.Warn("select product failed,err:%v",err)
		return
	}

	if len(productList) == 0  {
		logs.Warn("select product is not exist")
		valid = false
		return
	}
	if  total > productList[0].Total {
		logs.Warn("select Total is more Sum")
		valid = false
		return
	}
	valid = true
	return
}

func (p *ActivityModel) CreateActivity(activity *Activity)(err error)  {
	
	value,err := p.ProductValid(activity.ProductId,activity.Total)
	if err != nil {
		logs.Error("product exists failed,err:%v",err)
		logs.Error(err)
		return
	}
	if !value {
		err = fmt.Errorf("product id[%v] is not exist or total is more sum",activity.ProductId)
		logs.Error(err)
		return
	}

	if activity.StartTime<=0||activity.EndTime<=0 {
		err = fmt.Errorf("invalid start[%v]||end[%v] time",activity.StartTime,activity.EndTime)
		logs.Error(err)
		return
	}
	if activity.EndTime<=activity.StartTime {
		err = fmt.Errorf("start[%v] is  end[%v] time",activity.StartTime,activity.EndTime)
		logs.Error(err)
		return
	}

	now := time.Now().Unix()
	if activity.StartTime<=now||activity.EndTime<=now {
		err = fmt.Errorf("invalid start[%v]||end[%v] time then now",activity.StartTime,activity.EndTime)
		logs.Error(err)
		return
	}

	sql := "insert into activity (name,product_id,start_time,end_time,total,buy_limit,sec_speed)VALUES (?,?,?,?,?,?,?)"
	_,err = Db.Exec(sql,activity.ActivityName,activity.ProductId,activity.StartTime,activity.EndTime,activity.Total,activity.BuyLimit,activity.Speed)
	if err != nil {
		logs.Warn("insert from mysql failed,err:%v,sql:%v",err,sql)
		return
	}
	logs.Debug("insert into mysql success")

	err = p.SyncToEtcd(activity)
	if err != nil {
		logs.Error("sync to etcd failed,err:",err)
		return
	}
	return
}

//加载到Etcd
func (p *ActivityModel)SyncToEtcd(activity *Activity) (err error) {
	if strings.HasSuffix(EtcdPrefix,"/")==false{
		EtcdPrefix =EtcdPrefix + "/"
	}
	etcdKey := fmt.Sprintf("%s%s",EtcdPrefix,EtcdProductKey)
	secProductInfoList,err := loadProductFromEtcd(etcdKey)
	logs.Debug("product etcdKey = %v",etcdKey)
	var secProductInfo SecProductInfoConf
	secProductInfo.ProductId = activity.ProductId
	secProductInfo.StartTime=activity.StartTime
	secProductInfo.EndTime = activity.EndTime
	secProductInfo.Status=activity.Status
	secProductInfo.Total= activity.Total
	secProductInfo.OnePersonBuyLimit = activity.BuyLimit
	secProductInfo.SoldMaxlimit = activity.Speed
	secProductInfo.BuyRate=activity.BuyRate

	secProductInfoList = append(secProductInfoList,secProductInfo)

	data,err := json.Marshal(secProductInfoList)
	if err != nil {
		logs.Error("json marshal failed,err:%v",err)
		return
	}
	_,err =EtcdClient.Put(context.Background(),etcdKey,string(data))
	if err != nil {
		logs.Error("put to etcd failed,err:%v",err)
		return
	}
	logs.Debug("success load activity into etcd,activity:",secProductInfo)
	return
}

//从Etcd获取活动产品
func loadProductFromEtcd(key string) (secProductInfo []SecProductInfoConf,err error) {
	logs.Debug("start get from etcd")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	resp, err := EtcdClient.Get(ctx, key)
	if err != nil {
		//if err == fmt.Errorf("context deadline exceeded") {
		//	initSetProductWatcher(conf)
		//	return
		//}
		logs.Error("get [%s] from etcd key error :%v", key, err)
		return
	}
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
	return
}

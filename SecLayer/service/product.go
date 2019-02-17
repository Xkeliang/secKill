package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/beego/logs"
	etcd_Client "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

func loadProductFromEtcd(conf *SecLayerConf) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	key := conf.EtcdConfig.EtcdProductSecKey
	resp, err := secLayerContext.etcdClient.Get(ctx, key)
	if err != nil {
		if err == fmt.Errorf("context deadline exceeded") {
			initSetProductWatcher(conf)
			return
		}
		logs.Error("get [%s] from etcd key error :%v", key, err)
		return
	}
	var secProductInfo []SecProductInfoConf
	for k, y := range resp.Kvs {
		logs.Debug("key[%d] value[%s]", k, y)
		err = json.Unmarshal(y.Value, &secProductInfo)
		if err != nil {
			logs.Error("Unmarshal failed:%v", err)
			return
		}
		logs.Debug("sec info succ is [%v]", secProductInfo)
	}
	fmt.Println("load secProductInfo from etcd :%v",secProductInfo)
	logs.Debug("load secProductInfo", secProductInfo)
	updateProductInfo(conf, secProductInfo)
	initSetProductWatcher(conf)
	return
}

func updateProductInfo(conf *SecLayerConf, secProductInfo []SecProductInfoConf) {
	var tem map[int]*SecProductInfoConf = make(map[int]*SecProductInfoConf, 1024)
	for _, v := range secProductInfo {
		product := v
		product.secLimit = &SecLimit{}
		fmt.Println(v)
		tem[v.ProductId] = &product
	}
	//fmt.Println("-------------")
	secLayerContext.RWSecProductLock.Lock()
	conf.SecProductInfoMap = tem
	secLayerContext.RWSecProductLock.Unlock()

}

//监控更新
func initSetProductWatcher(conf *SecLayerConf) {
	go watchSecProductKey(conf, conf.EtcdConfig.EtcdProductSecKey)
}
func watchSecProductKey(conf *SecLayerConf, key string) {
	cli, err := etcd_Client.New(etcd_Client.Config{
		Endpoints:   []string{conf.EtcdConfig.EtcdAddr, "localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: time.Duration(conf.EtcdConfig.EtcdTimeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd error : /%v", err)
		return
	}
	logs.Debug("begin watch key:%s", key)
	for {
		rch := cli.Watch(context.Background(), key)
		var secProductInfo []SecProductInfoConf
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
					updateProductInfo(conf, secProductInfo)
				}
			}
		}
	}
}

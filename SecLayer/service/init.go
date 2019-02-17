package service

import (
	"github.com/beego/logs"
	etcd_Client "go.etcd.io/etcd/clientv3"
	"time"
)

func InitSecLayer(conf *SecLayerConf) (err error) {
	err = initRedis(conf)
	if err != nil {
		logs.Error("init redis err:%v", err)
		return

	}
	logs.Debug("init redis sucess")
	err = initEtcd(conf)
	if err != nil {
		logs.Error("init etcd err:%v", err)
		return
	}
	logs.Debug("init etcd sucess")
	return
}

//初始化etcd
func initEtcd(conf *SecLayerConf) (err error) {
	cli, err := etcd_Client.New(etcd_Client.Config{
		Endpoints:   []string{conf.EtcdConfig.EtcdAddr, "localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: time.Duration(conf.EtcdConfig.EtcdTimeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd error : /%v", err)
		return
	}
	secLayerContext.etcdClient = cli

	err = loadProductFromEtcd(conf)
	if err != nil {
		logs.Error("load product from etcd failed ,err :%v", err)
		return
	}

	secLayerContext.secLayerConf = conf
	secLayerContext.Read2HandleChan = make(chan *SecRequest, secLayerContext.secLayerConf.Read2handleChanSize)
	secLayerContext.Handle2writeChan = make(chan *SecResponse, secLayerContext.secLayerConf.Handle2writeChanSize)
	secLayerContext.HistoryMap = make(map[int]*UserBuyHistory, 10000)

	secLayerContext.productCountMgr = NewProductCountMgr()
	logs.Debug("init all succ")
	return

}

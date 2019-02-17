package model

import (
	"github.com/jmoiron/sqlx"
	etcd_Client "go.etcd.io/etcd/clientv3"
)

var (
	Db *sqlx.DB
	EtcdClient *etcd_Client.Client
	EtcdPrefix string
	EtcdProductKey string
)

func InitDB(db *sqlx.DB,etcdClient *etcd_Client.Client,prefix,productKey string)(err error)  {
	Db = db
	EtcdClient =etcdClient
	EtcdPrefix =prefix
	EtcdProductKey = productKey
	return
}


package service

import (
	"github.com/garyburd/redigo/redis"
	etcd_Client "go.etcd.io/etcd/clientv3"
	"sync"
	"time"
)

var (
	secLayerContext = &SecLayerContext{}
)

type SecLayerConf struct {
	Proxy2LayerRadis RedisConf
	Layer2ProxyRadis RedisConf

	EtcdConfig EtcdConf
	LogPath    string
	LogLevel   string

	WriteGoroutineNum int
	ReadGoroutineNum  int

	HandleUserGoroutineNum int
	SecProductInfoMap      map[int]*SecProductInfoConf

	Read2handleChanSize   int
	MaxrequestWaitTimeout int

	Handle2writeChanSize int

	SendToWriteChanTimeout  int
	SendToHandleChanTimeout int
	TokenPasswd             string
}

type RedisConf struct {
	RedisAddr        string
	RedisMaxIdle     int
	RedisMaxActive   int
	RedisIdleTimeout int
	RedisQueueName   string
}

//etcd
type EtcdConf struct {
	EtcdAddr          string
	EtcdTimeout       int
	EtcdSecKeyPrefix  string
	EtcdProductSecKey string
}

type SecLayerContext struct {
	Proxy2LayerRedisPool *redis.Pool
	Layer2ProxyRedisPool *redis.Pool
	etcdClient           *etcd_Client.Client
	secLayerConf         *SecLayerConf
	RWSecProductLock     sync.RWMutex

	waitGroup sync.WaitGroup

	Read2HandleChan  chan *SecRequest
	Handle2writeChan chan *SecResponse

	HistoryMap     map[int]*UserBuyHistory
	HistoryMapLock sync.Mutex

	productCountMgr *ProductCountMgr
}

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
	SoldMaxlimit int  `json:"sold_maxlimit"`
	//限速控制
	secLimit *SecLimit `json:"-"`
}

type SecRequest struct {
	ProductId      int
	Source         string
	AuthCode       string
	SecTime        string
	Nance          string
	UserId         int
	UserCookieSign string
	AccessTime     time.Time
	ClientAddr     string
	ClientRefence  string
}

type SecResponse struct {
	ProductId int
	UserId    int
	Token     string
	TokenTime int64
	Code      int
}

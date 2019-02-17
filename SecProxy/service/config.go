package service

import (
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
)

const (
	ProductStatusNomal = iota
	ProductStatusSaleOut
	ProductStatusForceSaleOut
)

//redis配置
type RedisConf struct {
	RedisAddr        string
	RedisMaxIdle     int
	RedisMaxActive   int
	RedisIdleTimeout int
}

//etcd配置
type EtcdConf struct {
	EtcdAddr          string
	EtcdTimeout       int
	EtcdSecKeyPrefix  string
	EtcdProductSecKey string
}

//用户与Ip抢购限制
type AccessLimitConf struct {
	IPSecAccessLimit   int
	UserSecAccessLimit int
	IPMinAccessLimit   int
	UserMinAccessLimit int
}

//秒杀系统的配置
type SecKillConf struct {
	RedisBlackConf       RedisConf
	RedisProxy2LayerConf RedisConf
	RedisLayer2ProxyConf RedisConf

	EtcdConf          EtcdConf
	LogPath           string
	LogLevel          string
	SecProductInfoMap map[int]*SecProductInfoConf
	RwSecKillConfLock sync.RWMutex
	CookieSecretKey   string

	//UserSecAcessLimit int
	ReferWhiteList []string
	//IpSecAccessLimit int

	RwBlackLock sync.RWMutex
	IpBlackMap  map[string]bool
	IdBlackMap  map[int]bool

	AccessLimitConf      AccessLimitConf
	blackRedisPool       *redis.Pool
	proxy2LayerRedisPool *redis.Pool
	layer2ProxyRedisPool *redis.Pool

	secLimitMgr *SecLimitMgr

	WriteProxy2LayerGoroutineNum int
	ReadLayer2ProxyGoroutineNum  int

	SecReqChan     chan *SecRequest
	SecReqChanSize int

	UserConMap      map[string]chan *SecResult
	UserConnMapLock sync.Mutex
}

//读取Etcd产品信息
type SecProductInfoConf struct {
	ProductId int   `json:"product_id"`
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
	Status    int   `json:"status"`
	Total     int   `json:"total"`
	Left      int   `json:"left"`
	OnePersonBuyLimit int	`json:"one_person_buy_limit"`
	BuyRate           float64 `json:"buy_rate"`

	//每秒最多能卖多少个
	SoldMaxlimit int  `json:"sold_maxlimit"`
	//限速控制
	secLimit *SecLimit `json:"-"`
}

//响应信息
type SecResult struct {
	ProductId int
	UserId    int
	Token     string
	TokenTime int64
	Code      int
}


//请求信息
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

	CloseNotify <-chan bool	 `json:"-"`

	ResultChan  chan *SecResult	 `json:"-"`
}

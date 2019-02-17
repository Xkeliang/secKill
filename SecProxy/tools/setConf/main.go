package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

const (
	etcdKey = "/oldboy/seckill/product"
)

type SecInfoConf struct {
	ProductId int `json:"product_id"`
	StartTime int `json:"start_time"`
	EndTime   int `json:"end_time"`
	Status    int `json:"status"`
	Total     int `json:"total"`
	Left      int `json:"left"`
}

func SetLogConfToEtcd() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2380", "localhost:2379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect failed err:", err)
		return
	}
	fmt.Println("connect success")
	defer cli.Close()

	var SecInfoConfArr []SecInfoConf

	SecInfoConfArr = append(
		SecInfoConfArr, SecInfoConf{
			1027,
			1544194914,
			1544195555,
			0,
			100,
			100,
		},
	)
	SecInfoConfArr = append(
		SecInfoConfArr, SecInfoConf{
			1026,
			1544194914,
			1544195555,
			0,
			100,
			100,
		},
	)

	data, err := json.Marshal(SecInfoConfArr)
	if err != nil {
		fmt.Println("marshal failed", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	fmt.Println(ctx)
	_, err = cli.Put(ctx, etcdKey, string(data))

	if err != nil {
		fmt.Println("put failed", err)
		return
	}
	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := cli.Get(ctx, etcdKey)
	cancel()
	fmt.Println("resp", resp)
	if err != nil {
		fmt.Println("get failed", err)
		return
	}
	fmt.Println(resp.Kvs)
	for _, ev := range resp.Kvs {
		fmt.Printf("----%s:%s\n", ev.Key, ev.Value)
	}

}
func main() {
	SetLogConfToEtcd()
}

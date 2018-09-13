package main

import (
	"context"
	"fmt"
	"github.com/smartwalle/etcd4go"
	"github.com/smartwalle/grpc4go"
	"github.com/smartwalle/grpc4go/sample/hw"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"time"
)

func main() {
	// 初始化 etcd 连接配置文件
	var config = clientv3.Config{}
	config.Endpoints = []string{"localhost:2379"}

	// 注册命名解析及服务发现
	var c, _ = etcd4go.NewClient(config)
	var r = grpc4go.NewETCDResolver(c)
	resolver.Register(r)

	// dial
	conn, err := grpc.Dial("etcd://mine/my_service", grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	cc := hw.NewFirstGRPCClient(conn)

	for {
		time.Sleep(time.Second * 1)
		rsp, err := cc.FirstCall(context.Background(), &hw.FirstRequest{Name: "Yang"})
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("rand", rsp.Message)
	}
}

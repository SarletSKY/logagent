package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

//默认使用的先加入数据
func main() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})

	if err != nil {
		fmt.Printf("connect client err:%v", err)
		return
	}
	fmt.Println("connect success")
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	//这里使用的value是json格式化。因为考虑到多个日志文件
	//value := `[{"log_path":"/home/zwx/local.log","topic":"web_log","service":"account","send_rate":50000},{"log_path":"/opt/go/xxx/redis.log","topic":"redis_log","service":"account","send_rate":50000}]`
	value := `[{"log_path":"/home/zwx/local.log","topic":"web_log","service":"account","send_rate":50000},{"log_path":"/opt/go/xxx/redis.log","topic":"redis_log","service":"account","send_rate":50000},{"log_path":"/home/zwx/mysql.log","topic":"mysql_log","service":"account","send_rate":50000}]`
	_, err = client.Put(ctx, "/192.168.126.129/xxx", value)
	cancel()
	if err != nil {
		fmt.Printf("put to etcd failed,err:%v\n", err)
		return
	}
}
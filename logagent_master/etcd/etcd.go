package etcd

import (
	"time"
	"github.com/coreos/etcd/clientv3"
	"fmt"
	"github.com/astaxie/beego/logs"
	"sync"
	"context"
)

//全局的etcd客户端
var etcdClient *clientv3.Client
//用于之后新增/删除用。
var logConf chan string
var Wg sync.WaitGroup

//初始化etcd
func InitEtcd(address string,keyfmt string, timeout time.Duration,ips []string) (err error){
	//初始化数据[etcd的value]
	logConf = make(chan string, 100)

	//因为有多个ip的key值存进---->多个不同ip的key
	var keys []string
	//对keys值完整补全---->因为keyfmt还要根据ip来区分
	for _,ip := range ips{
		keys = append(keys,fmt.Sprintf(keyfmt,ip))
	}
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:[]string{address},
		DialTimeout:timeout,
	})
	if err != nil {
		logs.Error("init etcd config failed:%v",err)
		return
	}
	logs.Debug("init etcd config success")

	Wg.Add(1)

	// 对logConf赋值
	for _, key := range keys{
		//根据不同的key的value存到logConf
		ctx, cancel := context.WithTimeout(context.Background(),time.Second)
		//对应某个key值的value获取[Get]值
		response, err := etcdClient.Get(ctx,key)
		cancel()
		if err != nil {
			logs.Error("etcd get value failed:%v",err)
			return err
		}
		logs.Debug("etcd get value success")
		for _,kv :=  range response.Kvs{
			logs.Debug("%v---%v",kv.Key, kv.Value)
			//将数据存进logConf
			logConf <- string(kv.Value)
		}
	}
	//watch
	go etcdWatch(keys)
	return
}


func etcdWatch(keys []string) {
	//存入watch的通道
	var watchChan []clientv3.WatchChan
	for _, key := range keys{
		WatchC := etcdClient.Watch(context.Background(),key)
		//让key加入watchChan
		watchChan = append(watchChan,WatchC)
	}
	for{
		for _,watchC := range watchChan{
			select {
			case watchResponse := <-watchC:
				for _, kv := range watchResponse.Events{
					logs.Debug("watch: type:%v---%v---%v",kv.Type,kv.Kv.Key,kv.Kv.Value)
					logConf <- string(kv.Kv.Value)
				}
			default:

			}
		}
		time.Sleep(time.Second)
	}
	Wg.Done()
}

//向外提供接口获取logConf
func GetLogConf() <- chan string{
	return logConf
}

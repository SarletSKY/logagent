package main

import (
	"gopkg.in/ini.v1"
	"再次练习/logagent_master/conf"
	"github.com/astaxie/beego/logs"
	"encoding/json"
	"fmt"
	"再次练习/logagent_master/kafka"
	"再次练习/logagent_master/utils"
	"time"
	"再次练习/logagent_master/etcd"
	"再次练习/logagent_master/taillog"
)

var cfg = new(conf.AppConf)

//初始化日志
func initLog() (err error){
	//设置一个config--->map
	config := make(map[string]interface{})
	config["filename"] = cfg.LogConf.LogPath
	config["level"] = getLevel(cfg.LogConf.LogLevel)
	//进行序列化
	configBytes, err := json.Marshal(config)
	if err != nil {
		fmt.Printf("marshal err:%v",err)
		return
	}
	//给日志设置配置
	logs.SetLogger(logs.AdapterFile,string(configBytes))
	return
}

//获取日志的等级
func getLevel(level string) int {
	switch level {
	case "info":
		return logs.LevelInfo
	case "debug":
		return logs.LevelDebug
	case "error":
		return logs.LevelError
	case "warn":
		return logs.LevelWarn
	case "alert":
		return logs.LevelAlert
	default:
		return logs.LevelDebug
	}
}
func main(){
	//初始化配置
	err := ini.MapTo(cfg,"./logagent_master/conf/config.ini")
	if err != nil {
		panic(err)
	}
	logs.Info("init ini success")
	//初始化日志
	err = initLog()
	if err != nil {
		logs.Error("init log failed:%v",err)
		return
	}
	logs.Info("init log success")
	//设置ip
	ips, err := utils.GetLocalIP()
	if err != nil {
		logs.Error("get ip local failed, err:%v", err)
		return
	}
	for _,val := range ips{
		logs.Info(val)
	}
	logs.Info("get ip local success")
	//初始化kafka
	err = kafka.InitKafka(cfg.KafkaConf.Address,cfg.KafkaConf.MaxSie,cfg.KafkaConf.ThreadCount)
	if err != nil {
		logs.Error("init kafka failed:%v",err)
		return
	}
	logs.Info("init kafka success")
	//初始化etcd
	err = etcd.InitEtcd(cfg.EtcdConf.Address,cfg.EtcdConf.Key,time.Duration(cfg.EtcdConf.Timeout) * time.Millisecond,ips)
	if err != nil{
		logs.Error("init etcd failed, err:%v",err)
		return
	}
	logs.Info("init etcd success")
	//初始化taillog
	//出吃化taillog管理者
	taiilMgr := taillog.NewTailMgr()
	taiilMgr.Process()
	//运行
	etcd.Wg.Wait()
}

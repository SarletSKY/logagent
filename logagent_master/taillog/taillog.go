package taillog

import (
	"github.com/hpcloud/tail"
	"sync"
	"再次练习/logagent_master/etcd"
	"encoding/json"
	"github.com/astaxie/beego/logs"
	"再次练习/logagent_master/kafka"
	"fmt"
)

type TailMgr struct {
	tailObjMap map[string]*TailObj //根据不通过key构造不同的tail
	lock sync.Mutex
}

//处理过程
func (mgr *TailMgr) Process() {
	//获取etcd的logConf
	logChan := etcd.GetLogConf()
	//对logConf与data的对象数据进行对应，并反序列化
	for conf := range logChan{
		logs.Debug("log conf :%v", conf)
		var dataLogConf []logConfig // data的4个字段
		err := json.Unmarshal([]byte(conf),&dataLogConf)
		if err != nil {
			logs.Error("unmarshal logConf failed, err:%v",err)
			return
		}
		logs.Debug("unmarshal logConf success:%v",dataLogConf)
		err = mgr.reloadConfig(dataLogConf)
		if err != nil {
			logs.Error("reload config from etcd failed,err:%v",err)
			return
		}
		logs.Debug("reload config from etcd success:%v",dataLogConf)
	}
}

//处理tailObj每个日志文本，增加/删除
func (mgr *TailMgr) reloadConfig(dataLogConf []logConfig) (err error){
	//对TailMgr  tailObjMap进行赋值
	for _, conf := range dataLogConf{
		tailObj, ok := mgr.tailObjMap[conf.LogPath]
		//如果获取不到值，说明还没有初始化tailObj
		if !ok {
			logs.Debug("conf:%v -- tailObj:%v",conf,tailObj)
			//进行tailObj初始化
			err = mgr.AddLogFile(conf)
			if err != nil {
				logs.Error("add log file failed, err:%v",err)
				continue
			}
			continue
		}
		//如果是初始化过了
		//将管理者的map值设置好/value设置好
		tailObj.logConf = conf
		mgr.tailObjMap[conf.LogPath] = tailObj
		logs.Info(mgr.tailObjMap)
	}
	//删除处理  map就是为了比较数据而生。
	for key,tailObj := range mgr.tailObjMap{
		var find = false
		for _,newValue := range dataLogConf{
			if key == newValue.LogPath {
				//相同路径[key]进行map下一个比较
				find = true
				break
			}
		}
		//当一轮下去找不到，就将exitchan进程退出掉
		if find == false{
			logs.Warn("log path :%s is remove",key)
			tailObj.exitChan <- true
			delete(mgr.tailObjMap,key)
		}
	}
	return
}

//进行单个tailObj初始化
func (mgr *TailMgr) AddLogFile(conf logConfig) (err error) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	_, ok := mgr.tailObjMap[conf.LogPath]
	if ok {
		err = fmt.Errorf("duplicate filename:%s\n", conf.LogPath)
		return
	}

	// 初始化taillog
	tailFile,err := tail.TailFile(conf.LogPath,tail.Config{
		ReOpen:true,
		Follow:true,
		Location:&tail.SeekInfo{Offset:0,Whence:2},
		MustExist:false,
		Poll:true,
	})

	if err != nil{
		logs.Error("init taillog failed,err:%v",err)
		return
	}
	logs.Info("init taillog success")
	//初始化tailObj
	tailObj := &TailObj{
		tail: tailFile,
		offset:0,
		secLimit:NewSecondLimit(int32(conf.SendRate)),
		logConf:conf,
		exitChan:make(chan bool,1),
	}

	mgr.tailObjMap[conf.LogPath] = tailObj
	logs.Info("new create taillog map:[%s]",mgr.tailObjMap)
	go tailObj.readLog()
	return
}

//tail的对象
type TailObj struct {
	tail   *tail.Tail
	offset int64
	secLimit *SecondLimit
	logConf logConfig	//value对象的数据
	exitChan chan bool
}

//某日志单个tailObj去读取
func (tailObj *TailObj) readLog() {
	//读取每行的日志内容
	for line := range tailObj.tail.Lines{
		if line.Err != nil{
			logs.Error("read taillog line failed, err:%v",line.Err)
			continue
		}
		//通过taillog给kafka发送数据
		kafka.AddMessage(line.Text,tailObj.logConf.Topic)
		tailObj.secLimit.Add(1)
		tailObj.secLimit.Wait()

		select {
		case <-tailObj.exitChan:
			logs.Warn("tailobj is exited,config:%v",tailObj.logConf)
			return
		default:

		}
	}
	etcd.Wg.Done()
}

//初始化管理者对象
func NewTailMgr() *TailMgr{
	tailMgr := &TailMgr{
		tailObjMap:make(map[string]*TailObj),
	}
	return tailMgr
}


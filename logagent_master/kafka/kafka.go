package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/astaxie/beego/logs"
)

//全局的客户端
var client sarama.SyncProducer
var kafkaSenderGlobal *KafkaSender

//kafka数据
type Message struct {
	topic string //分区
	line string //发送信息
}

//kafka的配置
type KafkaSender struct {
	client sarama.SyncProducer
	lineConfChan chan *Message
}

func (sender *KafkaSender) sendToKafKa() {
	for val := range sender.lineConfChan{
		//sarama信息对象
		msg := &sarama.ProducerMessage{}
		msg.Topic = val.topic
		msg.Value = sarama.StringEncoder(val.line)
		partition, offset, err := sender.client.SendMessage(msg)
		if err != nil {
			logs.Error("kafka send data failed, err:%v",err)
			return
		}
		logs.Debug("partition:%v,offset:%v",partition,offset)
	}
}

//初始化kafka
func InitKafka(address string,maxsize int, threadCount int) (err error){
	//初始化kafka发送数据的对象
	kafka := &KafkaSender{
		lineConfChan:make(chan *Message,maxsize),
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	client, err = sarama.NewSyncProducer([]string{address},config)
	if err != nil {
		logs.Error("init kafka client failed, err:%v",err)
		return
	}
	logs.Debug("init kafka client success")

	kafka.client = client
	kafkaSenderGlobal = kafka
	//启动多线程来启动往kafka发送数据
	for i:=0; i<threadCount; i++{
		go kafka.sendToKafKa()
	}
	return
}

//通过taillog给kafka发送值
func AddMessage(line string, topic string) (err error) {
	//我们通过tailf读取的日志文件内容先放到channel里面
	kafkaSenderGlobal.lineConfChan <- &Message{line: line, topic: topic}
	return
}



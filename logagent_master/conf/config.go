package conf

type AppConf struct {
	KafkaConf `ini:"kafka"`
	EtcdConf  `ini:"etcd"`
	LogConf   `ini:"log"`
}

type KafkaConf struct {
	Address     string `ini:"address"`
	ThreadCount int    `ini:"kafka_thread_count"`
	MaxSie      int    `ini:"chan_max_size"`
}

type EtcdConf struct {
	Address string `ini:"address"`
	Timeout int    `ini:"timeout"`
	Key     string `ini:"etcd_watch_key"`
}

type LogConf struct {
	LogPath  string `ini:"log_path"`
	LogLevel string `ini:"log_level"`
}

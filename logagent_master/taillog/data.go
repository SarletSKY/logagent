package taillog

type logConfig struct {
	Topic    string `json:"topic"`     //topic
	LogPath  string `json:"log_path"`  //log_path
	Service  string `json:"service"`   //service
	SendRate int    `json:"send_rate"` //send_rate
}

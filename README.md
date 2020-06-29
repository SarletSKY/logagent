# logagent
taillog+kafka+etcd+beego的logagent收集日志
请自行用go mod拉取依赖
##注意：go.mod文件的grpc一定要1.26.0，如果是1.27会报错，将它修改成以下代码：
google.golang.org/grpc v1.26.0 // indirect

module github.com/itcmdb/cmdb-service

go 1.23

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/itcmdb/shared v0.0.0
	github.com/spf13/viper v1.18.2
)

require (
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/itcmdb/shared => ../shared

module github.com/itcmdb/cmdb-service

go 1.23

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/itcmdb/shared v0.0.0
	github.com/spf13/viper v1.18.2
)

require (
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240528184218-531527333157 // indirect
	google.golang.org/grpc v1.65.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/itcmdb/shared => ../shared

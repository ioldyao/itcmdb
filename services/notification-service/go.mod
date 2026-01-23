module github.com/itcmdb/notification-service

go 1.21

require (
	github.com/itcmdb/shared v0.0.0
	github.com/gin-gonic/gin v1.10.0
	github.com/spf13/viper v1.18.2
)

replace github.com/itcmdb/shared => ../shared

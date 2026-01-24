module github.com/itcmdb/audit-service

go 1.23

require (
	github.com/IBM/sarama v1.43.0
	github.com/gin-gonic/gin v1.10.0
	github.com/itcmdb/shared v0.0.0
	github.com/spf13/viper v1.19.0
	go.uber.org/zap v1.27.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.7
)

replace github.com/itcmdb/shared => ../shared

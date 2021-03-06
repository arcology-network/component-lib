module github.com/arcology-network/component-lib

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/arcology-network/3rd-party v0.9.2-0.20210626004852-924da2642860
	github.com/arcology-network/common-lib v0.9.2-0.20210907015240-00e4f072f9b8
	github.com/arcology-network/concurrenturl v0.0.0-20210908071701-a1f430c90b99
	github.com/Shopify/sarama v1.24.1
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/prometheus/client_golang v1.1.0
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/smallnest/rpcx v0.0.0-20200516063136-b01b68f58652
	github.com/spf13/viper v1.7.1
	go.uber.org/zap v1.15.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

//replace github.com/arcology-network/common-lib => ../common-lib/

//replace github.com/arcology-network/concurrenturl => ../concurrenturl/

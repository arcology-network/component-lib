package lib

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

var asyncProducerCreator func([]string, *sarama.Config) (sarama.AsyncProducer, error)
var consumerCreator func([]string, string, []string, *cluster.Config) (Consumer, error)

func init() {
	asyncProducerCreator = func(brokers []string, config *sarama.Config) (sarama.AsyncProducer, error) {
		return sarama.NewAsyncProducer(brokers, config)
	}
	consumerCreator = func(brokers []string, groupID string, topics []string, config *cluster.Config) (Consumer, error) {
		return cluster.NewConsumer(brokers, groupID, topics, config)
	}
}

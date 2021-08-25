package lib

import (
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type mockConsumer struct {
	msgChan   chan *sarama.ConsumerMessage
	notifChan chan *cluster.Notification
	errChan   chan error
}

func createMockConsumer(brokers []string, groupID string, topics []string, config *cluster.Config) Consumer {
	return &mockConsumer{
		msgChan:   make(chan *sarama.ConsumerMessage, 100),
		notifChan: make(chan *cluster.Notification, 1),
		errChan:   make(chan error, 1),
	}
}

func (consumer *mockConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return consumer.msgChan
}

func (consumer *mockConsumer) Notifications() <-chan *cluster.Notification {
	return consumer.notifChan
}

func (consumer *mockConsumer) Errors() <-chan error {
	return consumer.errChan
}

func (consumer *mockConsumer) MarkOffset(*sarama.ConsumerMessage, string) {

}

func (consumer *mockConsumer) Close() error {
	return nil
}

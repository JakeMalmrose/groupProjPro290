package kafka

import (
	// "fmt"
	"log"
	"os"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	Producer sarama.SyncProducer
}

func (kafka *KafkaProducer) InitKafkaProducer() error {
	url := os.Getenv("KAFKA_BROKER")
	brokersUrl := []string{url}
	err := error(nil)
	kafka.Producer, err = ConnectProducer(brokersUrl)
	if err != nil {
		return err
	}
	return nil
}

func ConnectProducer(brokersUrl []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	// NewSyncProducer creates a new SyncProducer using the given broker addresses and configuration.
	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (kafka *KafkaProducer) PushCommentToQueue(topic string, key string, message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(message),
	}
	log.Println("Sending message to Kafka: ", msg, " with key: ", key, " and message: ", message)
	_, _, err := kafka.Producer.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}

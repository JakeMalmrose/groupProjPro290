package kafka

import (
	// "fmt"
	"os"

	"github.com/IBM/sarama"
)

// func Producer(tpoic string, message []byte) {
//     producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, nil)
//     if err != nil {
//         log.Fatalf("couldnt create a producer: %v", err)
//     }
// }

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

func PushCommentToQueue(topic string, key string, message []byte) error {
	url := os.Getenv("KAFKA_BROKERS")
	brokersUrl := []string{url}
	producer, err := ConnectProducer(brokersUrl)
	if err != nil {
		return err
	}
	// defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.StringEncoder(message),
	}
	_, _, err = producer.SendMessage(msg)
	if err != nil {
		return err
	}

	// fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)
	return nil
}

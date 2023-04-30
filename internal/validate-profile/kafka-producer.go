package validate_profile

import (
	"fmt"

	s3ObjectCreatedSchema "github.com/michaelprice232/image-processor-api/internal/s3-object-created-schema"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

func (c *Client) sendMessage(topic string, key []byte, event s3ObjectCreatedSchema.AWSEvent) error {
	log.Debugf("Sending Kafka message to topic %s (key: %v)", topic, string(key))

	// todo: extend Value object to include the failure reason
	jsonStr, err := s3ObjectCreatedSchema.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling AWSEvent into JSON string: %v", err)
	}

	err = c.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          jsonStr,
	}, nil)
	if err != nil {
		return fmt.Errorf("error sending message to topic %s with key %s and value %s", topic, key, jsonStr)
	}

	return nil
}

//// Go-routine to handle message delivery reports and possibly other event types (errors, stats, etc)
//go func() {
//	for e := range c.kafkaProducer.Events() {
//		switch ev := e.(type) {
//		case *kafka.Message:
//			if ev.TopicPartition.Error != nil {
//				fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
//			} else {
//				fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
//					*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
//			}
//		}
//	}
//}()

// Wait for all messages to be delivered
//p.Flush(15 * 1000)
//p.Close()

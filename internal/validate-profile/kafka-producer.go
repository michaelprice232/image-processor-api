package validate_profile

import (
	"fmt"
	"time"

	s3Object "github.com/michaelprice232/image-processor-api/internal/s3-object-created-schema"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

// kafkaResponseEvent is the event sent to Kafka with the outcome of the image validation
type kafkaResponseEvent struct {
	Event        s3Object.AWSEvent `json:"event"`             // Raw EventBridge S3 event. Includes the bucket and key name
	Outcome      string            `json:"outcome"`           // "success" || "failed"
	ErrorMessage string            `json:"message,omitempty"` // Required for failed validations only. Includes the failure reason
}

// sendMessage sends a single kafkaResponseEvent event to Kafka
func (c *Client) sendMessage(topic string, event kafkaResponseEvent) error {
	key := []byte(fmt.Sprintf("%s/%s-%d", event.Event.Detail.Bucket.Name, event.Event.Detail.Object.Key, time.Now().Unix()))

	log.Debugf("Sending Kafka message to topic %s (key: %v)", topic, string(key))

	jsonStr, err := s3Object.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshalling kafkaResponseEvent into JSON string: %v", err)
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

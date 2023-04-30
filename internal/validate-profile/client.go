package validate_profile

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Client struct {
	rekognitionClient *rekognition.Client
	kafkaProducer     *kafka.Producer
	successKafkaTopic string
	failedKafkaTopic  string
}

func NewClient(successTopic, failedTopic string) (*Client, error) {
	// todo: read Kafka properties from envars
	kafkaConfigMap := kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	}
	kafkaProducer, err := kafka.NewProducer(&kafkaConfigMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka Producer: %v", err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load SDK configuration: %v", err)
	}

	return &Client{
		rekognitionClient: rekognition.NewFromConfig(cfg),
		kafkaProducer:     kafkaProducer,
		successKafkaTopic: successTopic,
		failedKafkaTopic:  failedTopic,
	}, nil
}

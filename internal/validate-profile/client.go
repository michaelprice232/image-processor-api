package validate_profile

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
)

type Client struct {
	client *rekognition.Client
}

func NewClient() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load SDK configuration: %v", err)
	}

	return &Client{client: rekognition.NewFromConfig(cfg)}, nil
}

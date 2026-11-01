package messaging

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-catalog-service/internal/pkg/env"
)

// SQSClient is the subset of the SQS API the consumer needs (fakeable in tests).
type SQSClient interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// NewSQSClient reads AWS_REGION (SDK default chain) and AWS_ENDPOINT_URL (local MiniStack only).
func NewSQSClient(ctx context.Context) (*sqs.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(env.GetString("AWS_REGION", "us-east-1")))
	if err != nil {
		return nil, err
	}
	endpoint := env.GetString("AWS_ENDPOINT_URL", "")
	return sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	}), nil
}

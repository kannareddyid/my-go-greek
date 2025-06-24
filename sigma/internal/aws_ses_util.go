package internal

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

const (
	AwsSenderEmail = "no-reply@dojo86.com"
)

type AwsSesClass struct {
	sesClient *ses.Client
}

func NewAwsSesClass(ctx context.Context, sesClientDi *ses.Client) *AwsSesClass {
	return &AwsSesClass{
		sesClient: sesClientDi,
	}
}

func (s *AwsSesClass) SendEmail(ctx context.Context, recipient, subject, body string) error {
	input := &ses.SendEmailInput{
		Source: aws.String(AwsSenderEmail),
		Destination: &types.Destination{
			ToAddresses: []string{recipient},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String(subject),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(body),
				},
			},
		},
	}

	result, err := s.sesClient.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully. Message ID: %s", *result.MessageId)
	return nil
}
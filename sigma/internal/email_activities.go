package internal

import (
	"context"
	"log"
)

type EmailActivityClass struct {
	sesClient *AwsSesClass
}

func NewEmailActivityClass(awsSesClassDi *AwsSesClass) *EmailActivityClass {
	return &EmailActivityClass{
		sesClient: awsSesClassDi,
	}
}

func (e *EmailActivityClass) SendScheduledEmail1(ctx context.Context, email string) error {
	subject := "Scheduled Email 1"
	body := "This is the first email sent from Temporal Workflow."
	err := e.sesClient.SendEmail(ctx, email, subject, body)
	if err != nil {
		log.Printf("failed to send email 1: %v", err)
		return err
	}
	log.Printf("Successfully sent email 1 to %s", email)
	return nil
}

func (e *EmailActivityClass) SendScheduledEmail2(ctx context.Context, email string) error {
	subject := "Scheduled Email 2"
	body := "This is the second email sent from the same Temporal Workflow."
	err := e.sesClient.SendEmail(ctx, email, subject, body)
	if err != nil {
		log.Printf("failed to send email 2: %v", err)
		return err
	}
	log.Printf("Successfully sent email 2 to %s", email)
	return nil
}
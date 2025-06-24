package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sigma/internal"

	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

const (
	TaskQueue         = "SIGMA_EMAIL_TASK_QUEUE"
	WorkflowID        = "sigma_email_workflow"
	CronExpr          = "*/2 * * * *" // Every 2 minutes
	ScheduleID        = "sigma-schedule-id"
)

func main() {

	// Load .env.local file
	if err := godotenv.Load(".env.local"); err != nil {
		log.Fatalf("Error loading .env.local file: %v", err)
	}

	// Retrieve environment variables
	awsRegion := os.Getenv("AWS_REGION")
	awsSenderEmail := os.Getenv("AWS_SENDER_EMAIL")
	awsSesAccessKeyID := os.Getenv("AWS_SES_ACCESS_KEY_ID")
	awsSesSecretAccessKey := os.Getenv("AWS_SES_SECRET_ACCESS_KEY")
	toEmailAddress := os.Getenv("TO_EMAIL_ADDRESS")
	temporalHostPort := os.Getenv("TEMPORAL_HOST_PORT")
	temporalNamespace := os.Getenv("TEMPORAL_NAMESPACE")
	temporalKey := os.Getenv("TEMPORAL_KEY")

	// Validate required environment variables
	requiredEnvVars := map[string]string{
		"AWS_REGION":              awsRegion,
		"AWS_SENDER_EMAIL":        awsSenderEmail,
		"AWS_SES_ACCESS_KEY_ID":   awsSesAccessKeyID,
		"AWS_SES_SECRET_ACCESS_KEY": awsSesSecretAccessKey,
		"TO_EMAIL_ADDRESS":         toEmailAddress,
		"TEMPORAL_HOST_PORT":      temporalHostPort,
		"TEMPORAL_NAMESPACE":      temporalNamespace,
		"TEMPORAL_KEY":            temporalKey,
	}
	for key, value := range requiredEnvVars {
		if value == "" {
			log.Fatalf("Missing required environment variable: %s", key)
		}
	}

	ctx := context.Background()

	// Initialize AWS SES client
	awsSesCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				awsSesAccessKeyID,
				awsSesSecretAccessKey,
				"",
			),
		),
	)
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}
	sesClient := ses.NewFromConfig(awsSesCfg)
	awsSesClass := internal.NewAwsSesClass(ctx, sesClient)
	emailActivityClass := internal.NewEmailActivityClass(ctx, awsSesClass)
	emailWorkflowClass := internal.NewEmailWorkflowClass(ctx, toEmailAddress, emailActivityClass)

	// Connect to Temporal Cloud
	temporalClient, err := client.Dial(client.Options{
		HostPort:  temporalHostPort,
		Namespace: temporalNamespace,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{},
		},
		Credentials: client.NewAPIKeyStaticCredentials(temporalKey),
	})
	if err != nil {
		log.Fatalf("Unable to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	// Start the worker
	w := worker.New(temporalClient, TaskQueue, worker.Options{})
	
	// SOLUTION 1: If you modified EmailWorkflow to not require activities parameter
	w.RegisterWorkflow(emailWorkflowClass.EmailWorkflow)
	w.RegisterActivity(emailActivityClass.SendScheduledEmail1)
	w.RegisterActivity(emailActivityClass.SendScheduledEmail2)

	// Channel to handle worker errors
	workerErr := make(chan error, 1)
	go func() {
		err := w.Run(worker.InterruptCh())
		if err != nil {
			workerErr <- err
		}
	}()

	// Give worker a moment to start
	time.Sleep(2 * time.Second)
	log.Printf("Worker started")

	// Create or update schedule
	temporalClass := internal.NewTemporalClient(temporalClient, *emailWorkflowClass)

	// SOLUTION 1: Use modified EmailWorkflow
	err = temporalClass.CreateOrUpdateSchedule(ctx, ScheduleID, CronExpr, WorkflowID, TaskQueue)

	if err != nil {
		log.Fatalf("Failed to create or update schedule: %v", err)
	}
	log.Printf("Schedule created or updated successfully")

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal or worker error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
	case err := <-workerErr:
		log.Printf("Worker error: %v", err)
	}

	log.Println("Service shutting down...")
}
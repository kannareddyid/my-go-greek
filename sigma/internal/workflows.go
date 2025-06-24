// internal/workflow.go
package internal

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Keep your original workflow
func EmailWorkflow(ctx workflow.Context, email string, activities *EmailActivityClass) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting EmailWorkflow", "email", email)

	// Define activity options with proper timeouts
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute activities concurrently using futures
	future1 := workflow.ExecuteActivity(ctx, activities.SendScheduledEmail1, email)
	future2 := workflow.ExecuteActivity(ctx, activities.SendScheduledEmail2, email)

	// Wait for both activities to complete
	var err1, err2 error

	err1 = future1.Get(ctx, nil)
	if err1 != nil {
		logger.Error("SendScheduledEmail1 failed", "error", err1)
	}

	err2 = future2.Get(ctx, nil)
	if err2 != nil {
		logger.Error("SendScheduledEmail2 failed", "error", err2)
	}

	// Handle errors based on your business logic
	if err1 != nil && err2 != nil {
		logger.Error("Both email activities failed")
		return err1
	}

	if err1 != nil || err2 != nil {
		logger.Warn("One email activity failed, but continuing workflow")
	}

	logger.Info("EmailWorkflow completed successfully")
	return nil
}

// New wrapper workflow for scheduling - only takes serializable parameters
func ScheduledEmailWorkflow(ctx workflow.Context, email string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting ScheduledEmailWorkflow", "email", email)

	// Define activity options with proper timeouts
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute activities by name (they'll be resolved at runtime)
	future1 := workflow.ExecuteActivity(ctx, "SendScheduledEmail1", email)
	future2 := workflow.ExecuteActivity(ctx, "SendScheduledEmail2", email)

	// Wait for both activities to complete
	var err1, err2 error

	err1 = future1.Get(ctx, nil)
	if err1 != nil {
		logger.Error("SendScheduledEmail1 failed", "error", err1)
	}

	err2 = future2.Get(ctx, nil)
	if err2 != nil {
		logger.Error("SendScheduledEmail2 failed", "error", err2)
	}

	// Handle errors based on your business logic
	if err1 != nil && err2 != nil {
		logger.Error("Both email activities failed")
		return err1
	}

	if err1 != nil || err2 != nil {
		logger.Warn("One email activity failed, but continuing workflow")
	}

	logger.Info("ScheduledEmailWorkflow completed successfully")
	return nil
}
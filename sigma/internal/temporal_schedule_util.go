package internal

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
)

type TemporalClass struct {
	temporalClient client.Client
}

func NewTemporalClient(temporalClientDi client.Client) *TemporalClass {
	return &TemporalClass{
		temporalClient: temporalClientDi,
	}
}

func (tc *TemporalClass) CreateOrUpdateSchedule(
	ctx context.Context,
	scheduleID string,
	cronExpression string,
	workflowID string,
	workflowName string, // Use string for registered workflow name
	taskQueue string,
	args ...interface{},
) error {
	scheduleClient := tc.temporalClient.ScheduleClient()

	// Get the schedule handle
	scheduleHandle := scheduleClient.GetHandle(ctx, scheduleID)

	// Check if the schedule exists
	_, err := scheduleHandle.Describe(ctx)
	if err != nil {
		_, err = scheduleClient.Create(ctx, client.ScheduleOptions{
			ID: scheduleID,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{cronExpression},
			},
			Action: &client.ScheduleWorkflowAction{
				Workflow:     workflowName, // Set to nil since we're using WorkflowName
				ID:           workflowID,
				TaskQueue:    taskQueue,
				Args:         args,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create schedule: %w", err)
		}
		log.Printf("Schedule %s created with workflow ID %s (note: Temporal may append timestamps to workflow ID)", scheduleID, workflowID)
		return nil
	}

	// Define the update function
	updateSchedule := func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
		newSpec := client.ScheduleSpec{
			CronExpressions: []string{cronExpression},
		}

		newState := client.ScheduleState{
			Paused: false,
		}

		newAction := client.ScheduleWorkflowAction{
			Workflow:     workflowName,
			ID:           workflowID,
			TaskQueue:    taskQueue,
			Args:         args,
		}

		return &client.ScheduleUpdate{
			Schedule: &client.Schedule{
				Spec:   &newSpec,
				State:  &newState,
				Action: &newAction,
			},
		}, nil
	}

	// Apply the update
	err = scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: updateSchedule,
	})
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}
	log.Printf("Schedule %s updated with workflow ID %s (note: Temporal may append timestamps to workflow ID)", scheduleID, workflowID)
	return nil
}
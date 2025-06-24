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
	workflowFunc interface{}, // The actual workflow function, not a string
	taskQueue string,
	args ...interface{},
) error {
	scheduleClient := tc.temporalClient.ScheduleClient()

	// Get the schedule handle
	scheduleHandle := scheduleClient.GetHandle(ctx, scheduleID)

	// Check if the schedule exists
	_, err := scheduleHandle.Describe(ctx)
	if err != nil {
		// Schedule doesn't exist, create it
		_, err = scheduleClient.Create(ctx, client.ScheduleOptions{
			ID: scheduleID,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{cronExpression},
			},
			Action: &client.ScheduleWorkflowAction{
				Workflow:  workflowFunc, // Pass the actual workflow function
				ID:        workflowID,
				TaskQueue: taskQueue,
				Args:      args,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create schedule: %w", err)
		}
		log.Printf("Schedule %s created with workflow ID %s", scheduleID, workflowID)
		return nil
	}

	// Schedule exists, update it
	err = scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			// Create new schedule configuration using a pointer to the existing schedule
			// schedule := input.Description.Schedule
			
			// Create updated schedule with new configuration
			updatedSchedule := &client.Schedule{
				Spec: &client.ScheduleSpec{
					CronExpressions: []string{cronExpression},
				},
				State: &client.ScheduleState{
					Paused: false,
				},
				Action: &client.ScheduleWorkflowAction{
					Workflow:  workflowFunc,
					ID:        workflowID,
					TaskQueue: taskQueue,
					Args:      args,
				},
			}

			return &client.ScheduleUpdate{
				Schedule: updatedSchedule,
			}, nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}
	
	log.Printf("Schedule %s updated with workflow ID %s", scheduleID, workflowID)
	return nil
}
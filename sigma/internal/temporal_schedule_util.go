package internal

import (
	"context"
	"fmt"
	"log"
	"go.temporal.io/sdk/client"
)

type TemporalClass struct {
	temporalClient client.Client
	emailWorkflowClass EmailWorkflowClass 
}

func NewTemporalClient(temporalClientDi client.Client, emailWorkflowClassDi EmailWorkflowClass) *TemporalClass {
	return &TemporalClass{
		temporalClient: temporalClientDi,
		emailWorkflowClass: emailWorkflowClassDi,
	}
}

func (tc *TemporalClass) CreateOrUpdateSchedule(
// CreateOrUpdateSchedule creates a new schedule or updates an existing one.
	ctx context.Context,
	scheduleID string,
	cronExpression string,
	workflowID string,
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
				Workflow:  tc.emailWorkflowClass.EmailWorkflow, //"EmailWorkflow",
				ID:        workflowID,
				TaskQueue: taskQueue,
				Args:      args,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create schedule: %w", err)
		}
		log.Printf("Schedule %s created", scheduleID)
		return nil
	}

	// Schedule exists, update it
	err = scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			// Get the current schedule, but create a new one to avoid modifying the original
			currentSchedule := input.Description.Schedule
			// Check if currentSchedule is the zero value
			if currentSchedule == (client.Schedule{}) {
				return nil, fmt.Errorf("current schedule is zero value")
			}

			// Create a new schedule structure with updated values
			updatedSchedule := &client.Schedule{
				Spec: &client.ScheduleSpec{
					CronExpressions: []string{cronExpression},
				},
				// State: &client.ScheduleState{
				// 	Paused: false,
				// },
				
				State:  currentSchedule.State,

				Action: &client.ScheduleWorkflowAction{
					Workflow:  tc.emailWorkflowClass.EmailWorkflow, //"EmailWorkflow",
					ID:        workflowID,
					TaskQueue: taskQueue,
					Args:      args,
				},
				
				// Preserve other fields from the current schedule
				Policy: currentSchedule.Policy,
			}

			return &client.ScheduleUpdate{
				Schedule: updatedSchedule,
			}, nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update schedule: %w", err)
	}
	log.Printf("Schedule %s updated", scheduleID)
	return nil
}
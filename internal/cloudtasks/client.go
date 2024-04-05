package cloudtasks

import (
	"context"
	"fmt"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "cloud.google.com/go/cloudtasks/apiv2/cloudtaskspb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func NewClient(ctx context.Context) (*cloudtasks.Client, error) {
    // Initialize and return a new Cloud Tasks client
    return cloudtasks.NewClient(ctx)
}

func CreateTask(client *cloudtasks.Client, projectID, locationID, queueID, targetURL string, scheduleTime time.Time) (*taskspb.Task, error) {
    // Create a new task with the specified target URL and schedule time
	fmt.Printf("Creating task with target URL: %s\n", targetURL)
    req := &taskspb.CreateTaskRequest{
        Parent: fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, locationID, queueID),
        Task: &taskspb.Task{
            MessageType: &taskspb.Task_HttpRequest{
                HttpRequest: &taskspb.HttpRequest{
                    HttpMethod: taskspb.HttpMethod_POST,
                    Url:        targetURL,
                },
            },
            ScheduleTime: timestamppb.New(scheduleTime),
        },
    }
	fmt.Printf("Request: %v\n", req)
    return client.CreateTask(context.Background(), req)
}

package cloudtasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/iterator"

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
    return client.CreateTask(context.Background(), req)
}

func DeleteTask(client *cloudtasks.Client, projectID, locationID, queueID, taskID string) error {
    // Delete the task with the specified ID
    fmt.Printf("Deleting task with ID: %s\n", taskID)
    req := &taskspb.DeleteTaskRequest{
        Name: fmt.Sprintf("projects/%s/locations/%s/queues/%s/tasks/%s", projectID, locationID, queueID, taskID),
    }
    return client.DeleteTask(context.Background(), req)
}

func DeleteAllTasks(client *cloudtasks.Client, projectID, locationID, queueID string) error {
    // Delete all tasks in the specified queue
    fmt.Printf("Deleting all tasks in queue: %s\n", queueID)
    req := &taskspb.ListTasksRequest{
        Parent: fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, locationID, queueID),
    }
    it := client.ListTasks(context.Background(), req)
    for {
        task, err := it.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return err
        }
        // Extract the task ID from the task name
        taskID := extractTaskID(task.Name)
        if err := DeleteTask(client, projectID, locationID, queueID, taskID); err != nil {
            continue
        }
    }
    return nil
}

func extractTaskID(taskName string) string {
    // Extract the task ID from the task name
    parts := strings.Split(taskName, "/")
    if len(parts) >= 6 {
        return parts[5]
    }
    return ""
}

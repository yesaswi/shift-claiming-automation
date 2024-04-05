package cloudtasks

import (
	"context"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
)

func NewClient(ctx context.Context) (*cloudtasks.Client, error) {
	// Initialize and return a new Cloud Tasks client
	return cloudtasks.NewClient(ctx)
}
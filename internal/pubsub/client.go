package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
)

func NewClient(ctx context.Context, projectID string) (*pubsub.Client, error) {
	// Initialize and return a new Pub/Sub client
	return pubsub.NewClient(ctx, projectID)
}

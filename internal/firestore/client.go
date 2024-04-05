package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
)

func NewClient(ctx context.Context, projectID string, databaseID string) (*firestore.Client, error) {
	// Initialize and return a new Firestore client
	return firestore.NewClientWithDatabase(ctx, projectID, databaseID)
}
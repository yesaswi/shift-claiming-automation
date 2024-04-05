package shiftclaiming

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"go.uber.org/zap"
)

type Service struct {
	log            *zap.Logger
	firestoreClient *firestore.Client
}

func NewService(log *zap.Logger, firestoreClient *firestore.Client) *Service {
	return &Service{
		log:            log,
		firestoreClient: firestoreClient,
	}
}

func (s *Service) StartClaiming() error {
	s.log.Info("Starting shift claiming...")

	// Update the start/stop flag in Firestore
	configDoc := s.firestoreClient.Collection("configuration").Doc("config")
	_, err := configDoc.Set(context.Background(), map[string]interface{}{
		"startStopFlag": true,
	})
	if err != nil {
		return fmt.Errorf("failed to update start/stop flag: %v", err)
	}
	return nil
}

func (s *Service) StopClaiming() error {
	s.log.Info("Stopping shift claiming...")

	// Update the start/stop flag in Firestore
	configDoc := s.firestoreClient.Collection("configuration").Doc("config")
	_, err := configDoc.Set(context.Background(), map[string]interface{}{
		"startStopFlag": false,
	})
	if err != nil {
		return fmt.Errorf("failed to update start/stop flag: %v", err)
	}
	return nil
}

func (s *Service) ClaimShift() error {
	s.log.Info("Claiming shift...")

	// TODO: Implement the logic to claim a shift
	return nil
}


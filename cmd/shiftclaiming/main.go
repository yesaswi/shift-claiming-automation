package main

import (
	cloudtasks2 "cloud.google.com/go/cloudtasks/apiv2"
	firestore2 "cloud.google.com/go/firestore"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/yesaswi/shift-claiming-automation/internal/cloudtasks"
	"github.com/yesaswi/shift-claiming-automation/internal/firestore"
	"github.com/yesaswi/shift-claiming-automation/internal/shiftclaiming"
	"github.com/yesaswi/shift-claiming-automation/pkg/config"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf(`{"message": "Failed to load configuration", "error": "%v", "severity": "critical"}`+"\n", err)
		os.Exit(1)
	}

	// Initialize the Firestore client
	firestoreClient, err := firestore.NewClient(context.Background(), cfg.ProjectID, cfg.DatabaseID)
	if err != nil {
		fmt.Printf(`{"message": "Failed to initialize Firestore client", "error": "%v", "severity": "critical"}`+"\n", err)
		os.Exit(1)
	}
	defer func(firestoreClient *firestore2.Client) {
		err := firestoreClient.Close()
		if err != nil {
			fmt.Printf(`{"message": "Failed to close Firestore client", "error": "%v", "severity": "error"}`+"\n", err)
		}
	}(firestoreClient)

	// Initialize the Cloud Tasks client
	cloudTasksClient, err := cloudtasks.NewClient(context.Background())
	if err != nil {
		fmt.Printf(`{"message": "Failed to initialize Cloud Tasks client", "error": "%v", "severity": "critical"}`+"\n", err)
		os.Exit(1)
	}
	defer func(cloudTasksClient *cloudtasks2.Client) {
		err := cloudTasksClient.Close()
		if err != nil {
			fmt.Printf(`{"message": "Failed to close Cloud Tasks client", "error": "%v", "severity": "error"}`+"\n", err)
		}
	}(cloudTasksClient)

	// Initialize the Shift Claiming Service
	service := shiftclaiming.NewService(firestoreClient, cloudTasksClient)

	// Create a new HTTP router
	router := mux.NewRouter()

	// Register the start and stop command handlers
	router.HandleFunc("/start", service.HandleStartCommand).Methods(http.MethodPost)
	router.HandleFunc("/stop", service.HandleStopCommand).Methods(http.MethodPost)
	router.HandleFunc("/claim", service.HandleClaimCommand).Methods(http.MethodPost)

	// Start the HTTP server
	port := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf(`{"message": "Starting server on port %s", "severity": "info"}`+"\n", port)
	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf(`{"message": "Server error", "error": "%v", "severity": "critical"}`+"\n", err)
			os.Exit(1)
		}
	}()

	// Wait for an interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Shutdown the server gracefully
	fmt.Println(`{"message": "Shutting down the server...", "severity": "info"}`)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf(`{"message": "Server shutdown error", "error": "%v", "severity": "error"}`+"\n", err)
	}
	fmt.Println(`{"message": "Server stopped.", "severity": "info"}`)
}

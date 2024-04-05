package main

import (
	"context"
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
	"github.com/yesaswi/shift-claiming-automation/pkg/logger"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize the logger
	log, err := logger.NewLogger()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Initialize the Firestore client
	firestoreClient, err := firestore.NewClient(context.Background(), cfg.ProjectID, cfg.DatabaseID)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to initialize Firestore client: %v", err))
		os.Exit(1)
	}
	defer firestoreClient.Close()

	// Initialize the Cloud Tasks client
	cloudTasksClient, err := cloudtasks.NewClient(context.Background())
	if err != nil {
		log.Error(fmt.Sprintf("Failed to initialize Cloud Tasks client: %v", err))
		os.Exit(1)
	}
	defer cloudTasksClient.Close()

	// Initialize the Shift Claiming Service
	service := shiftclaiming.NewService(log, firestoreClient, cloudTasksClient)

	// Create a new HTTP router
	router := mux.NewRouter()

	// Register the start and stop command handlers
	router.HandleFunc("/start", service.HandleStartCommand).Methods(http.MethodPost)
	router.HandleFunc("/stop", service.HandleStopCommand).Methods(http.MethodPost)
	router.HandleFunc("/claim", service.HandleClaimCommand).Methods(http.MethodPost)

	// Start the HTTP server
	port := fmt.Sprintf(":%d", cfg.Port)
	log.Info(fmt.Sprintf("Starting server on port %s", port))
	server := &http.Server{
		Addr:    port,
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("Server error: %v", err))
			os.Exit(1)
		}
	}()

	// Wait for an interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Shutdown the server gracefully
	log.Info("Shutting down the server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("Server shutdown error: %v", err))
	}
	log.Info("Server stopped.")
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-web-starter/internal/config"
	"go-web-starter/internal/infrastructure/logger"
	"go-web-starter/internal/infrastructure/messaging"
)

func main() {
	// Create logger
	loggerConfig := &config.LoggerConfig{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	}
	
	appLogger, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}

	// Create Kafka configuration
	kafkaConfig := &config.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		GroupID: "go-web-starter-example",
	}

	// Create messaging manager
	messagingManager, err := messaging.NewManager(kafkaConfig, appLogger)
	if err != nil {
		log.Fatal("Failed to create messaging manager:", err)
	}
	defer messagingManager.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumers
	fmt.Println("=== Starting Message Consumers ===")
	startConsumers(ctx, messagingManager, appLogger)

	// Wait a moment for consumers to start
	time.Sleep(2 * time.Second)

	// Test message publishing
	fmt.Println("\n=== Publishing Messages ===")
	testMessagePublishing(ctx, messagingManager)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\n=== Messaging Example Running ===")
	fmt.Println("Press Ctrl+C to stop...")

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutdown signal received, stopping consumers...")

	// Stop all consumers
	messagingManager.StopAllConsumers()

	fmt.Println("Messaging example completed!")
}

func startConsumers(ctx context.Context, manager *messaging.Manager, logger *logger.Logger) {
	// Start user event consumer
	userHandler := messaging.MessageHandlerFunc(func(ctx context.Context, message *messaging.Message) error {
		logger.WithField("topic", message.Topic).Info("Processing user event")
		
		var envelope messaging.EnvelopedMessage
		if err := message.GetValueAsJSON(&envelope); err != nil {
			logger.WithError(err).Error("Failed to parse enveloped message")
			return err
		}

		logger.WithField("event_type", envelope.Metadata.EventType).WithField("source", envelope.Metadata.Source).Info("User event received")
		
		// Process based on event type
		switch envelope.Metadata.EventType {
		case messaging.EventUserCreated:
			var userEvent messaging.UserEvent
			if err := messaging.DeserializeJSON(envelope.Payload.([]byte), &userEvent); err == nil {
				logger.WithField("user_id", userEvent.UserID).Info("User created event processed")
			}
		case messaging.EventUserUpdated:
			var userEvent messaging.UserEvent
			if err := messaging.DeserializeJSON(envelope.Payload.([]byte), &userEvent); err == nil {
				logger.WithField("user_id", userEvent.UserID).Info("User updated event processed")
			}
		case messaging.EventUserDeleted:
			var userEvent messaging.UserEvent
			if err := messaging.DeserializeJSON(envelope.Payload.([]byte), &userEvent); err == nil {
				logger.WithField("user_id", userEvent.UserID).Info("User deleted event processed")
			}
		}

		return nil
	})

	if err := manager.StartUserEventConsumer(ctx, userHandler); err != nil {
		logger.WithError(err).Error("Failed to start user event consumer")
	} else {
		logger.Info("✓ User event consumer started")
	}

	// Start system event consumer
	systemHandler := messaging.MessageHandlerFunc(func(ctx context.Context, message *messaging.Message) error {
		logger.WithField("topic", message.Topic).Info("Processing system event")
		
		var envelope messaging.EnvelopedMessage
		if err := message.GetValueAsJSON(&envelope); err != nil {
			logger.WithError(err).Error("Failed to parse enveloped message")
			return err
		}

		logger.WithField("event_type", envelope.Metadata.EventType).WithField("source", envelope.Metadata.Source).Info("System event received")
		return nil
	})

	if err := manager.StartSystemEventConsumer(ctx, systemHandler); err != nil {
		logger.WithError(err).Error("Failed to start system event consumer")
	} else {
		logger.Info("✓ System event consumer started")
	}

	// Start notification consumer
	notificationHandler := messaging.MessageHandlerFunc(func(ctx context.Context, message *messaging.Message) error {
		logger.WithField("topic", message.Topic).Info("Processing notification event")
		
		var envelope messaging.EnvelopedMessage
		if err := message.GetValueAsJSON(&envelope); err != nil {
			logger.WithError(err).Error("Failed to parse enveloped message")
			return err
		}

		logger.WithField("event_type", envelope.Metadata.EventType).WithField("source", envelope.Metadata.Source).Info("Notification event received")
		
		// Process notification
		switch envelope.Metadata.EventType {
		case messaging.EventEmailNotification:
			logger.Info("Processing email notification")
		case messaging.EventSMSNotification:
			logger.Info("Processing SMS notification")
		}

		return nil
	})

	if err := manager.StartNotificationConsumer(ctx, notificationHandler); err != nil {
		logger.WithError(err).Error("Failed to start notification consumer")
	} else {
		logger.Info("✓ Notification consumer started")
	}

	// Start audit event consumer
	auditHandler := messaging.MessageHandlerFunc(func(ctx context.Context, message *messaging.Message) error {
		logger.WithField("topic", message.Topic).Info("Processing audit event")
		
		var envelope messaging.EnvelopedMessage
		if err := message.GetValueAsJSON(&envelope); err != nil {
			logger.WithError(err).Error("Failed to parse enveloped message")
			return err
		}

		logger.WithField("event_type", envelope.Metadata.EventType).WithField("source", envelope.Metadata.Source).Info("Audit event received")
		return nil
	})

	if err := manager.StartAuditEventConsumer(ctx, auditHandler); err != nil {
		logger.WithError(err).Error("Failed to start audit event consumer")
	} else {
		logger.Info("✓ Audit event consumer started")
	}
}

func testMessagePublishing(ctx context.Context, manager *messaging.Manager) {
	// Test user events
	fmt.Println("Publishing user events...")
	
	userData := map[string]interface{}{
		"username": "john_doe",
		"email":    "john@example.com",
		"name":     "John Doe",
	}

	if err := manager.PublishUserCreated(ctx, 123, userData); err != nil {
		fmt.Printf("Error publishing user created event: %v\n", err)
	} else {
		fmt.Println("✓ User created event published")
	}

	time.Sleep(500 * time.Millisecond)

	changes := map[string]interface{}{
		"name": "John Smith",
	}

	if err := manager.PublishUserUpdated(ctx, 123, changes); err != nil {
		fmt.Printf("Error publishing user updated event: %v\n", err)
	} else {
		fmt.Println("✓ User updated event published")
	}

	time.Sleep(500 * time.Millisecond)

	if err := manager.PublishUserDeleted(ctx, 123); err != nil {
		fmt.Printf("Error publishing user deleted event: %v\n", err)
	} else {
		fmt.Println("✓ User deleted event published")
	}

	// Test system events
	fmt.Println("\nPublishing system events...")
	
	serviceData := map[string]interface{}{
		"version":     "1.0.0",
		"environment": "development",
		"port":        8080,
	}

	if err := manager.PublishSystemStartup(ctx, serviceData); err != nil {
		fmt.Printf("Error publishing system startup event: %v\n", err)
	} else {
		fmt.Println("✓ System startup event published")
	}

	time.Sleep(500 * time.Millisecond)

	healthData := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"checks": map[string]interface{}{
			"database": "ok",
			"redis":    "ok",
		},
	}

	if err := manager.PublishHealthCheck(ctx, healthData); err != nil {
		fmt.Printf("Error publishing health check event: %v\n", err)
	} else {
		fmt.Println("✓ Health check event published")
	}

	// Test notification events
	fmt.Println("\nPublishing notification events...")

	notificationData := map[string]interface{}{
		"template": "welcome",
		"language": "en",
	}

	if err := manager.PublishEmailNotification(ctx, "john@example.com", "Welcome!", "Welcome to our service!", notificationData); err != nil {
		fmt.Printf("Error publishing email notification: %v\n", err)
	} else {
		fmt.Println("✓ Email notification published")
	}

	time.Sleep(500 * time.Millisecond)

	if err := manager.PublishSMSNotification(ctx, "+1234567890", "Welcome to our service!", notificationData); err != nil {
		fmt.Printf("Error publishing SMS notification: %v\n", err)
	} else {
		fmt.Println("✓ SMS notification published")
	}

	// Test audit events
	fmt.Println("\nPublishing audit events...")

	auditDetails := map[string]interface{}{
		"username": "john_doe",
		"email":    "john@example.com",
	}

	if err := manager.PublishAuditEvent(ctx, 123, "CREATE", "user", auditDetails, "192.168.1.1", "Mozilla/5.0"); err != nil {
		fmt.Printf("Error publishing audit event: %v\n", err)
	} else {
		fmt.Println("✓ Audit event published")
	}

	// Wait for messages to be processed
	time.Sleep(2 * time.Second)

	// Show metrics
	fmt.Println("\n=== Messaging Metrics ===")
	metrics := manager.GetMetrics()
	for key, value := range metrics {
		fmt.Printf("%s: %v\n", key, value)
	}
}
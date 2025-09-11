package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Message represents a message in the messaging system
type Message struct {
	Topic     string            `json:"topic"`
	Partition int32             `json:"partition"`
	Offset    int64             `json:"offset"`
	Key       []byte            `json:"key,omitempty"`
	Value     []byte            `json:"value"`
	Headers   map[string]string `json:"headers,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// GetKeyAsString returns the message key as a string
func (m *Message) GetKeyAsString() string {
	if m.Key == nil {
		return ""
	}
	return string(m.Key)
}

// GetValueAsString returns the message value as a string
func (m *Message) GetValueAsString() string {
	return string(m.Value)
}

// GetValueAsJSON unmarshals the message value as JSON into the provided destination
func (m *Message) GetValueAsJSON(dest interface{}) error {
	return json.Unmarshal(m.Value, dest)
}

// GetHeader returns a header value by key
func (m *Message) GetHeader(key string) (string, bool) {
	if m.Headers == nil {
		return "", false
	}
	value, exists := m.Headers[key]
	return value, exists
}

// MessageHandler defines the interface for handling messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, message *Message) error
}

// MessageHandlerFunc is a function type that implements MessageHandler
type MessageHandlerFunc func(ctx context.Context, message *Message) error

// HandleMessage implements MessageHandler interface
func (f MessageHandlerFunc) HandleMessage(ctx context.Context, message *Message) error {
	return f(ctx, message)
}

// MessageProducer defines the interface for message producers
type MessageProducer interface {
	SendMessage(topic string, key, value []byte) error
	SendJSONMessage(topic string, key string, value interface{}) error
	Close() error
}

// MessageConsumer defines the interface for message consumers
type MessageConsumer interface {
	ConsumeMessages(ctx context.Context, topics []string, handler MessageHandler) error
	Close() error
}

// MessagingService combines producer and consumer functionality
type MessagingService interface {
	MessageProducer
	MessageConsumer
	GetMetrics() map[string]interface{}
}

// KafkaMessagingService implements MessagingService using Kafka
type KafkaMessagingService struct {
	client *KafkaClient
}

// NewKafkaMessagingService creates a new Kafka-based messaging service
func NewKafkaMessagingService(client *KafkaClient) MessagingService {
	return &KafkaMessagingService{
		client: client,
	}
}

// SendMessage sends a message to a topic
func (k *KafkaMessagingService) SendMessage(topic string, key, value []byte) error {
	return k.client.SendMessage(topic, key, value)
}

// SendJSONMessage sends a JSON message to a topic
func (k *KafkaMessagingService) SendJSONMessage(topic string, key string, value interface{}) error {
	return k.client.SendJSONMessage(topic, key, value)
}

// ConsumeMessages starts consuming messages from topics
func (k *KafkaMessagingService) ConsumeMessages(ctx context.Context, topics []string, handler MessageHandler) error {
	return k.client.ConsumeMessages(ctx, topics, handler)
}

// GetMetrics returns messaging metrics
func (k *KafkaMessagingService) GetMetrics() map[string]interface{} {
	return k.client.GetProducerMetrics()
}

// Close closes the messaging service
func (k *KafkaMessagingService) Close() error {
	return k.client.Close()
}

// SerializeJSON serializes a value to JSON bytes
func SerializeJSON(value interface{}) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// DeserializeJSON deserializes JSON bytes to a value
func DeserializeJSON(data []byte, dest interface{}) error {
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// MessageMetadata contains metadata about a message
type MessageMetadata struct {
	MessageID   string            `json:"message_id"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	ContentType string            `json:"content_type"`
	Timestamp   time.Time         `json:"timestamp"`
	Source      string            `json:"source"`
	EventType   string            `json:"event_type"`
	Version     string            `json:"version"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// EnvelopedMessage wraps a message with metadata
type EnvelopedMessage struct {
	Metadata MessageMetadata `json:"metadata"`
	Payload  interface{}     `json:"payload"`
}

// NewEnvelopedMessage creates a new enveloped message
func NewEnvelopedMessage(eventType, source string, payload interface{}) *EnvelopedMessage {
	return &EnvelopedMessage{
		Metadata: MessageMetadata{
			MessageID:   generateMessageID(),
			ContentType: "application/json",
			Timestamp:   time.Now().UTC(),
			Source:      source,
			EventType:   eventType,
			Version:     "1.0",
			Headers:     make(map[string]string),
		},
		Payload: payload,
	}
}

// SetCorrelationID sets the correlation ID for the message
func (e *EnvelopedMessage) SetCorrelationID(correlationID string) {
	e.Metadata.CorrelationID = correlationID
}

// SetHeader sets a header value
func (e *EnvelopedMessage) SetHeader(key, value string) {
	if e.Metadata.Headers == nil {
		e.Metadata.Headers = make(map[string]string)
	}
	e.Metadata.Headers[key] = value
}

// GetHeader gets a header value
func (e *EnvelopedMessage) GetHeader(key string) (string, bool) {
	if e.Metadata.Headers == nil {
		return "", false
	}
	value, exists := e.Metadata.Headers[key]
	return value, exists
}

// ToJSON serializes the enveloped message to JSON
func (e *EnvelopedMessage) ToJSON() ([]byte, error) {
	return SerializeJSON(e)
}

// FromJSON deserializes JSON to an enveloped message
func (e *EnvelopedMessage) FromJSON(data []byte) error {
	return DeserializeJSON(data, e)
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	// Simple implementation - in production, you might want to use UUID
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// Common message types and topics
const (
	// User events
	TopicUserEvents    = "user.events"
	EventUserCreated   = "user.created"
	EventUserUpdated   = "user.updated"
	EventUserDeleted   = "user.deleted"
	EventUserActivated = "user.activated"
	EventUserDeactivated = "user.deactivated"

	// System events
	TopicSystemEvents = "system.events"
	EventSystemStartup = "system.startup"
	EventSystemShutdown = "system.shutdown"
	EventHealthCheck = "system.health_check"

	// Notification events
	TopicNotifications = "notifications"
	EventEmailNotification = "notification.email"
	EventSMSNotification = "notification.sms"
	EventPushNotification = "notification.push"

	// Audit events
	TopicAuditEvents = "audit.events"
	EventAuditLog = "audit.log"
)

// UserEvent represents a user-related event
type UserEvent struct {
	UserID    int64                  `json:"user_id"`
	Action    string                 `json:"action"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SystemEvent represents a system-related event
type SystemEvent struct {
	Service   string                 `json:"service"`
	Action    string                 `json:"action"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NotificationEvent represents a notification event
type NotificationEvent struct {
	Type      string                 `json:"type"`
	Recipient string                 `json:"recipient"`
	Subject   string                 `json:"subject,omitempty"`
	Content   string                 `json:"content"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// AuditEvent represents an audit event
type AuditEvent struct {
	UserID    int64                  `json:"user_id,omitempty"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
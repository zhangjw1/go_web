package messaging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessagingService is a mock implementation of MessagingService for testing
type MockMessagingService struct {
	mock.Mock
}

func (m *MockMessagingService) SendMessage(topic string, key, value []byte) error {
	args := m.Called(topic, key, value)
	return args.Error(0)
}

func (m *MockMessagingService) SendJSONMessage(topic string, key string, value interface{}) error {
	args := m.Called(topic, key, value)
	return args.Error(0)
}

func (m *MockMessagingService) ConsumeMessages(ctx context.Context, topics []string, handler MessageHandler) error {
	args := m.Called(ctx, topics, handler)
	return args.Error(0)
}

func (m *MockMessagingService) GetMetrics() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockMessagingService) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestMessage_GetKeyAsString(t *testing.T) {
	tests := []struct {
		name string
		key  []byte
		want string
	}{
		{
			name: "nil key should return empty string",
			key:  nil,
			want: "",
		},
		{
			name: "empty key should return empty string",
			key:  []byte{},
			want: "",
		},
		{
			name: "valid key should return string",
			key:  []byte("test-key"),
			want: "test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{Key: tt.key}
			got := msg.GetKeyAsString()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMessage_GetValueAsString(t *testing.T) {
	tests := []struct {
		name  string
		value []byte
		want  string
	}{
		{
			name:  "empty value should return empty string",
			value: []byte{},
			want:  "",
		},
		{
			name:  "valid value should return string",
			value: []byte("test-value"),
			want:  "test-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{Value: tt.value}
			got := msg.GetValueAsString()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMessage_GetValueAsJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   []byte
		dest    interface{}
		wantErr bool
	}{
		{
			name:    "valid JSON should unmarshal successfully",
			value:   []byte(`{"name": "test", "value": 123}`),
			dest:    &map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "invalid JSON should return error",
			value:   []byte(`{invalid json}`),
			dest:    &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{Value: tt.value}
			err := msg.GetValueAsJSON(tt.dest)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMessage_GetHeader(t *testing.T) {
	headers := map[string]string{
		"content-type": "application/json",
		"timestamp":    "1234567890",
	}

	msg := &Message{Headers: headers}

	tests := []struct {
		name      string
		key       string
		wantValue string
		wantExists bool
	}{
		{
			name:       "existing header should return value",
			key:        "content-type",
			wantValue:  "application/json",
			wantExists: true,
		},
		{
			name:       "non-existing header should return empty and false",
			key:        "non-existing",
			wantValue:  "",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := msg.GetHeader(tt.key)
			assert.Equal(t, tt.wantValue, value)
			assert.Equal(t, tt.wantExists, exists)
		})
	}
}

func TestMessageHandlerFunc(t *testing.T) {
	called := false
	handler := MessageHandlerFunc(func(ctx context.Context, message *Message) error {
		called = true
		return nil
	})

	msg := &Message{Topic: "test"}
	err := handler.HandleMessage(context.Background(), msg)

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestNewEnvelopedMessage(t *testing.T) {
	eventType := "test.event"
	source := "test-service"
	payload := map[string]interface{}{"key": "value"}

	envelope := NewEnvelopedMessage(eventType, source, payload)

	assert.Equal(t, eventType, envelope.Metadata.EventType)
	assert.Equal(t, source, envelope.Metadata.Source)
	assert.Equal(t, payload, envelope.Payload)
	assert.Equal(t, "application/json", envelope.Metadata.ContentType)
	assert.Equal(t, "1.0", envelope.Metadata.Version)
	assert.NotEmpty(t, envelope.Metadata.MessageID)
	assert.NotZero(t, envelope.Metadata.Timestamp)
	assert.NotNil(t, envelope.Metadata.Headers)
}

func TestEnvelopedMessage_SetCorrelationID(t *testing.T) {
	envelope := NewEnvelopedMessage("test.event", "test-service", nil)
	correlationID := "test-correlation-id"

	envelope.SetCorrelationID(correlationID)

	assert.Equal(t, correlationID, envelope.Metadata.CorrelationID)
}

func TestEnvelopedMessage_SetHeader(t *testing.T) {
	envelope := NewEnvelopedMessage("test.event", "test-service", nil)
	key := "custom-header"
	value := "custom-value"

	envelope.SetHeader(key, value)

	gotValue, exists := envelope.GetHeader(key)
	assert.True(t, exists)
	assert.Equal(t, value, gotValue)
}

func TestEnvelopedMessage_GetHeader(t *testing.T) {
	envelope := NewEnvelopedMessage("test.event", "test-service", nil)
	
	// Test non-existing header
	value, exists := envelope.GetHeader("non-existing")
	assert.False(t, exists)
	assert.Empty(t, value)

	// Test existing header
	envelope.SetHeader("test-key", "test-value")
	value, exists = envelope.GetHeader("test-key")
	assert.True(t, exists)
	assert.Equal(t, "test-value", value)
}

func TestEnvelopedMessage_ToJSON(t *testing.T) {
	payload := map[string]interface{}{"key": "value"}
	envelope := NewEnvelopedMessage("test.event", "test-service", payload)

	data, err := envelope.ToJSON()

	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it's valid JSON by unmarshaling back
	var result EnvelopedMessage
	err = DeserializeJSON(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, envelope.Metadata.EventType, result.Metadata.EventType)
}

func TestEnvelopedMessage_FromJSON(t *testing.T) {
	// Create original envelope
	payload := map[string]interface{}{"key": "value"}
	original := NewEnvelopedMessage("test.event", "test-service", payload)
	
	// Serialize to JSON
	data, err := original.ToJSON()
	assert.NoError(t, err)

	// Deserialize from JSON
	var result EnvelopedMessage
	err = result.FromJSON(data)

	assert.NoError(t, err)
	assert.Equal(t, original.Metadata.EventType, result.Metadata.EventType)
	assert.Equal(t, original.Metadata.Source, result.Metadata.Source)
}

func TestSerializeJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "valid object should serialize successfully",
			value:   map[string]interface{}{"key": "value"},
			wantErr: false,
		},
		{
			name:    "nil should serialize successfully",
			value:   nil,
			wantErr: false,
		},
		{
			name:    "function should fail to serialize",
			value:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := SerializeJSON(tt.value)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, data)
			}
		})
	}
}

func TestDeserializeJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		dest    interface{}
		wantErr bool
	}{
		{
			name:    "valid JSON should deserialize successfully",
			data:    []byte(`{"key": "value"}`),
			dest:    &map[string]interface{}{},
			wantErr: false,
		},
		{
			name:    "invalid JSON should return error",
			data:    []byte(`{invalid json}`),
			dest:    &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeserializeJSON(tt.data, tt.dest)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserEvent(t *testing.T) {
	userID := int64(123)
	action := EventUserCreated
	data := map[string]interface{}{"username": "test"}
	timestamp := time.Now()

	event := UserEvent{
		UserID:    userID,
		Action:    action,
		Data:      data,
		Timestamp: timestamp,
	}

	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, action, event.Action)
	assert.Equal(t, data, event.Data)
	assert.Equal(t, timestamp, event.Timestamp)
}

func TestSystemEvent(t *testing.T) {
	service := "test-service"
	action := EventSystemStartup
	data := map[string]interface{}{"version": "1.0.0"}
	timestamp := time.Now()

	event := SystemEvent{
		Service:   service,
		Action:    action,
		Data:      data,
		Timestamp: timestamp,
	}

	assert.Equal(t, service, event.Service)
	assert.Equal(t, action, event.Action)
	assert.Equal(t, data, event.Data)
	assert.Equal(t, timestamp, event.Timestamp)
}

func TestNotificationEvent(t *testing.T) {
	eventType := EventEmailNotification
	recipient := "test@example.com"
	subject := "Test Subject"
	content := "Test Content"
	data := map[string]interface{}{"template": "welcome"}
	timestamp := time.Now()

	event := NotificationEvent{
		Type:      eventType,
		Recipient: recipient,
		Subject:   subject,
		Content:   content,
		Data:      data,
		Timestamp: timestamp,
	}

	assert.Equal(t, eventType, event.Type)
	assert.Equal(t, recipient, event.Recipient)
	assert.Equal(t, subject, event.Subject)
	assert.Equal(t, content, event.Content)
	assert.Equal(t, data, event.Data)
	assert.Equal(t, timestamp, event.Timestamp)
}

func TestAuditEvent(t *testing.T) {
	userID := int64(123)
	action := "CREATE"
	resource := "user"
	details := map[string]interface{}{"username": "test"}
	ipAddress := "192.168.1.1"
	userAgent := "test-agent"
	timestamp := time.Now()

	event := AuditEvent{
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Details:   details,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Timestamp: timestamp,
	}

	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, action, event.Action)
	assert.Equal(t, resource, event.Resource)
	assert.Equal(t, details, event.Details)
	assert.Equal(t, ipAddress, event.IPAddress)
	assert.Equal(t, userAgent, event.UserAgent)
	assert.Equal(t, timestamp, event.Timestamp)
}
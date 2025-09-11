package messaging

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"

	"go-web-starter/internal/config"
	"go-web-starter/internal/infrastructure/logger"
)

// KafkaClient wraps Kafka producer and consumer functionality
type KafkaClient struct {
	config   *config.KafkaConfig
	logger   *logger.Logger
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
}

// NewKafkaClient creates a new Kafka client instance
func NewKafkaClient(cfg *config.KafkaConfig, log *logger.Logger) (*KafkaClient, error) {
	// Create Kafka configuration
	kafkaConfig := sarama.NewConfig()
	
	// Producer configuration
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = 3
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy
	kafkaConfig.Producer.Flush.Frequency = 500 * time.Millisecond
	kafkaConfig.Producer.Flush.Messages = 100
	
	// Consumer configuration
	kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	kafkaConfig.Consumer.Group.Session.Timeout = 10 * time.Second
	kafkaConfig.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	kafkaConfig.Consumer.Return.Errors = true
	
	// Version configuration
	kafkaConfig.Version = sarama.V2_6_0_0

	// Create producer
	producer, err := sarama.NewSyncProducer(cfg.Brokers, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Create consumer group
	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, kafkaConfig)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	log.WithField("brokers", cfg.Brokers).WithField("group_id", cfg.GroupID).Info("Kafka client created successfully")

	return &KafkaClient{
		config:   cfg,
		logger:   log,
		producer: producer,
		consumer: consumer,
	}, nil
}

// SendMessage sends a message to a Kafka topic
func (k *KafkaClient) SendMessage(topic string, key, value []byte) error {
	start := time.Now()
	
	message := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("timestamp"),
				Value: []byte(fmt.Sprintf("%d", time.Now().Unix())),
			},
		},
	}

	partition, offset, err := k.producer.SendMessage(message)
	duration := time.Since(start)

	if err != nil {
		k.logger.WithError(err).WithField("topic", topic).WithField("duration", duration).Error("Failed to send Kafka message")
		return fmt.Errorf("failed to send message to topic %s: %w", topic, err)
	}

	k.logger.LogKafkaMessage(topic, "SEND", len(value), duration)
	k.logger.WithField("topic", topic).WithField("partition", partition).WithField("offset", offset).WithField("duration", duration).Debug("Message sent successfully")

	return nil
}

// SendJSONMessage sends a JSON message to a Kafka topic
func (k *KafkaClient) SendJSONMessage(topic string, key string, value interface{}) error {
	jsonData, err := SerializeJSON(value)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	var keyBytes []byte
	if key != "" {
		keyBytes = []byte(key)
	}

	return k.SendMessage(topic, keyBytes, jsonData)
}

// ConsumeMessages starts consuming messages from specified topics
func (k *KafkaClient) ConsumeMessages(ctx context.Context, topics []string, handler MessageHandler) error {
	consumerHandler := &ConsumerGroupHandler{
		handler: handler,
		logger:  k.logger,
	}

	k.logger.WithField("topics", topics).WithField("group_id", k.config.GroupID).Info("Starting Kafka consumer")

	// Start consuming in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				k.logger.Info("Kafka consumer context cancelled")
				return
			default:
				if err := k.consumer.Consume(ctx, topics, consumerHandler); err != nil {
					k.logger.WithError(err).Error("Error consuming Kafka messages")
					// Wait before retrying
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()

	// Handle consumer errors
	go func() {
		for err := range k.consumer.Errors() {
			k.logger.WithError(err).Error("Kafka consumer error")
		}
	}()

	return nil
}

// GetProducerMetrics returns producer metrics
func (k *KafkaClient) GetProducerMetrics() map[string]interface{} {
	// Sarama doesn't expose detailed metrics directly
	// This is a placeholder for custom metrics collection
	return map[string]interface{}{
		"producer_active": k.producer != nil,
		"consumer_active": k.consumer != nil,
	}
}

// Close closes the Kafka client connections
func (k *KafkaClient) Close() error {
	var errors []string

	if k.producer != nil {
		if err := k.producer.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("producer close error: %v", err))
		} else {
			k.logger.Info("Kafka producer closed")
		}
	}

	if k.consumer != nil {
		if err := k.consumer.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("consumer close error: %v", err))
		} else {
			k.logger.Info("Kafka consumer closed")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("kafka close errors: %s", strings.Join(errors, "; "))
	}

	k.logger.Info("Kafka client closed successfully")
	return nil
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	handler MessageHandler
	logger  *logger.Logger
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info("Kafka consumer group session setup")
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info("Kafka consumer group session cleanup")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			start := time.Now()
			
			// Create message wrapper
			msg := &Message{
				Topic:     message.Topic,
				Partition: message.Partition,
				Offset:    message.Offset,
				Key:       message.Key,
				Value:     message.Value,
				Headers:   convertHeaders(message.Headers),
				Timestamp: message.Timestamp,
			}

			// Handle the message
			err := h.handler.HandleMessage(session.Context(), msg)
			duration := time.Since(start)

			if err != nil {
				h.logger.WithError(err).WithField("topic", message.Topic).WithField("partition", message.Partition).WithField("offset", message.Offset).WithField("duration", duration).Error("Failed to handle Kafka message")
				// Depending on your error handling strategy, you might want to:
				// 1. Continue processing (current behavior)
				// 2. Return the error to stop processing
				// 3. Send to a dead letter queue
				continue
			}

			// Mark message as processed
			session.MarkOffset(message.Topic, message.Partition, message.Offset+1, "")
			
			h.logger.LogKafkaMessage(message.Topic, "RECEIVE", len(message.Value), duration)
			h.logger.WithField("topic", message.Topic).WithField("partition", message.Partition).WithField("offset", message.Offset).WithField("duration", duration).Debug("Message processed successfully")

		case <-session.Context().Done():
			return nil
		}
	}
}

// convertHeaders converts Sarama headers to our header format
func convertHeaders(headers []*sarama.RecordHeader) map[string]string {
	result := make(map[string]string)
	for _, header := range headers {
		result[string(header.Key)] = string(header.Value)
	}
	return result
}
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bingaigc/gp/internal/domain/entity"
	"github.com/bingaigc/gp/pkg/errors"
)

// EventType represents the type of event
type EventType string

const (
	EventTypeSignalGenerated EventType = "signal.generated"
	EventTypeSignalAccepted  EventType = "signal.accepted"
	EventTypeSignalRejected  EventType = "signal.rejected"
	EventTypeOrderPlaced     EventType = "order.placed"
	EventTypeOrderFilled     EventType = "order.filled"
	EventTypeRiskAlert       EventType = "risk.alert"
	EventTypePriceUpdate     EventType = "price.update"
)

// Event represents a domain event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// KafkaClient handles message queue operations
type KafkaClient struct {
	brokers   []string
	topic     string
	connected bool
}

// NewKafkaClient creates a new Kafka client
func NewKafkaClient(brokers []string, topic string) *KafkaClient {
	return &KafkaClient{
		brokers:   brokers,
		topic:     topic,
		connected: false,
	}
}

// Connect establishes connection to Kafka
func (k *KafkaClient) Connect(ctx context.Context) error {
	// In production, this would initialize Kafka producer/consumer
	// For now, we simulate the connection
	k.connected = true
	return nil
}

// Publish sends an event to Kafka
func (k *KafkaClient) Publish(ctx context.Context, event *Event) error {
	if !k.connected {
		return errors.New(errors.ErrInfrastructure, "Kafka not connected")
	}

	// Validate event
	if event.ID == "" || event.Type == "" {
		return errors.New(errors.ErrValidation, "Invalid event: missing ID or Type")
	}

	// In production, this would publish to Kafka
	// For now, we log the event
	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, errors.ErrInfrastructure, "Failed to marshal event")
	}

	// Simulate Kafka publish
	_ = data
	return nil
}

// Subscribe listens for events from Kafka
func (k *KafkaClient) Subscribe(ctx context.Context, eventTypes []EventType, handler func(*Event) error) error {
	if !k.connected {
		return errors.New(errors.ErrInfrastructure, "Kafka not connected")
	}

	// In production, this would subscribe to Kafka topics
	// and call handler for each message
	return nil
}

// Close closes the Kafka connection
func (k *KafkaClient) Close() error {
	k.connected = false
	return nil
}

// SignalEvent creates a signal event
func SignalEvent(signal *entity.Signal, eventType EventType) *Event {
	return &Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    "alpha-detector",
		Data: map[string]interface{}{
			"stock_code":  signal.Stock.Code,
			"stock_name":  signal.Stock.Name,
			"signal_type": signal.Type.String(),
			"score":       signal.Score,
			"confidence":  signal.Confidence,
		},
	}
}

// EventBus coordinates event publishing and subscription
type EventBus struct {
	kafka *KafkaClient
}

// NewEventBus creates a new event bus
func NewEventBus(kafka *KafkaClient) *EventBus {
	return &EventBus{
		kafka: kafka,
	}
}

// PublishSignalGenerated publishes a signal generated event
func (e *EventBus) PublishSignalGenerated(ctx context.Context, signal *entity.Signal) error {
	event := SignalEvent(signal, EventTypeSignalGenerated)
	return e.kafka.Publish(ctx, event)
}

// PublishSignalAccepted publishes a signal accepted event
func (e *EventBus) PublishSignalAccepted(ctx context.Context, signal *entity.Signal) error {
	event := SignalEvent(signal, EventTypeSignalAccepted)
	return e.kafka.Publish(ctx, event)
}

// PublishSignalRejected publishes a signal rejected event
func (e *EventBus) PublishSignalRejected(ctx context.Context, signal *entity.Signal, reason string) error {
	event := SignalEvent(signal, EventTypeSignalRejected)
	event.Data["rejection_reason"] = reason
	return e.kafka.Publish(ctx, event)
}

// PublishRiskAlert publishes a risk alert event
func (e *EventBus) PublishRiskAlert(ctx context.Context, alertType string, message string, severity string) error {
	event := &Event{
		ID:        fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Type:      EventTypeRiskAlert,
		Timestamp: time.Now(),
		Source:    "risk-controller",
		Data: map[string]interface{}{
			"alert_type": alertType,
			"message":    message,
			"severity":   severity,
		},
	}
	return e.kafka.Publish(ctx, event)
}

package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"iot-control-service/internal/config"
	"iot-control-service/internal/models"
)

// Client wraps the MQTT client
type Client struct {
	client mqtt.Client
	config *config.Config
}

// NewClient creates a new MQTT client
func NewClient(cfg *config.Config) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.MQTT.Broker, cfg.MQTT.Port))
	opts.SetClientID(cfg.MQTT.ClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)

	if cfg.MQTT.Username != "" {
		opts.SetUsername(cfg.MQTT.Username)
	}
	if cfg.MQTT.Password != "" {
		opts.SetPassword(cfg.MQTT.Password)
	}

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("MQTT client connected")
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return &Client{
		client: client,
		config: cfg,
	}, nil
}

// PublishTelemetry publishes telemetry data to MQTT
func (c *Client) PublishTelemetry(deviceID string, telemetry *models.Telemetry) error {
	topic := fmt.Sprintf("mqtt/iot/%s/telemetry", deviceID)
	return c.publish(topic, telemetry)
}

// PublishCommand publishes a command to a device
func (c *Client) PublishCommand(deviceID string, command *models.DeviceCommand) error {
	topic := fmt.Sprintf("mqtt/iot/%s/command", deviceID)
	return c.publish(topic, command)
}

// PublishBroadcast publishes a broadcast message to all devices
func (c *Client) PublishBroadcast(message map[string]interface{}) error {
	topic := "mqtt/iot/broadcast/announcement"
	return c.publish(topic, message)
}

// SubscribeToTelemetry subscribes to telemetry from a device
func (c *Client) SubscribeToTelemetry(deviceID string, handler func(*models.Telemetry)) error {
	topic := fmt.Sprintf("mqtt/iot/%s/telemetry", deviceID)
	return c.subscribe(topic, func(topic string, payload []byte) {
		var telemetry models.Telemetry
		if err := json.Unmarshal(payload, &telemetry); err != nil {
			log.Printf("Failed to unmarshal telemetry: %v", err)
			return
		}
		handler(&telemetry)
	})
}

// SubscribeToAck subscribes to command acknowledgments from a device
func (c *Client) SubscribeToAck(deviceID string, handler func(*models.CommandAck)) error {
	topic := fmt.Sprintf("mqtt/iot/%s/ack", deviceID)
	return c.subscribe(topic, func(topic string, payload []byte) {
		var ack models.CommandAck
		if err := json.Unmarshal(payload, &ack); err != nil {
			log.Printf("Failed to unmarshal ack: %v", err)
			return
		}
		handler(&ack)
	})
}

// SubscribeToAllTelemetry subscribes to telemetry from all devices
func (c *Client) SubscribeToAllTelemetry(handler func(string, *models.Telemetry)) error {
	topic := "mqtt/iot/+/telemetry"
	return c.subscribe(topic, func(topic string, payload []byte) {
		var telemetry models.Telemetry
		if err := json.Unmarshal(payload, &telemetry); err != nil {
			log.Printf("Failed to unmarshal telemetry: %v", err)
			return
		}
		// Extract device ID from topic: mqtt/iot/{deviceId}/telemetry
		deviceID := extractDeviceIDFromTopic(topic)
		handler(deviceID, &telemetry)
	})
}

// SubscribeToAllAcks subscribes to acknowledgments from all devices
func (c *Client) SubscribeToAllAcks(handler func(string, *models.CommandAck)) error {
	topic := "mqtt/iot/+/ack"
	return c.subscribe(topic, func(topic string, payload []byte) {
		var ack models.CommandAck
		if err := json.Unmarshal(payload, &ack); err != nil {
			log.Printf("Failed to unmarshal ack: %v", err)
			return
		}
		// Extract device ID from topic: mqtt/iot/{deviceId}/ack
		deviceID := extractDeviceIDFromTopic(topic)
		handler(deviceID, &ack)
	})
}

// publish publishes a message to a topic
func (c *Client) publish(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	token := c.client.Publish(topic, c.config.MQTT.QoS, false, data)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to publish to %s: %w", topic, token.Error())
	}

	return nil
}

// subscribe subscribes to a topic with a handler
func (c *Client) subscribe(topic string, handler func(string, []byte)) error {
	token := c.client.Subscribe(topic, c.config.MQTT.QoS, func(client mqtt.Client, msg mqtt.Message) {
		handler(msg.Topic(), msg.Payload())
	})

	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", topic, token.Error())
	}

	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

// extractDeviceIDFromTopic extracts device ID from MQTT topic
func extractDeviceIDFromTopic(topic string) string {
	// Topic format: mqtt/iot/{deviceId}/telemetry or mqtt/iot/{deviceId}/ack
	parts := splitTopic(topic)
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

// splitTopic splits a topic string by '/'
func splitTopic(topic string) []string {
	var parts []string
	var current string
	for _, char := range topic {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// Disconnect disconnects from the MQTT broker
func (c *Client) Disconnect() {
	c.client.Disconnect(250)
	log.Println("MQTT client disconnected")
}

// IsConnected checks if the client is connected
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

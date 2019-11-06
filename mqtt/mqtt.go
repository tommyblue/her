package mqtt

import (
	"errors"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	mqttClient MQTT.Client
	topics     []string
	outCh      chan Message
}

type Message struct {
	Topic   string
	Message []byte
}

func NewClient(brokerUrl string, outCh chan Message) (*Client, error) {
	opts := MQTT.NewClientOptions().AddBroker(brokerUrl)
	opts.SetClientID("her")

	client := &Client{
		outCh:      outCh,
		mqttClient: MQTT.NewClient(opts),
	}

	return client, nil
}

func (c *Client) Connect() error {
	if token := c.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	if !c.mqttClient.IsConnected() {
		return errors.New("MQTT Client is disconnected")
	}

	return nil
}

func (c *Client) Subscribe(topic string) error {
	log.Info("Subscribing ", topic)
	if token := c.mqttClient.Subscribe(topic, 0, c.MsgCallback); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	c.topics = append(c.topics, topic)
	return nil
}

func (c *Client) Stop() error {
	log.Info("Stopping mqtt")
	for _, topic := range c.topics {
		if token := c.mqttClient.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	log.Info("Disconnetting MQTT")
	c.mqttClient.Disconnect(250)
	return nil
}

func (c *Client) MsgCallback(client MQTT.Client, msg MQTT.Message) {
	c.outCh <- Message{
		Topic:   msg.Topic(),
		Message: msg.Payload(),
	}
}

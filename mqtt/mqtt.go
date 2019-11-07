package mqtt

import (
	"errors"
	"os"
	"sync"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	mqttClient MQTT.Client
	topics     []string
	stopWg     *sync.WaitGroup
	shutdownCh chan os.Signal
	outCh      chan Message
	inCh       chan Message
}

type Message struct {
	Topic   string
	Message []byte
}

func NewClient(stopWg *sync.WaitGroup, shutdownCh chan os.Signal, brokerUrl string, inCh, outCh chan Message) (*Client, error) {
	opts := MQTT.NewClientOptions().AddBroker(brokerUrl)
	opts.SetClientID("her")

	client := &Client{
		inCh:       inCh,
		outCh:      outCh,
		mqttClient: MQTT.NewClient(opts),
		stopWg:     stopWg,
		shutdownCh: shutdownCh,
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

	go func() {
		for msg := range c.inCh {
			if err := c.Publish(msg); err != nil {
				log.Error(err)
			}
		}
	}()

	go func() {
		<-c.shutdownCh
		if err := c.stop(); err != nil {
			log.Error(err)
		}
		c.stopWg.Done()
	}()

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

func (c *Client) Publish(msg Message) error {
	token := c.mqttClient.Publish(msg.Topic, 0, true, msg.Message)
	token.Wait()
	return token.Error()
}

func (c *Client) stop() error {
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

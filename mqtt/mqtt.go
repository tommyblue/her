package mqtt

import (
	"bytes"
	"errors"
	"os"
	"sync"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tommyblue/her/her"
)

type Client struct {
	mqttClient   MQTT.Client
	topics       []string
	stopWg       *sync.WaitGroup
	shutdownCh   chan os.Signal
	outCh        chan her.Message
	inCh         chan her.Message
	lastMessages map[string]her.Message
}

func NewClient(stopWg *sync.WaitGroup, shutdownCh chan os.Signal, inCh, outCh chan her.Message) (*Client, error) {
	brokerUrl := viper.GetString("mqtt.broker_url")
	opts := MQTT.NewClientOptions().AddBroker(brokerUrl)
	opts.SetClientID("her")

	client := &Client{
		inCh:         inCh,
		outCh:        outCh,
		mqttClient:   MQTT.NewClient(opts),
		stopWg:       stopWg,
		shutdownCh:   shutdownCh,
		lastMessages: make(map[string]her.Message),
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

func (c *Client) Publish(msg her.Message) error {
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
	message := her.Message{
		Topic:   msg.Topic(),
		Message: msg.Payload(),
	}

	if !bytes.Equal(c.lastMessages[message.Topic].Message, message.Message) {
		c.outCh <- message
	}

	c.lastMessages[message.Topic] = message
}

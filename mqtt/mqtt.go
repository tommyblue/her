package mqtt

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tommyblue/her/her"
)

type Client struct {
	mqttClient    MQTT.Client
	subscriptions map[string]her.SubscriptionConf
	stopWg        *sync.WaitGroup
	shutdownCh    chan os.Signal
	outCh         chan her.Message
	inCh          chan her.Message
	lastMessages  map[string]her.Message
}

func NewClient(stopWg *sync.WaitGroup, shutdownCh chan os.Signal, inCh, outCh chan her.Message) (*Client, error) {
	brokerUrl := viper.GetString("mqtt.broker_url")
	opts := MQTT.NewClientOptions().AddBroker(brokerUrl)
	opts.SetClientID("her")

	client := &Client{
		subscriptions: make(map[string]her.SubscriptionConf),
		inCh:          inCh,
		outCh:         outCh,
		mqttClient:    MQTT.NewClient(opts),
		stopWg:        stopWg,
		shutdownCh:    shutdownCh,
		lastMessages:  make(map[string]her.Message),
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
			log.Debug("Received: ", msg)
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

func (c *Client) Subscribe(s her.SubscriptionConf) error {
	log.Info("Subscribing ", s.Topic, ", repeat: ", s.Repeat, ", repeat_only_if_different: ", s.RepeatOnlyIfDifferent)
	if token := c.mqttClient.Subscribe(s.Topic, 0, c.MsgCallback); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	c.subscriptions[s.Topic] = s
	return nil
}

func (c *Client) Publish(msg her.Message) error {
	token := c.mqttClient.Publish(msg.Topic, 0, true, msg.Message)
	token.Wait()
	return token.Error()
}

func (c *Client) stop() error {
	log.Info("Stopping mqtt")

	for topic := range c.subscriptions {
		if token := c.mqttClient.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			return token.Error()
		}
		delete(c.subscriptions, topic)
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

	if message.Topic == "" && (message.Message == nil || bytes.Equal(message.Message, []byte(""))) {
		return
	}

	s, ok := c.subscriptions[message.Topic]
	if !ok {
		log.Errorf("Cannot find topic %s among subscribed topics\n", message.Topic)
		return
	}

	log.Info(fmt.Sprintf("Received MQTT message: Topic: %s Message: %s", message.Topic, message.Message))

	if shouldSendMessage(s, message, c.lastMessages[message.Topic].Message) {
		log.Info(fmt.Sprintf("Sending %v", message))
		c.outCh <- message
	}

	c.lastMessages[message.Topic] = message
}

func shouldSendMessage(s her.SubscriptionConf, message her.Message, lastMessage []byte) bool {
	return s.Repeat && (!s.RepeatOnlyIfDifferent || !bytes.Equal(lastMessage, message.Message))
}

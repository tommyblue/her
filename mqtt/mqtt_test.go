package mqtt

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/tommyblue/her/her"
)

func TestShouldSendMessage(t *testing.T) {
	type args struct {
		s           her.SubscriptionConf
		message     her.Message
		lastMessage []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Always send (w/o last message)", args{
			s:           her.SubscriptionConf{Topic: "test", Repeat: true, RepeatOnlyIfDifferent: false},
			message:     her.Message{Topic: "test", Message: []byte("any")},
			lastMessage: nil,
		}, true},
		{"Always send (w/ last message)", args{
			s:           her.SubscriptionConf{Topic: "test", Repeat: true, RepeatOnlyIfDifferent: false},
			message:     her.Message{Topic: "test", Message: []byte("any")},
			lastMessage: []byte("any"),
		}, true},
		{"Never send (w/o last message)", args{
			s:           her.SubscriptionConf{Topic: "test", Repeat: false, RepeatOnlyIfDifferent: false},
			message:     her.Message{Topic: "test", Message: []byte("any")},
			lastMessage: nil,
		}, false},
		{"Never send (w/ last message)", args{
			s:           her.SubscriptionConf{Topic: "test", Repeat: false, RepeatOnlyIfDifferent: false},
			message:     her.Message{Topic: "test", Message: []byte("any")},
			lastMessage: []byte("any"),
		}, false},
		{"Send w/o repetitions (w/ last message)", args{
			s:           her.SubscriptionConf{Topic: "test", Repeat: true, RepeatOnlyIfDifferent: true},
			message:     her.Message{Topic: "test", Message: []byte("any")},
			lastMessage: []byte("any"),
		}, false},
		{"Send w/o repetitions (w/o last message)", args{
			s:           her.SubscriptionConf{Topic: "test", Repeat: true, RepeatOnlyIfDifferent: true},
			message:     her.Message{Topic: "test", Message: []byte("any")},
			lastMessage: nil,
		}, true},
		{"Send w/o repetitions (w/ different last message)", args{
			s:           her.SubscriptionConf{Topic: "test", Repeat: true, RepeatOnlyIfDifferent: true},
			message:     her.Message{Topic: "test", Message: []byte("any")},
			lastMessage: []byte("another"),
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSendMessage(tt.args.s, tt.args.message, tt.args.lastMessage); got != tt.want {
				t.Errorf("shouldSendMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mqttClientMock struct {
	tokenError  error
	isConnected bool
}

func (m mqttClientMock) IsConnected() bool      { return m.isConnected }
func (m mqttClientMock) IsConnectionOpen() bool { return true }
func (m mqttClientMock) Connect() MQTT.Token {
	return mqttTokenMock{
		errorReturn: m.tokenError,
	}
}
func (m mqttClientMock) Disconnect(quiesce uint) {}
func (m mqttClientMock) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	return mqttTokenMock{}
}
func (m mqttClientMock) Subscribe(topic string, qos byte, callback MQTT.MessageHandler) MQTT.Token {
	return mqttTokenMock{}
}
func (m mqttClientMock) SubscribeMultiple(filters map[string]byte, callback MQTT.MessageHandler) MQTT.Token {
	return mqttTokenMock{}
}
func (m mqttClientMock) Unsubscribe(topics ...string) MQTT.Token             { return mqttTokenMock{} }
func (m mqttClientMock) AddRoute(topic string, callback MQTT.MessageHandler) {}
func (m mqttClientMock) OptionsReader() MQTT.ClientOptionsReader             { return MQTT.ClientOptionsReader{} }

type mqttTokenMock struct {
	errorReturn error
}

func (t mqttTokenMock) Wait() bool                        { return true }
func (t mqttTokenMock) WaitTimeout(tm time.Duration) bool { return true }
func (t mqttTokenMock) Error() error                      { return t.errorReturn }

func TestConnect(t *testing.T) {
	t.Run("Connect Token error", func(t *testing.T) {
		client := &Client{
			mqttClient: mqttClientMock{
				tokenError:  fmt.Errorf("Error"),
				isConnected: false,
			},
		}
		err := client.Connect()
		if err == nil {
			t.Errorf("Should return error")
		}
	})
	t.Run("Connect IsConnected error", func(t *testing.T) {
		client := &Client{
			mqttClient: mqttClientMock{
				tokenError:  nil,
				isConnected: false,
			},
		}
		err := client.Connect()
		if err.Error() != errors.New("MQTT Client is disconnected").Error() {
			t.Errorf("Should return error")
		}
	})
}

func TestConnectInCh(t *testing.T) {
	inCh := make(chan her.Message)
	outCh := make(chan her.Message)
	shutdownCh := make(chan os.Signal, 1)
	l := make(map[string]her.Message)
	l["test"] = her.Message{
		Topic:   "t",
		Message: []byte("m"),
	}
	s := make(map[string]her.SubscriptionConf)
	s[l["test"].Topic] = her.SubscriptionConf{
		Label: "l",
	}
	wantMsg := "l: m\n"
	client := &Client{
		mqttClient: mqttClientMock{
			tokenError:  nil,
			isConnected: true,
		},
		inCh:          inCh,
		outCh:         outCh,
		shutdownCh:    shutdownCh,
		lastMessages:  l,
		subscriptions: s,
	}
	err := client.Connect()
	if err != nil {
		t.Errorf("Should not return error")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for msg := range outCh {
			if !bytes.Equal(msg.Message, []byte(wantMsg)) {
				t.Errorf("unexpected message %s", msg.Message)
			}
			break // exit after the first message
		}
		wg.Done()
	}()

	inCh <- her.Message{
		Command: "status",
	}
	wg.Wait()

	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
}

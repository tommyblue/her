package bot

import (
	"fmt"
	"os"
	"sync"
	"testing"

	viper "github.com/spf13/viper"

	"github.com/tommyblue/her/her"
)

type MockBot struct {
	commands         int
	connectRetErr    bool
	stopRetErr       bool
	receivedMsg      string
	receiveWg        *sync.WaitGroup
	calledReceive    bool
	receiveReturnErr bool
}

func (b *MockBot) Connect() error {
	if b.connectRetErr {
		return fmt.Errorf("Some error")
	}
	return nil
}
func (b *MockBot) Stop() error {
	if b.stopRetErr {
		return fmt.Errorf("Some error")
	}
	return nil
}
func (b *MockBot) SendMessage(msg string) error {
	b.receivedMsg = msg
	b.calledReceive = true
	b.receiveWg.Done()
	if b.receiveReturnErr {
		return fmt.Errorf("some error")
	}
	return nil
}
func (b *MockBot) AddCommand(c her.CommandConf) error {
	b.commands++
	return nil
}

func TestNewBot(t *testing.T) {
	var stopWg sync.WaitGroup
	shutdownCh := make(chan os.Signal)
	outCh := make(chan her.Message)
	inCh := make(chan her.Message)

	t.Run("Without config value", func(t *testing.T) {
		_, err := NewBot(&stopWg, shutdownCh, outCh, inCh)
		if err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("Without unkown config value", func(t *testing.T) {
		viper.Set("bot.type", "unknown")
		_, err := NewBot(&stopWg, shutdownCh, outCh, inCh)
		if err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("With correct config value but missing bot.token", func(t *testing.T) {
		viper.Set("bot.type", "telegram")
		viper.Set("bot.channel_id", 1)
		_, err := NewBot(&stopWg, shutdownCh, outCh, inCh)
		if err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("With correct config value but wrong bot.channel_id", func(t *testing.T) {
		viper.Set("bot.type", "telegram")
		viper.Set("bot.token", "token")
		viper.Set("bot.channel_id", 0)
		_, err := NewBot(&stopWg, shutdownCh, outCh, inCh)
		if err == nil {
			t.Errorf("Expected error")
		}
	})

	t.Run("With correct config value", func(t *testing.T) {
		viper.Set("bot.type", "telegram")
		viper.Set("bot.channel_id", 1)
		viper.Set("bot.token", "token")
		_, err := NewBot(&stopWg, shutdownCh, outCh, inCh)
		if err != nil {
			t.Errorf("Unexpected error")
		}
	})
}

func TestAddCommand(t *testing.T) {
	b := &Bot{
		bot: &MockBot{},
	}
	c := her.CommandConf{}
	_ = b.AddCommand(c)
	if b.bot.(*MockBot).commands != 1 {
		t.Errorf("Command not added")
	}
}

func TestConnect(t *testing.T) {
	t.Run("Return error at connect", func(t *testing.T) {
		b := &Bot{
			bot: &MockBot{
				connectRetErr: true,
			},
		}
		err := b.Connect()
		if err == nil {
			t.Errorf("Expected error")
		}
	})
	t.Run("Return error at stop", func(t *testing.T) {
		shutdownCh := make(chan os.Signal, 1)
		b := &Bot{
			shutdownCh: shutdownCh,
			bot: &MockBot{
				stopRetErr: true,
			},
		}
		var wg sync.WaitGroup
		wg.Add(1)
		var err error
		go func() {
			err = b.Connect()
			wg.Done()
		}()

		b.shutdownCh <- os.Interrupt
		wg.Wait()

		if err == nil {
			t.Errorf("Expected error")
		}
	})
	t.Run("Send message", func(t *testing.T) {
		inCh := make(chan her.Message)
		shutdownCh := make(chan os.Signal, 1)
		var stopWg sync.WaitGroup
		stopWg.Add(1)
		var receiveWg sync.WaitGroup
		b := &Bot{
			inCh:       inCh,
			shutdownCh: shutdownCh,
			stopWg:     &stopWg,
			bot: &MockBot{
				receiveWg: &receiveWg,
			},
		}
		var wg sync.WaitGroup
		wg.Add(1)
		var err error
		go func() {
			err = b.Connect()
			wg.Done()
		}()

		msg := her.Message{Topic: "topic", Message: []byte("msg")}
		receiveWg.Add(1)
		inCh <- msg
		want := "[topic] msg"
		receiveWg.Wait()
		if b.bot.(*MockBot).receivedMsg != want {
			t.Errorf("want: %s, got: %s", want, b.bot.(*MockBot).receivedMsg)
		}
		if b.bot.(*MockBot).calledReceive != true {
			t.Errorf("Message not received")
		}
		b.shutdownCh <- os.Interrupt
		wg.Wait()

		if err != nil {
			t.Errorf("Unexpected error")
		}
	})
	t.Run("Send empty message", func(t *testing.T) {
		inCh := make(chan her.Message)
		shutdownCh := make(chan os.Signal, 1)
		var stopWg sync.WaitGroup
		stopWg.Add(1)
		var receiveWg sync.WaitGroup
		b := &Bot{
			inCh:       inCh,
			shutdownCh: shutdownCh,
			stopWg:     &stopWg,
			bot: &MockBot{
				receiveWg: &receiveWg,
			},
		}
		var wg sync.WaitGroup
		wg.Add(1)
		var err error
		go func() {
			err = b.Connect()
			wg.Done()
		}()

		msg := her.Message{Topic: "", Message: []byte("")}
		receiveWg.Add(1)
		inCh <- msg
		b.shutdownCh <- os.Interrupt
		wg.Wait()

		if b.bot.(*MockBot).calledReceive == true {
			t.Errorf("Received unwanted message")
		}

		if err != nil {
			t.Errorf("Unexpected error")
		}
	})
	t.Run("Don't stop at bot.SendMessage error", func(t *testing.T) {
		inCh := make(chan her.Message)
		shutdownCh := make(chan os.Signal, 1)
		var stopWg sync.WaitGroup
		stopWg.Add(1)
		var receiveWg sync.WaitGroup
		b := &Bot{
			inCh:       inCh,
			shutdownCh: shutdownCh,
			stopWg:     &stopWg,
			bot: &MockBot{
				receiveWg:        &receiveWg,
				receiveReturnErr: true,
			},
		}
		var wg sync.WaitGroup
		wg.Add(1)
		var err error
		go func() {
			err = b.Connect()
			wg.Done()
		}()

		msg := her.Message{Topic: "topic", Message: []byte("msg")}
		receiveWg.Add(1)
		inCh <- msg
		want := "[topic] msg"
		receiveWg.Wait()
		if b.bot.(*MockBot).receivedMsg != want {
			t.Errorf("want: %s, got: %s", want, b.bot.(*MockBot).receivedMsg)
		}
		if b.bot.(*MockBot).calledReceive != true {
			t.Errorf("Message not received")
		}

		// bot.SendMessage returned an error, but we must still be able to send messages
		b.bot.(*MockBot).receiveReturnErr = false
		msg = her.Message{Topic: "topic2", Message: []byte("msg2")}
		receiveWg.Add(1)
		inCh <- msg
		want = "[topic2] msg2"
		receiveWg.Wait()
		if b.bot.(*MockBot).receivedMsg != want {
			t.Errorf("want: %s, got: %s", want, b.bot.(*MockBot).receivedMsg)
		}
		if b.bot.(*MockBot).calledReceive != true {
			t.Errorf("Message not received")
		}

		// Shut down
		b.shutdownCh <- os.Interrupt
		wg.Wait()

		if err != nil {
			t.Errorf("Unexpected error")
		}
	})
}

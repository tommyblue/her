package mqtt

import (
	"testing"

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

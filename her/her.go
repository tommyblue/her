package her

type Message struct {
	Topic   string
	Message []byte
}

type SubscriptionConf struct {
	Topic                 string
	Repeat                bool
	RepeatOnlyIfDifferent bool `mapstructure:"repeat_only_if_different"`
}

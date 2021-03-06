package her

type Message struct {
	Topic   string
	Message []byte
	Command string
}

type SubscriptionConf struct {
	Label                 string
	Topic                 string
	Repeat                bool
	RepeatOnlyIfDifferent bool `mapstructure:"repeat_only_if_different"`
	Alarm                 *AlarmConf
}

type CommandConf struct {
	Command     string
	Topic       string
	Message     string
	FeedbackMsg string `mapstructure:"feedback_message"`
	Help        string
}

type IntentConf struct {
	Action  string
	Room    string
	Topic   string
	Message string
}

type AlarmConf struct {
	Operator string
	Value    float64
}

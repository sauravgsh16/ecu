package clientnew

type ExchangeConfig struct {
	Type   string
	NoWait string
}

type PublishConfig struct {
	Mandatory bool
	Immediate bool
}

type ConsumerConfig struct {
	Consumer string
	NoWait   bool
}

type Config struct {
	Exchange ExchangeConfig
	Publish  PublishConfig
	Consumer ConsumerConfig
}

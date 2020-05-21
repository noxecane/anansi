package anansi

type BasicEnv struct {
	AppEnv string `default:"dev" split_words:"true"`
	Name   string `required:"true"`
	Port   int    `required:"true"`
	Scheme string `required:"true"`
	Secret []byte `required:"true"`
}

package config

import "github.com/caarlos0/env/v8"

type Options struct {
	DB struct {
		User                   string `env:"DB_USER" envDefault:"postgres"`
		Password               string `env:"DB_PASS"`
		Name                   string `env:"DB_NAME" envDefault:"postgres"`
		InstanceConnectionName string `env:"INSTANCE_CONNECTION_NAME"`
	}
	MathRequestTopic       string `env:"MATH_REQUEST_TOPIC,required"`
	MathResultSubscription string `env:"MATH_RESULT_SUBSCRIPTION,required"`
	GoogleCloudProject     string `env:"GOOGLE_CLOUD_PROJECT,required"`
	ListenAddr             string `env:"LISTEN_ADDR" envDefault:"0.0.0.0:8080"`
}

func Parse() (*Options, error) {
	opts := &Options{}
	if err := env.Parse(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

package config

import "github.com/caarlos0/env/v8"

type Options struct {
	DB struct {
		User                   string `env:"DB_USER" envDefault:"postgres"`
		Password               string `env:"DB_PASS"`
		Name                   string `env:"DB_NAME" envDefault:"postgres"`
		InstanceConnectionName string `env:"INSTANCE_CONNECTION_NAME"`
	}
	ListenAddr string `env:"LISTEN_ADDR" envDefault:"127.0.0.1:8080"`
}

func Parse() (*Options, error) {
	opts := &Options{}
	if err := env.Parse(opts); err != nil {
		return nil, err
	}
	return opts, nil
}

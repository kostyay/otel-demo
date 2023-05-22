package config

import (
	"fmt"

	"github.com/caarlos0/env/v8"
)

type DB struct {
	User     string `env:"DB_USER" envDefault:"postgres"`
	Password string `env:"DB_PASS"`
	Name     string `env:"DB_NAME" envDefault:"postgres"`
	Host     string `env:"DB_HOST" envDefault:"127.0.0.1"`
	Port     string `env:"DB_PORT"`
}

type Options struct {
	DB                     DB
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

func (db DB) DSN() string {
	portQuery := ""
	if db.Port != "" {
		portQuery = fmt.Sprintf(" port=%s ", db.Port)
	}
	return fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable %s", db.Host, db.User, db.Name, db.Password, portQuery)
}

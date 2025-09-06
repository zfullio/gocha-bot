package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/pkg/errors"
)

type Configuration struct {
	Log              Log    `env-prefix:"LOG_"         env-required:"true" yaml:"log"`
	TgToken          string `env:"TG_TOKEN"            env-required:"true" yaml:"tg_token"`
	DbDataSourceName string `env:"DB_DATA_SOURCE_NAME" env-required:"true" yaml:"db_data_source_name"`
	AlertCooldown    int    `env:"ALERT_COOLDOWN"      env-required:"true" yaml:"alert_cooldown"`
	UpdateInterval   int    `env:"UPDATE_INTERVAL"     env-required:"true" yaml:"update_interval"`
	Host             string `env:"HOST"                env-required:"true" yaml:"host"`
	Port             int    `env:"PORT"                env-required:"true" yaml:"port"`
	BaseUrl          string `env:"BASE_URL"            env-required:"true" yaml:"host"`
	IsDev            bool   `env:"IS_DEV"`
}

type Log struct {
	Level string `env:"LEVEL" env-required:"true" yaml:"level"`
}

func NewConfig() (*Configuration, error) {
	var envFiles []string

	if _, err := os.Stat(".env"); err == nil {
		log.Println("found .env file, adding it to env config files list")

		envFiles = append(envFiles, ".env")
	}

	cfg := &Configuration{}

	if len(envFiles) > 0 {
		for _, file := range envFiles {
			err := cleanenv.ReadConfig(file, cfg)
			if err != nil {
				return nil, errors.Wrapf(err, "error while read env config file: %s", err)
			}
		}
	} else {
		err := cleanenv.ReadEnv(cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "error while opening env: %s", err)
		}
	}

	return cfg, nil
}

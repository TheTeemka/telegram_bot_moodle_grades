package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	TelegramToken   string `mapstructure:"TELEGRAM_TOKEN" validate:"required"`
	TelegramID      int64  `mapstructure:"TELEGRAM_ID" validate:"required,min=1"`
	MoodleLoginSite string `mapstructure:"MOODLE_LOGIN_SITE" validate:"required,url"`
	MoodleGradeSite string `mapstructure:"MOODLE_GRADE_SITE" validate:"required,url"`
	MoodleUser      string `mapstructure:"MOODLE_USER" validate:"required"`
	MoodlePass      string `mapstructure:"MOODLE_PASS" validate:"required"`
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}

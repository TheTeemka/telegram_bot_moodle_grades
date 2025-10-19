package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	TelegramConfig TelegramConfig `mapstructure:",squash"`
	MoodleConfig   MoodleConfig   `mapstructure:",squash"`

	CsvFilesDir string `mapstructure:"CSV_FILES_DIR" validate:"required"`
}

type MoodleConfig struct {
	MoodleMainPage  string `mapstructure:"MOODLE_MAIN_PAGE" validate:"required,url"`
	MoodleLoginPage string `mapstructure:"MOODLE_LOGIN_PAGE" validate:"required,url"`
	MoodleGradePage string `mapstructure:"MOODLE_GRADE_PAGE" validate:"required,url"`
	MoodleUser      string `mapstructure:"MOODLE_USER" validate:"required"`
	MoodlePass      string `mapstructure:"MOODLE_PASS" validate:"required"`
}

type TelegramConfig struct {
	TelegramToken string `mapstructure:"TELEGRAM_TOKEN" validate:"required"`
	TelegramID    int64  `mapstructure:"TELEGRAM_ID" validate:"required,min=1"`
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

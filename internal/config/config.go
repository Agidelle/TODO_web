package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port     int    `mapstructure:"TODO_PORT"`
	DBdriver string `mapstructure:"TODO_DRIVER"`
	DBPath   string `mapstructure:"TODO_DBFILE"`
	Password string `mapstructure:"TODO_PASSWORD"`
	JWTKey   string `mapstructure:"TODO_JWTSECRET"`
}

func LoadCfg() (*Config, error) {
	//Конфиг для разработки из .env
	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println(".env файл не найден, используем переменные окружения или значения по умолчанию")
		} else {
			fmt.Println(err)
		}
	}

	viper.AutomaticEnv()

	viper.BindEnv("TODO_PORT")
	viper.BindEnv("TODO_DRIVER")
	viper.BindEnv("TODO_DBFILE")
	viper.BindEnv("TODO_PASSWORD")
	viper.BindEnv("TODO_JWTSECRET")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("конфигурации не загружены, ошибка: %v", err)
	}

	//Проверки конфига
	if cfg.DBdriver == "" {
		return nil, fmt.Errorf("не указан драйвер БД")
	}
	if cfg.DBPath == "" {
		return nil, fmt.Errorf("не указан путь к файлу БД")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("некорректный номер порта: %d", cfg.Port)
	}

	return &cfg, nil
}

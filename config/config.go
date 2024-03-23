package config

import (
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Redis        RedisConfig
	FloodControl FloodControlConfig
}

type RedisConfig struct {
	Host string
	Port string
	DB   int
}

type FloodControlConfig struct {
	MaxNumberOfRequests int64         // Максимальное количество запросов в единицу времени
	MaxNumberOfRetries  int           // Максимальное количество попыток обращения к БД
	TimeInterval        time.Duration // Время, за которое не должно быть превышений лимита вызовов
	FloodControlTTL     time.Duration // Время, через которое должен происходить повторный запрос к БД
}

// Init - функция получения конфига
func Init(path string) (*Config, error) {
	var cfg Config

	if err := parseConfigFile(path); err != nil {
		return nil, err
	}

	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// parseConfigFile - Получение расположения файла конфигурации
func parseConfigFile(filePath string) error {
	viper.SetConfigFile(filePath)
	return viper.ReadInConfig()
}

// unmarshal - чтение из файла конфигурации и занесение в структуру
func unmarshal(cfg *Config) error {
	if err := viper.UnmarshalKey("redis", &cfg.Redis); err != nil {
		return err
	}

	if err := viper.UnmarshalKey("flood-control", &cfg.FloodControl); err != nil {
		return err
	}

	return nil
}

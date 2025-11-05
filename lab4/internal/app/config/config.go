package config

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	// JWT/Session settings
	JWTSecret     string
	JWTTTLMinutes int
	CookieName    string
	// Redis settings (for JWT blacklist)
	RedisAddr     string
	RedisPassword string
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	// Defaults
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "dev-secret-change-me"
	}
	if cfg.JWTTTLMinutes == 0 {
		cfg.JWTTTLMinutes = 60
	}
	if cfg.CookieName == "" {
		cfg.CookieName = "access_token"
	}
	if cfg.RedisAddr == "" {
		cfg.RedisAddr = "127.0.0.1:6379"
	}

	log.Info("config parsed")

	return cfg, nil
}

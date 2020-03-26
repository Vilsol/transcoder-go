package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func InitializeConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()

	log.Info("Config initialized")
}

// Package config provides the application config based on config file
package config

import "github.com/spf13/viper"

type Config struct {
	TcpPort   string `mapstructure:"TCP_PORT"`
	HttpPort  string `mapstructure:"HTTP_PORT"`
	RedisPort string `mapstructure:"REDIS_PORT"`
	Address   string `mapstructure:"ADDRESS"`
	LogFile   string `mapstructure:"LOG_FILE"`
}

func LoadConfig() (config Config, err error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

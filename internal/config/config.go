package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DBHost        string `mapstructure:"db_host"`
	DBPort        int    `mapstructure:"db_port"`
	DBUser        string `mapstructure:"db_user"`
	DBPassword    string `mapstructure:"db_password"`
	DBName        string `mapstructure:"db_name"`
	DBChannelName string `mapstructure:"db_channel_name"`
}

func NewConfig() (*Config, error) {
	config := &Config{}
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("db_host", "localhost")
	viper.SetDefault("db_port", 5432)
	viper.SetDefault("db_user", "postgres")
	viper.SetDefault("db_password", "postgres")
	viper.SetDefault("db_name", "postgres")
	viper.SetDefault("db_channel_name", "test")

	_ = viper.ReadInConfig()

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}

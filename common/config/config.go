package config

import (
	"fmt"
	"github.com/spf13/viper"
	libos "imagine/common/os"
	"imagine/utils"
)

type Config[T any] interface {
	ReadConfig() (map[string]T, error)
}

func ReadConfig() (viper.Viper, error) {
	configPath := libos.CurrentWorkingDirectory + "/config"

	v := viper.New() // Create a new viper instance to avoid global state issues
	v.SetConfigName(utils.AppName)
	v.SetConfigType("json")
	v.AddConfigPath(configPath)

	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.Viper{}, fmt.Errorf("can't find config file: %w", err)
		} else {
			return viper.Viper{}, fmt.Errorf("error reading config file: %w", err)
		}
	}
	return *v, nil
}

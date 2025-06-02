package http


import (
	"log/slog"
	"github.com/spf13/viper"

	libos "imagine/common/os"
	"imagine/utils"
	"fmt"
)

type Server struct {
	Port int
	Host string
}

type ImagineServer struct {
	Port   int
	Host   string
	Key    string
	Logger *slog.Logger
}

func (server ImagineServer) ReadConfig() (viper.Viper, error) {
	configPath := libos.CurrentWorkingDirectory + "/config"

	viper.SetConfigName(utils.AppName)
	viper.SetConfigType("json")
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.Viper{}, fmt.Errorf("can't find config file: %w", err)
		}
		// For other errors when reading the config
		return viper.Viper{}, fmt.Errorf("error reading config file: %w", err)
	}

	return *viper.GetViper(), nil
}

package http

***REMOVED***
***REMOVED***
	"log/slog"

***REMOVED***

	"github.com/spf13/viper"

	libos "imagine/common/os"
***REMOVED***

type Server struct {
	Port int
	Host string
***REMOVED***

type ImagineServer struct {
	Port   int
	Host   string
	Key    string
	Logger *slog.Logger
***REMOVED***

func (server ImagineServer***REMOVED*** ReadConfig(***REMOVED*** (viper.Viper, error***REMOVED*** {
	configPath := libos.CurrentWorkingDirectory + "/config"

	viper.SetConfigName(utils.AppName***REMOVED***
	viper.SetConfigType("json"***REMOVED***
	viper.AddConfigPath(configPath***REMOVED***

	err := viper.ReadInConfig(***REMOVED***

***REMOVED***
		if _, ok := err.(viper.ConfigFileNotFoundError***REMOVED***; ok {
			return viper.Viper{***REMOVED***, fmt.Errorf("can't find config file: %w", err***REMOVED***
	***REMOVED*** else {
			return viper.Viper{***REMOVED***, fmt.Errorf("error reading config file: %w", err***REMOVED***
	***REMOVED***
***REMOVED***

	return *viper.GetViper(***REMOVED***, nil
***REMOVED***

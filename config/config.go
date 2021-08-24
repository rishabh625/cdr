package config

import (
	object "cdr/object"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
)

func New(configPath string) (object.Configuration, error) {
	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		viper.AddConfigPath(basepath)
	}
	viper.SetConfigType("json")
	viper.AutomaticEnv()
	if os.Getenv("ENV") == "PROD" {
		viper.SetConfigName("configuration.production")
	} else if os.Getenv("ENV") == "PT" {
		viper.SetConfigName("configuration.pt")
	} else {
		viper.SetConfigName("configuration.development")
	}
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("Error reading config file, %s", err)
	}
	var configuration object.Configuration
	err := viper.Unmarshal(&configuration)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatalf("Error while unmarshaling the configuration")
	}
	return configuration, err
}

package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	PolygonAPIKey string
	PolygonAPIURL string
	Ticker        string
	Interval      int
	Timespan      string
	StartDate     string
	EndDate       string
}

var AppConfig *Config

func InitConfig() {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	AppConfig = &Config{
		PolygonAPIKey: viper.GetString("POLYGON_API_KEY"),
		PolygonAPIURL: viper.GetString("POLYGON_API_URL"),
		Ticker:        viper.GetString("TICKER"),
		Interval:      viper.GetInt("INTERVAL"),
		Timespan:      viper.GetString("TIMESPAN"),
		StartDate:     viper.GetString("START_DATE"),
		EndDate:       viper.GetString("END_DATE"),
	}
}

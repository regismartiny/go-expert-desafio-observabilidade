package configs

import "github.com/spf13/viper"

type conf struct {
	TemperatureAPIToken   string `mapstructure:"TEMPERATURE_API_TOKEN"`
	TemperatureAPIBaseURL string `mapstructure:"TEMPERATURE_API_BASE_URL"`
}

func LoadConfig(path string) (*conf, error) {
	var cfg *conf
	viper.SetConfigName("app_config")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	return cfg, nil
}

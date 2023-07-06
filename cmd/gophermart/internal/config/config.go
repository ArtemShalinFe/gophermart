package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Address string
	Accrual string
	DSN     string
}

func GetConfig() *Config {

	c := &Config{}

	pflag.StringVarP(&c.Address, "address", "a", "", "Gophermart-service address and port")
	pflag.StringVarP(&c.Accrual, "accrual", "r", "", "Accrual-service address and port")
	pflag.StringVarP(&c.DSN, "dsn", "d", "", "Postgresql DSN string")
	pflag.Parse()

	viper.AutomaticEnv()
	viper.SetDefault("RUN_ADDRESS", "localhost:8078")
	viper.SetDefault("DATABASE_URI", "")
	viper.SetDefault("ACCRUAL_SYSTEM_ADDRESS", "localhost:8080")

	if c.Address == "" {
		c.Address = viper.GetString("RUN_ADDRESS")
	}

	if c.Accrual == "" {
		c.Accrual = viper.GetString("ACCRUAL_SYSTEM_ADDRESS")
	}

	if c.DSN == "" {
		c.DSN = viper.GetString("DATABASE_URI")
	}

	return c

}

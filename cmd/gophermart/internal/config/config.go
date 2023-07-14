package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Address string
	Accrual string
	DSN     string
	Key     []byte
}

const envAddress = "RUN_ADDRESS"
const envDSN = "DATABASE_URI"
const envAccrualAddress = "ACCRUAL_SYSTEM_ADDRESS"
const envSecretKey = "KEY"

func GetConfig() *Config {
	c := &Config{}

	var key string
	pflag.StringVarP(&c.Address, "address", "a", "", "Gophermart-service address and port")
	pflag.StringVarP(&c.Accrual, "accrual", "r", "", "Accrual-service address and port")
	pflag.StringVarP(&c.DSN, "dsn", "d", "", "Postgresql DSN string")
	pflag.StringVarP(&key, "key", "k", "", "Secret key")
	pflag.Parse()

	viper.AutomaticEnv()
	viper.SetDefault(envAddress, "localhost:8078")
	viper.SetDefault(envDSN, "")
	viper.SetDefault(envAccrualAddress, "localhost:8080")
	viper.SetDefault(envSecretKey, "gophermart")

	if c.Address == "" {
		c.Address = viper.GetString(envAddress)
	}

	if c.DSN == "" {
		c.DSN = viper.GetString(envDSN)
	}

	if c.Accrual == "" {
		c.Accrual = viper.GetString(envAccrualAddress)
	}

	if key == "" {
		key = viper.GetString(envSecretKey)
	}
	c.Key = []byte(key)

	return c
}

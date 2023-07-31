package config

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	Address         string
	Accrual         string
	AccrualInterval int
	DSN             string
	TokenExp        time.Duration
	Key             []byte
}

const envAddress = "RUN_ADDRESS"
const envDSN = "DATABASE_URI"
const envAccrualAddress = "ACCRUAL_SYSTEM_ADDRESS"
const envSecretKey = "KEY"
const envAccrualInterval = "ACCRUAL_INTERVAL_SECOND"
const envTokenExp = "JWT_TOKEN_EXP"

func GetConfig() *Config {
	c := &Config{}

	var key string
	var tokExp int
	pflag.StringVarP(&c.Address, "address", "a", "", "Gophermart address and port")
	pflag.StringVarP(&c.Accrual, "accrual", "r", "", "Accrual address and port")
	pflag.IntVarP(&c.AccrualInterval, "accrualInterval", "i", 0, "Gophermart using this timeout between requests to the accrual service")
	pflag.StringVarP(&c.DSN, "dsn", "d", "", "Postgresql DSN string")
	pflag.StringVarP(&key, "key", "k", "", "Secret key")
	pflag.IntVarP(&tokExp, "tokenExpiration", "t", 0, "jwt token expiration")
	pflag.Parse()

	viper.AutomaticEnv()
	viper.SetDefault(envAddress, "localhost:8078")
	viper.SetDefault(envDSN, "")
	viper.SetDefault(envAccrualAddress, "localhost:8080")
	viper.SetDefault(envSecretKey, "gophermart")
	viper.SetDefault(envAccrualInterval, 2)
	viper.SetDefault(envTokenExp, 1)

	if c.Address == "" {
		c.Address = viper.GetString(envAddress)
	}

	if c.DSN == "" {
		c.DSN = viper.GetString(envDSN)
	}

	if c.Accrual == "" {
		c.Accrual = viper.GetString(envAccrualAddress)
	}

	if c.AccrualInterval == 0 {
		c.AccrualInterval = viper.GetInt(envAccrualInterval)
	}

	if tokExp == 0 {
		c.TokenExp = time.Hour * time.Duration(viper.GetInt(envTokenExp))
	}

	if key == "" {
		key = viper.GetString(envSecretKey)
	}
	c.Key = []byte(key)

	return c
}

func (c *Config) String() string {
	return fmt.Sprintf("Address: %s, Accrual: %s, AccrualInterval: %d, DSN: %s", c.Address, c.Accrual, c.AccrualInterval, c.DSN)
}

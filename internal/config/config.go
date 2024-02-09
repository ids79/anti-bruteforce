package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger       LoggerConf       `toml:"logger"`
	Database     DBConf           `toml:"database"`
	HTTPServer   HTTP             `toml:"http-server"`
	GRPCServer   GRPC             `toml:"grpc-server"`
	Limits       BruteforceLimits `toml:"bruteforce-limits"`
	TickInterval int              `toml:"tick-interval"`
	ExpireLimit  int              `toml:"expire-limit"`
}

type LoggerConf struct {
	Level       string `toml:"level"`
	LogEncoding string `toml:"log_encoding"`
}

type DBConf struct {
	ConnectString string `toml:"connect_str"`
}

type HTTP struct {
	Address string `toml:"address"`
}

type GRPC struct {
	Address string `toml:"address"`
}

type BruteforceLimits struct {
	N int `toml:"n"`
	M int `toml:"m"`
	K int `toml:"k"`
}

func NewConfig(configFile string) Config {
	var c Config
	tomlFile, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal("error reading the configuration file")
	}
	_, err = toml.Decode(string(tomlFile), &c)
	if err != nil {
		log.Fatal("error decode the configuration file")
	}
	if c.Logger.Level == "" {
		c.Logger.Level = "ERROR"
	}
	if c.Logger.LogEncoding == "" {
		c.Logger.LogEncoding = "console"
	}
	return c
}

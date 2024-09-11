package config

import (
	"encoding/json"
	"flag"
	"os"

	"log"
)

type Config struct {
	DBConnection string `json:"database_dsn"`
	Port         int    `json:"port"`
}

func MustLoad() *Config {
	path := parseConfigPath()

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}

	var cfg Config

	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("error decoding json: %v", err)
	}

	return &cfg
}

func parseConfigPath() string {
	var res string

	if os.Getenv("CONFIG_PATH") != "" {
		res = os.Getenv("CONFIG_PATH")

		return res
	}

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	return res
}

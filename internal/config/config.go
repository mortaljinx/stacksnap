package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DockerSocket    string        `mapstructure:"docker_socket"`
	OutputDirectory string        `mapstructure:"output_directory"`
	ScanInterval    time.Duration `mapstructure:"scan_interval"`
	KeepVersions    int           `mapstructure:"keep_versions"`
	LogLevel        string        `mapstructure:"log_level"`
}

func Load() *Config {
	v := viper.New()

	// Defaults
	v.SetDefault("docker_socket", "/var/run/docker.sock")
	v.SetDefault("output_directory", "/backups")
	v.SetDefault("scan_interval", "5m")
	v.SetDefault("keep_versions", 10)
	v.SetDefault("log_level", "info")

	// Optional config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/config")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err == nil {
		log.Println("Loaded config file:", v.ConfigFileUsed())
	}

	// Environment variables
	v.SetEnvPrefix("STACKSNAP")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config: %v", err)
	}

	return &cfg
}

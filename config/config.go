package config

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	DB struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
		SSLMode  string
	}
	Server struct {
		Host string
		Port int
	}
	JWT struct {
		Secret   string
		Duration time.Duration
	}
}

// DSN generates a connection string for postgres
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DB.User,
		c.DB.Password,
		c.DB.Host,
		c.DB.Port,
		c.DB.Name,
		c.DB.SSLMode,
	)
}

// SafeString returns a string for logging (without passwords)
func (c *Config) SafeString() string {
	return fmt.Sprintf(
		"DB{host=%s port=%d user=%s password=%s name=%s sslmode=%s}\n"+
			"Server{host=%s port=%d}\n"+
			"JWT{secret=%s duration=%s}",
		c.DB.Host, c.DB.Port, c.DB.User, "[hidden]", c.DB.Name, c.DB.SSLMode,
		c.Server.Host, c.Server.Port,
		"[hidden]", c.JWT.Duration,
	)
}

func LoadConfig() (*Config, error) {
	// default values
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.user", "user")
	viper.SetDefault("db.password", "pass")
	viper.SetDefault("db.name", "gophkeeper")
	viper.SetDefault("db.sslmode", "disable")

	viper.SetDefault("server.host", "")
	viper.SetDefault("server.port", 50051)

	viper.SetDefault("jwt.secret", "supersecret")
	viper.SetDefault("jwt.duration", "24h")

	// config file support (config.yaml, config.json, config.toml, etc.)
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// env variables (for example, DB_HOST, SERVER_PORT)
	viper.SetEnvPrefix("GOPHKEEPER")
	viper.AutomaticEnv()

	// command line arguments
	pflag.String("db.host", "", "Database host")
	pflag.Int("db.port", 0, "Database port")
	pflag.String("db.user", "", "Database user")
	pflag.String("db.password", "", "Database password")
	pflag.String("db.name", "", "Database name")
	pflag.String("db.sslmode", "", "Database sslmode")
	pflag.String("server.host", "", "Server host")
	pflag.Int("server.port", 0, "Server port")
	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)

	// reading the file (if any)
	_ = viper.ReadInConfig()

	cfg := &Config{}
	cfg.DB.Host = viper.GetString("db.host")
	cfg.DB.Port = viper.GetInt("db.port")
	cfg.DB.User = viper.GetString("db.user")
	cfg.DB.Password = viper.GetString("db.password")
	cfg.DB.Name = viper.GetString("db.name")
	cfg.DB.SSLMode = viper.GetString("db.sslmode")

	cfg.Server.Host = viper.GetString("server.host")
	cfg.Server.Port = viper.GetInt("server.port")

	cfg.JWT.Secret = viper.GetString("jwt.secret")
	durStr := viper.GetString("jwt.duration")
	dur, err := time.ParseDuration(durStr)
	if err != nil {
		return nil, fmt.Errorf("parse jwt duration error: %w", err)
	}
	cfg.JWT.Duration = dur

	return cfg, nil
}

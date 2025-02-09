package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseDriver string

const (
	DatabaseDriverMySQL    DatabaseDriver = "mysql"
	DatabaseDriverPostgres DatabaseDriver = "postgres"
	DatabaseDriverSQLite   DatabaseDriver = "sqlite"
)

func (d DatabaseDriver) IsValid() bool {
	switch d {
	case DatabaseDriverMySQL, DatabaseDriverPostgres, DatabaseDriverSQLite:
		return true
	}
	return false
}

type Config struct {
	Database struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"database"`

	Discord struct {
		Token string `yaml:"token"`
	} `yaml:"discord"`
}

func New(path string) (*Config, error) {
	cfg, err := loadConfig(path)

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config:", err)
		fmt.Fprintln(os.Stderr, "using default config")

		cfg = &Config{}
	}

	overridden := overrideWithEnv(cfg)
	if overridden {
		fmt.Fprintln(os.Stderr, "some config values have been overridden by environment variables")
	}

	if !DatabaseDriver(cfg.Database.Driver).IsValid() {
		return nil, fmt.Errorf("invalid database driver: %s", cfg.Database.Driver)
	}

	return cfg, nil
}

func (c Config) GormDialector() gorm.Dialector {
	switch DatabaseDriver(c.Database.Driver) {
	case DatabaseDriverMySQL:
		return mysql.Open(c.Database.DSN)
	case DatabaseDriverPostgres:
		return postgres.Open(c.Database.DSN)
	case DatabaseDriverSQLite:
		return sqlite.Open(c.Database.DSN)
	default:
		panic("invalid database driver")
	}
}

func loadConfig(path string) (*Config, error) {
	cfg := &Config{}
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	err = yaml.NewDecoder(r).Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func overrideWithEnv(cfg *Config) bool {
	overridden := false
	if token := os.Getenv("DISCORD_TOKEN"); token != "" {
		cfg.Discord.Token = token
		overridden = true
	}

	if driver := os.Getenv("DATABASE_DRIVER"); driver != "" {
		cfg.Database.Driver = driver
		overridden = true
	}

	if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
		cfg.Database.DSN = dsn
		overridden = true
	}

	return overridden
}

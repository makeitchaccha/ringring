package config

import (
	"fmt"
	"os"

	"github.com/golang/freetype/truetype"
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

type configDTO struct {
	Database struct {
		Driver string `yaml:"driver"`
		DSN    string `yaml:"dsn"`
	} `yaml:"database"`

	Discord struct {
		Token string `yaml:"token"`
		Font  string `yaml:"font"`
	} `yaml:"discord"`
}

type Config struct {
	Dialector gorm.Dialector
	Token     string
	Font      *truetype.Font
}

func New(path string) (*Config, error) {
	raw, err := loadConfig(path)

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config:", err)
		fmt.Fprintln(os.Stderr, "using default config")

		raw = &configDTO{}
	}

	overridden := overrideWithEnv(raw)
	if overridden {
		fmt.Fprintln(os.Stderr, "some config values have been overridden by environment variables")
	}

	if !DatabaseDriver(raw.Database.Driver).IsValid() {
		return nil, fmt.Errorf("invalid database driver: %s", raw.Database.Driver)
	}

	if raw.Discord.Token == "" {
		return nil, fmt.Errorf("discord token is required")
	}

	var font *truetype.Font
	if raw.Discord.Font != "" {
		f, err := LoadFont(raw.Discord.Font)
		if err != nil {
			return nil, fmt.Errorf("failed to load font: %w", err)
		}
		font = f
	}

	cfg := &Config{
		Dialector: getDialector(raw.Database.Driver, raw.Database.DSN),
		Token:     raw.Discord.Token,
		Font:      font,
	}

	return cfg, nil
}

func getDialector(driver, dsn string) gorm.Dialector {
	switch DatabaseDriver(driver) {
	case DatabaseDriverMySQL:
		return mysql.Open(dsn)
	case DatabaseDriverPostgres:
		return postgres.Open(dsn)
	case DatabaseDriverSQLite:
		return sqlite.Open(dsn)
	}
	return nil
}

func loadConfig(path string) (*configDTO, error) {
	cfg := &configDTO{}
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

func overrideWithEnv(cfg *configDTO) bool {
	overridden := overrideString("DISCORD_TOKEN", &cfg.Discord.Token) ||
		overrideString("DATABASE_DRIVER", &cfg.Database.Driver) ||
		overrideString("DATABASE_DSN", &cfg.Database.DSN) ||
		overrideString("DISCORD_FONT", &cfg.Discord.Font)

	return overridden
}

func overrideString(key string, target *string) bool {
	if value, ok := os.LookupEnv(key); ok {
		*target = value
		return true
	}
	return false
}

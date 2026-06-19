package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server ServerConfig `envPrefix:"SERVER_"`
	MySQL  MySQLConfig  `envPrefix:"MYSQL_"`
	Redis  RedisConfig  `envPrefix:"REDIS_"`
	App    AppConfig    `envPrefix:"APP_"`
	JWT    JWTConfig
}

type ServerConfig struct {
	Port          int    `env:"PORT,required"`
	PrivatePort   int    `env:"PRIVATE_PORT,required"`
	ServiceName   string `env:"SERVICE_NAME"         envDefault:"secunda-task-manager"`
	LogLevel      string `env:"LOG_LEVEL"            envDefault:"info"`
	LogForcePlain bool   `env:"LOG_FORCE_PLAIN_TEXT" envDefault:"false"`
}

type MySQLConfig struct {
	Host            string        `env:"HOST,required"`
	Port            int           `env:"PORT"              envDefault:"3306"`
	User            string        `env:"USER,required"`
	Password        string        `env:"PASSWORD,required"`
	DBName          string        `env:"DBNAME,required"`
	MaxOpenConns    int           `env:"MAX_OPEN_CONNS"    envDefault:"25"`
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS"    envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME" envDefault:"5m"`
	ConnMaxIdleTime time.Duration `env:"CONN_MAX_IDLE_TIME" envDefault:"1m"`
}

type RedisConfig struct {
	Host     string `env:"HOST,required"`
	Port     int    `env:"PORT"      envDefault:"6379"`
	Password string `env:"PASSWORD"`
	DB       int    `env:"DB"        envDefault:"0"`
	PoolSize int    `env:"POOL_SIZE" envDefault:"10"`
}

type AppConfig struct {
	AllowedOrigins string `env:"ALLOWED_ORIGINS"`
}

type JWTConfig struct {
	Issuer  string        `env:"TOKEN_ISSUER,required"`
	Leeway  time.Duration `env:"LEEWAY"               envDefault:"5s"`
	Access  TokenConfig   `envPrefix:"ACCESS_TOKEN_"`
	Refresh TokenConfig   `envPrefix:"REFRESH_TOKEN_"`
}

type TokenConfig struct {
	Key string        `env:"KEY,required"`
	TTL time.Duration `env:"TTL,required"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

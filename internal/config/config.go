package config

import (
	"log"
	"log/slog"
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"

	"go-esb-store/internal/model"
)

type Config struct {
	App      App
	ESB      ESB
	Telegram Telegram
}

type App struct {
	Version  string     `env:"APP_VERSION" env-default:"0.0.1"`
	LogLevel slog.Level `env:"APP_LOG_LEVEL" env-default:"info"`
	Mode     model.Mode `env:"APP_MODE" env-default:"prod"`
}

type ESB struct {
	BaseURL       url.URL       `env:"ESB_BASE_URL" required:"true"`
	APIKey        string        `env:"ESB_API_KEY" required:"true"`
	Timeout       time.Duration `env:"ESB_TIMEOUT" envDefault:"60s"`
	LimitPageSize int           `env:"ESB_LIMIT_PAGE_SIZE" envDefault:"100"`
}

type Telegram struct {
	Mode  model.Mode
	Token string `env:"TG_TOKEN" required:"true"`
}

func Must() *Config {
	var config Config

	if err := env.Parse(&config); err != nil {
		log.Fatalln(err)
	}

	config.Telegram.Mode = config.App.Mode

	return &config
}

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
	YDB      YDB
}

type App struct {
	Name     string     `env:"APP_NAME" envDefault:"esb"`
	Version  string     `env:"APP_VERSION" envDefault:"0.0.1"`
	LogLevel slog.Level `env:"APP_LOG_LEVEL" envDefault:"info"`
	Mode     model.Mode `env:"APP_MODE" envDefault:"prod"`
}

type ESB struct {
	BaseURL       url.URL       `env:"ESB_BASE_URL" required:"true"`
	APIKey        string        `env:"ESB_API_KEY" required:"true"`
	Timeout       time.Duration `env:"ESB_TIMEOUT" envDefault:"60s"`
	LimitPageSize int           `env:"ESB_LIMIT_PAGE_SIZE" envDefault:"100"`
}

type Telegram struct {
	Token  string `env:"TG_TOKEN" required:"true"`
	ChatID int64  `env:"TG_CHAT_ID" required:"true"`
}

type YDB struct {
	BaseURL      url.URL           `env:"YDB_BASE_URL" required:"true"`
	Path         string            `env:"YDB_PATH" required:"true"`
	CredsFile    string            `env:"YDB_CREDS_FILE" required:"true"`
	DatabaseName string            `env:"YDB_DATABASE_NAME" required:"true"`
	TablesMap    map[string]string `env:"YDB_TABLES_MAP" required:"true"`
	BatchSize    int               `env:"YDB_BATCH_SIZE" envDefault:"500"`
	Timeout      time.Duration     `env:"YDB_TIMEOUT" envDefault:"60s"`
	Mode         model.Mode
}

func Must() *Config {
	var config Config

	if err := env.Parse(&config); err != nil {
		log.Fatalln(err)
	}

	config.YDB.Mode = config.App.Mode

	return &config
}

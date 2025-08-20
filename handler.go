package main

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"go-esb-store/internal/config"
	"go-esb-store/internal/esb"
	"go-esb-store/internal/model"
	"go-esb-store/pkg/logger"
	"go-esb-store/pkg/trigger"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

func Handler(ctx context.Context, event interface{}) (*Response, error) {
	cfg := config.Must()
	triggerType := trigger.DetectType(event)

	logger.Init(cfg.App.LogLevel)
	if cfg.App.Mode == model.Dev && cfg.App.LogLevel == slog.LevelDebug && triggerType == string(trigger.LocalSource) {
		fmt.Println("RUNNING IN DEVELOPMENT MODE")
		fmt.Printf("config: %+v\n", cfg)
	}
	logger.Info("main.Handler: Starting...", "trigger_type", triggerType)

	esbClient, err := esb.New(&cfg.ESB)
	if err != nil {
		return nil, err
	}
	tgClient, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: 200,
		Body:       "OK",
	}, nil
}

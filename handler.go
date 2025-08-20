package main

import (
	"context"
	"fmt"
	"log/slog"

	"go-esb-store/internal/app"
	"go-esb-store/internal/config"
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

	a, err := app.New(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err = a.Run(ctx); err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: 200,
		Body:       "OK",
	}, nil
}

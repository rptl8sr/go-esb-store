package app

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go-esb-store/internal/config"

	"go-esb-store/internal/esb"
	"go-esb-store/internal/ydb"
)

type App struct {
	esb *esb.ClientWithDefaults
	tg  *tgbotapi.BotAPI
	ydb *ydb.Client
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	esbClient, err := esb.NewESBClient(&cfg.ESB)
	if err != nil {
		return nil, err
	}

	tgClient, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		return nil, err
	}

	ydbClient, err := ydb.NewYDBClient(ctx, &cfg.YDB)
	if err != nil {
		return nil, err
	}

	return &App{
		esb: esbClient,
		tg:  tgClient,
		ydb: ydbClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	rawStores, err := a.esb.GetStores(ctx)
	if err != nil {
		return err
	}
}

package main

import (
	"context"
	"go-esb-store/internal/config"
	"go-esb-store/pkg/trigger"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

func Handler(ctx context.Context, event interface{}) (*Response, error) {
	cfg := config.Must()
	triggerType := trigger.DetectType(event)

}

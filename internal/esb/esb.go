package esb

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go-esb-store/internal/config"
)

type ClientWithDefaults struct {
	*ClientWithResponses
	PageSize int
}

func New(cfg *config.ESB) (*ClientWithDefaults, error) {
	raw, err := newClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ClientWithDefaults{
		ClientWithResponses: raw,
		PageSize:            cfg.LimitPageSize,
	}, nil
}

func newClient(cfg *config.ESB) (*ClientWithResponses, error) {
	return NewClientWithResponses(
		cfg.BaseURL.String(),
		WithHTTPClient(&http.Client{Timeout: cfg.Timeout * time.Second}),
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey))
			return nil
		}),
	)
}

func (c *ClientWithDefaults) applyTop(p *GetStoresParams) *GetStoresParams {
	if p == nil {
		p = &GetStoresParams{}
	}

	if p.Top == nil && c.PageSize > 0 {
		p.Top = &c.PageSize
	}

	return p
}

func (c *ClientWithDefaults) GetStores(ctx context.Context, params *GetStoresParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	params = c.applyTop(params)
	return c.ClientWithResponses.GetStores(ctx, params, reqEditors...)
}

func (c *ClientWithDefaults) GetStoresWithResponse(ctx context.Context, params *GetStoresParams, reqEditors ...RequestEditorFn) (*GetStoresResponse, error) {
	params = c.applyTop(params)
	return c.ClientWithResponses.GetStoresWithResponse(ctx, params, reqEditors...)
}

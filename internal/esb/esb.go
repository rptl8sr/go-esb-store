package esb

import (
	"context"
	"errors"
	"fmt"
	"go-esb-store/internal/utils"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go-esb-store/internal/config"
	"go-esb-store/pkg/logger"
)

type ClientWithDefaults struct {
	*ClientWithResponses
	PageSize int
}

func NewESBClient(cfg *config.ESB) (*ClientWithDefaults, error) {
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

func (c *ClientWithDefaults) GetStores(ctx context.Context) ([]Store, error) {
	pages, err := c.getStoresPagesCount(ctx)
	if err != nil {
		logger.Error("esb.GetStores: error getting stores", "error", err)
		return nil, err
	}
	if pages < 1 {
		logger.Error(fmt.Sprintf("esb.GetStores: %s", ErrNoPageToFetch))
		return nil, ErrNoPageToFetch
	}

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		stores = make([]Store, 0, pages*c.PageSize)
		errCh  = make(chan error, pages)
	)

	for i := 0; i < pages; i++ {
		wg.Add(1)
		page := i
		go func() {
			defer wg.Done()

			storesPage, er := c.getStoresPageData(ctx, page)
			if er != nil {
				errCh <- er
				return
			}
			if len(storesPage) == 0 {
				return
			}

			mu.Lock()
			stores = append(stores, storesPage...)
			mu.Unlock()
		}()
	}

	wg.Wait()
	close(errCh)

	var errs error
	for e := range errCh {
		if e != nil {
			logger.Error("esb.GetStores: error getting stores", "error", e)
			errs = errors.Join(errs, e)
		}
	}
	if errs != nil {
		return nil, errs
	}

	if len(stores) == 0 {
		logger.Error(fmt.Sprintf("esb.GetStores: %s", ErrNoStoresData))
		return nil, ErrNoStoresData
	}

	logger.Info("esb.GetStores: got stores", "count", len(stores), "pages", pages, "limit", c.PageSize)
	return stores, nil
}

func (c *ClientWithDefaults) getStoresPagesCount(ctx context.Context) (int, error) {
	filter := GetStoresCountParamsFilterPrimaryCountryRegionIdEqRUS

	res, err := c.GetStoresCountWithResponse(
		ctx,
		&GetStoresCountParams{
			Filter: &filter,
		},
	)

	if err != nil {
		logger.Error("esb.getStoresPagesCount: error getting store count", "error", err)
		return -1, err
	}

	if res.StatusCode() != http.StatusOK {
		logger.Error("esb.getStoresPagesCount: non-200 response", "status", res.Status())
		return -1, fmt.Errorf("%w: %s", ErrUnexpectedStatus, res.Status())
	}

	cleanedBody := utils.CleanString(string(res.Body))
	count, err := strconv.Atoi(cleanedBody)
	if err != nil {
		logger.Error("esb.getStoresPagesCount: atoi failed", "error", err, "body", cleanedBody)
		return -1, fmt.Errorf("%w: %q", ErrInvalidStoresCount, cleanedBody)
	}

	pages := int(math.Ceil(float64(count) / float64(c.PageSize)))
	logger.Info("esb.getStoresPagesCount: got store count", "count", count, "pages", pages, "limit", c.PageSize)

	return pages, nil
}

func (c *ClientWithDefaults) getStoresPageData(ctx context.Context, page int) ([]Store, error) {
	filter := GetStoresParamsFilterPrimaryCountryRegionIdEqRUS
	skip := page * c.PageSize

	logger.Info("esb.getStoresPageData: getting stores", "page", page, "limit", c.PageSize, "skip", skip, "time", time.Now().String())
	res, err := c.GetStoresWithResponse(
		ctx,
		&GetStoresParams{
			Filter: &filter,
			Skip:   &skip,
			Top:    &c.PageSize,
		},
	)

	if err != nil {
		logger.Error("esb.getStoresPageData: error getting stores page", "error", err, "page", page)
		return nil, err
	}

	if res.StatusCode() != http.StatusOK {
		logger.Error("esb.getStoresPageData: non-200 response", "status", res.Status(), "page", page)
		return nil, fmt.Errorf("%w: %s", ErrUnexpectedStatus, res.Status())
	}

	var stores []Store
	if res.JSON200 != nil && res.JSON200.Value != nil {
		stores = *res.JSON200.Value
	}

	logger.Info("esb.getStoresPageData: got stores", "count", len(stores), "page", page, "limit", c.PageSize, "skip", skip, "time", time.Now().String())
	return stores, nil
}

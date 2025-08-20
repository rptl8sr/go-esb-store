package app

import (
	"context"
	"fmt"
	"strconv"

	"go-esb-store/internal/config"
	"go-esb-store/internal/esb"
	"go-esb-store/internal/model"
	"go-esb-store/internal/utils"
	"go-esb-store/internal/ydb"
	"go-esb-store/pkg/logger"
)

type App struct {
	esb *esb.ClientWithDefaults
	ydb *ydb.Client
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	esbClient, err := esb.NewESBClient(&cfg.ESB)
	if err != nil {
		return nil, err
	}

	ydbClient, err := ydb.NewYDBClient(ctx, &cfg.YDB)
	if err != nil {
		return nil, err
	}

	return &App{
		esb: esbClient,
		ydb: ydbClient,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	rawStores, err := a.esb.GetStores(ctx)
	if err != nil {
		return err
	}

	stores := make([]model.Store, 0, len(rawStores))
	for i, rs := range rawStores {
		s, e := a.rawToModelStore(rs)
		if e != nil {
			logger.Error("app.Run: failed to convert raw store", "error", e, "store", rs, "index", i)
			continue
		}
		stores = append(stores, *s)
	}

	if err = a.ydb.SetStores(ctx, stores); err != nil {
		return err
	}

	return nil
}

func (a *App) rawToModelStore(rawStore esb.Store) (*model.Store, error) {
	store := &model.Store{}

	// Must: Store number
	if rawStore.StoreFactsNumber == nil {
		logger.Error("service.rawToModelStore: invalid store facts number (nil)", "rawStore", rawStore)
		return nil, ErrInvalidStoreFactsNumber
	}
	numStr := utils.CleanString(*rawStore.StoreFactsNumber)
	if numStr == "" {
		logger.Warn("service.rawToModelStore: empty store facts number", "rawStore", rawStore)
		return nil, ErrEmptyStoreFactsNumber
	}
	number, err := strconv.Atoi(numStr)
	if err != nil {
		logger.Error("service.rawToModelStore: error parsing store facts number", "error", err, "value", numStr, "rawStore", rawStore)
		return nil, fmt.Errorf("%w: %q", ErrParseStoreFactsNumber, numStr)
	}
	store.Number = number

	// Must: Name
	if rawStore.NameAlias == nil {
		logger.Error("service.rawToModelStore: invalid store name alias (nil)", "rawStore", rawStore)
		return nil, ErrInvalidStoreName
	}
	if name := utils.CleanString(*rawStore.NameAlias); name == "" {
		logger.Error("service.rawToModelStore: invalid store name alias (empty)", "rawStore", rawStore)
		return nil, ErrInvalidStoreName
	} else {
		store.Name = name
	}

	// Must: Address
	if rawStore.PrimaryAddress == nil {
		logger.Error("service.rawToModelStore: invalid primary address (nil)", "rawStore", rawStore)
		return nil, ErrInvalidStoreAddress
	}
	if addr := utils.CleanString(*rawStore.PrimaryAddress); addr == "" {
		logger.Error("service.rawToModelStore: invalid primary address (empty)", "rawStore", rawStore)
		return nil, ErrInvalidStoreAddress
	} else {
		store.Address = addr
	}

	// Optional: Mall
	if rawStore.FacilityShoppingCenterName != nil {
		store.Mall = utils.CleanString(*rawStore.FacilityShoppingCenterName)
	} else {
		store.Mall = ""
	}

	// Optional: Franchise
	if rawStore.FranchiseePartnerName != nil && *rawStore.FranchiseePartnerName != "" {
		store.Franchise = utils.CleanString(*rawStore.FranchiseePartnerName)
	} else {
		logger.Warn("service.rawToModelStore: empty store franchise", "rawStore", rawStore)
		store.Franchise = ""
	}

	// Optional: Brand
	if rawStore.BrandId != nil && *rawStore.BrandId == "" {
		store.Brand = utils.CleanString(*rawStore.BrandId)
	} else {
		logger.Warn("service.rawToModelStore: empty store brand", "rawStore", rawStore)
		store.Brand = ""
	}

	// Optional: Format
	if rawStore.StoreFormatId != nil && *rawStore.StoreFormatId != "" {
		store.Format = utils.CleanString(*rawStore.StoreFormatId)
	} else {
		logger.Warn("service.rawToModelStore: empty store format", "rawStore", rawStore)
		store.Format = ""
	}

	// Optional: Status
	if rawStore.Status != nil && *rawStore.Status != "" {
		switch *rawStore.Status {
		case esb.Dead:
			store.Status = model.Dead
		case esb.Closed:
			store.Status = model.Closed
		case esb.Refranchised:
			store.Status = model.Refranchised
		case esb.Open:
			store.Status = model.Open
		case esb.New:
			store.Status = model.New
		case esb.PreOpening:
			store.Status = model.PreOpening
		default:
			logger.Warn("service.rawToModelStore: invalid store status", "rawStore", rawStore)
			store.Status = model.Undefined
		}
	}

	return store, nil
}

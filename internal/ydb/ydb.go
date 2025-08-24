package ydb

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	ycdev "github.com/ydb-platform/ydb-go-yc"
	ycprod "github.com/ydb-platform/ydb-go-yc-metadata"

	"go-esb-store/internal/config"
	"go-esb-store/internal/model"
	"go-esb-store/pkg/logger"
)

const (
	defaultBatchSize       = 500
	storesTableNameDefault = "stores"
)

type Client struct {
	driver       *ydb.Driver
	databaseName string
	tablesMap    map[string]string
	batchSize    int
}

func NewYDBClient(ctx context.Context, cfg *config.YDB) (*Client, error) {
	creds, ca, err := initCreds(cfg.Mode, cfg.CredsFile)
	if err != nil {
		return nil, err
	}
	logger.Debug("ydb.NewYDBClient: creds inited", "mode", cfg.Mode)

	driver, err := ydb.Open(
		ctx,
		cfg.BaseURL.JoinPath(cfg.Path).String(),
		ca,
		creds,
		ydb.WithDialTimeout(cfg.Timeout),
	)
	logger.Debug("ydb.NewYDBClient: open driver")

	if err != nil {
		return nil, fmt.Errorf("failed to connect to YDB: %s", err)
	}

	c := &Client{
		driver:       driver,
		databaseName: cfg.DatabaseName,
		tablesMap:    cfg.TablesMap,
		batchSize:    cfg.BatchSize,
	}

	if cfg.Mode == model.Dev {
		if err = c.initTables(ctx); err != nil {
			return nil, err
		}
		logger.Debug("app.New: table created", "table", c.databaseName)
	}

	return c, nil
}

func (c *Client) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}

func (c *Client) SetStores(ctx context.Context, stores []model.Store) error {
	if len(stores) == 0 {
		return nil
	}

	batchSize := c.batchSize
	if batchSize < 1 {
		batchSize = defaultBatchSize
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for i := 0; i < len(stores); i += batchSize {
		end := i + batchSize
		if end > len(stores) {
			end = len(stores)
		}
		batch := stores[i:end]

		wg.Add(1)
		go func(b []model.Store) {
			defer wg.Done()

			if err := c.setStores(ctx, b); err != nil {
				select {
				case errCh <- err:
					cancel()
				default:
				}
			}
		}(batch)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func (c *Client) setStores(ctx context.Context, stores []model.Store) error {
	if len(stores) == 0 {
		return nil
	}

	var tableName string
	if v, ok := c.tablesMap[storesTableNameDefault]; !ok {
		tableName = storesTableNameDefault
	} else {
		tableName = v
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("upsert into %s (number, name, address, mall, franchise, brand, format, status) values\n", tableName))

	for i, s := range stores {
		fmt.Fprintf(&b,
			"(%d,%s,%s,%s,%s,%s,%s,%s)",
			s.Number,
			quoteYQL(s.Name),
			quoteYQL(s.Address),
			quoteYQL(s.Mall),
			quoteYQL(s.Franchise),
			quoteYQL(s.Brand),
			quoteYQL(s.Format),
			quoteYQL(string(s.Status)),
		)

		if i < len(stores)-1 {
			b.WriteString(",\n")
		}
	}
	b.WriteString(";")

	err := c.exec(ctx, b.String(), nil)
	if err != nil {
		logger.Error("ydb.SetStores: failed to store stores", "error", err)
		return err
	}

	return nil
}

func (c *Client) exec(ctx context.Context, query string, params *table.QueryParameters) error {
	return c.driver.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		_, _, err := s.Execute(ctx, table.DefaultTxControl(), query, params)
		return err
	})
}

func (c *Client) execScheme(ctx context.Context, query string) error {
	return c.driver.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
		return s.ExecuteSchemeQuery(ctx, query)
	})
}

func (c *Client) initTables(ctx context.Context) error {
	var tableName string
	if v, ok := c.tablesMap[storesTableNameDefault]; !ok {
		tableName = storesTableNameDefault
	} else {
		tableName = v
	}

	query := fmt.Sprintf(`create table if not exists %s (
	    number Int64,
	    name Utf8,
	    address Utf8,
	    mall Utf8,
	    franchise Utf8,
	    brand Utf8,
	    format Utf8,
	    status Utf8,
	    temporary_closed Bool,
	    primary key (number),
	    index idx_stores_name global on (name)
	);`, tableName)

	if err := c.execScheme(ctx, query); err != nil {
		logger.Error("ydb.initTables: failed to init tables", "error", err)
		return err
	}

	return nil
}

func initCreds(mode model.Mode, p string) (ydb.Option, ydb.Option, error) {
	var creds ydb.Option
	var ca ydb.Option

	switch mode {
	case model.Prod:
		ca = ycprod.WithInternalCA()
		creds = ycprod.WithCredentials()
	case model.Dev:
		ca = ycdev.WithInternalCA()
		if p == "" {
			return nil, nil, fmt.Errorf("creds file is required in dev mode")
		}
		creds = ycdev.WithServiceAccountKeyFileCredentials(p)
	default:
		return nil, nil, fmt.Errorf("mode '%s' not supported", mode)
	}

	return creds, ca, nil
}

func quoteYQL(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString("\\\\")
		case '"':
			b.WriteString("\\\"")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			if r < 0x20 {
				// Управляющие символы — в \xHH
				fmt.Fprintf(&b, "\\x%02X", r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

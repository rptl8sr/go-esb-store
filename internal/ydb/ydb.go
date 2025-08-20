package ydb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	ycdev "github.com/ydb-platform/ydb-go-yc"
	ycprod "github.com/ydb-platform/ydb-go-yc-metadata"

	"go-esb-store/internal/config"
	"go-esb-store/internal/model"
	"go-esb-store/pkg/logger"
)

const (
	storesTableNameDefault = "stores"
)

type Client struct {
	driver       *ydb.Driver
	databaseName string
	tablesMap    map[string]string
}

func NewYDBClient(ctx context.Context, cfg *config.YDB) (*Client, error) {
	creds, ca, err := initCreds(cfg.Mode, cfg.CredsFile)

	driver, err := ydb.Open(
		ctx,
		cfg.BaseURL.JoinPath(cfg.Path).String(),
		ca,
		creds,
		ydb.WithDialTimeout(time.Duration(cfg.Timeout)),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to YDB: %s", err)
	}

	c := &Client{
		driver:       driver,
		databaseName: cfg.DatabaseName,
		tablesMap:    cfg.TablesMap,
	}

	if cfg.Mode == model.Dev {
		if err = c.initTables(ctx); err != nil {
			return nil, err
		}
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

	var tableName string
	if v, ok := c.tablesMap[storesTableNameDefault]; !ok {
		tableName = storesTableNameDefault
	} else {
		tableName = v
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("upsert into %s (number, name, address, mall, company, brand, format, status, temporary_closed) values\n", tableName))

	for i, s := range stores {
		fmt.Fprintf(&b,
			"(%d,%s,%s,%s,%s,%s,%s,%s,%t)",
			s.Number,
			quoteYQL(s.Name),
			quoteYQL(s.Address),
			quoteYQL(s.Mall),
			quoteYQL(s.Company),
			quoteYQL(s.Brand),
			quoteYQL(s.Format),
			string(s.Status),
			s.TemporaryClosed,
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

func (c *Client) initTables(ctx context.Context) error {
	var tableName string
	if v, ok := c.tablesMap[storesTableNameDefault]; !ok {
		tableName = storesTableNameDefault
	} else {
		tableName = v
	}

	query := fmt.Sprintf(`
	create table if not exists stores (
	    number Int64,
	    name Utf8,
	    address Utf8,
	    mall Utf8,
	    company Utf8,
	    brand Utf8,
	    format Utf8,
	    status Utf8,
	    temporary_closed Bool,
	    primary key (number),
	    index idx_stores_name global on (name) cover (number)
	);`, tableName)

	if err := c.exec(ctx, query, nil); err != nil {
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

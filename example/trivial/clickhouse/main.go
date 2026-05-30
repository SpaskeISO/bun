package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/chdialect"
	"github.com/uptrace/bun/dialect/feature"
	"github.com/uptrace/bun/extra/bundebug"
)

type Event struct {
	bun.BaseModel `bun:"table:events,alias:e"`

	ID        int64     `bun:"id,pk"`
	Name      string    `bun:"name"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull"`
}

func main() {
	ctx := context.Background()

	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "password",
		},
	})
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		panic(fmt.Errorf("ping clickhouse: %w", err))
	}

	db := bun.NewDB(conn, chdialect.New()).
		WithQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	var version string
	if err := conn.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		panic(err)
	}
	fmt.Println("ClickHouse version:", version)
	fmt.Printf("Dialect: %s\n", db.Dialect().Name())
	fmt.Printf("CTE enabled: %v\n", db.HasFeature(feature.CTE))
	fmt.Printf("AutoIncrement enabled: %v\n", db.HasFeature(feature.AutoIncrement))

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS events (
			id Int64,
			name String,
			created_at DateTime
		) ENGINE = MergeTree()
		ORDER BY id
	`); err != nil {
		panic(fmt.Errorf("create table: %w", err))
	}

	now := time.Now().UTC().Truncate(time.Second)
	ev := &Event{
		ID:        time.Now().UnixNano(),
		Name:      "bun-chdialect-smoke-test",
		CreatedAt: now,
	}

	if _, err := db.NewInsert().Model(ev).Exec(ctx); err != nil {
		panic(fmt.Errorf("insert: %w", err))
	}

	var rows []Event
	if err := db.NewSelect().
		Model(&rows).
		Where("name = ?", ev.Name).
		OrderExpr("id DESC").
		Limit(5).
		Scan(ctx); err != nil {
		panic(fmt.Errorf("select: %w", err))
	}

	fmt.Printf("inserted and read back %d row(s): %+v\n", len(rows), rows)
}

package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Init_db() error {
	conn_str := os.Getenv("DATABASE_URL")
	config, err := pgxpool.ParseConfig(conn_str)
	if err != nil {
		return err
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err := DB.Ping(context.Background()); err != nil {
		return fmt.Errorf("error pinging db: %w", err)
	}

	fmt.Println("✅ Database connected successfully")

	return nil
}

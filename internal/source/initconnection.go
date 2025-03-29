package source

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// postgres://jack:secret@pg.example.com:5432/mydb?sslmode=verify-ca&pool_max_conns=10&pool_max_conn_lifetime=1h30m
func InitDB(ctx context.Context, pathToEnv string) (*pgxpool.Pool, error) {
	if err := godotenv.Load(pathToEnv); err != nil {
		return nil, fmt.Errorf("source: .env error - %v", err)
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	config.MaxConns = int32(5)
	config.MinConns = int32(1)
	config.MaxConnLifetime = 2 * time.Hour
	config.MaxConnIdleTime = 10 * time.Minute

	config.HealthCheckPeriod = 1 * time.Minute

	config.ConnConfig.ConnectTimeout = 1 * time.Second
	config.ConnConfig.DialFunc = (&net.Dialer{
		KeepAlive: config.HealthCheckPeriod,
		Timeout:   config.ConnConfig.ConnectTimeout,
	}).DialContext

	return pgxpool.NewWithConfig(ctx, config)
}

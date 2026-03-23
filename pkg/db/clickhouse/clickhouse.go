// Package clickhouse provides a framework-level ClickHouse connection helper
// following the Optional-init pattern established by pkg/broker/kafka.
//
// Usage:
//
//	conn := clickhouse.NewConnOptional(cfg, logger)
//	if conn == nil {
//	    // ClickHouse not configured — handle gracefully
//	}
//	defer conn.Close()
package clickhouse

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	conf "github.com/Servora-Kit/servora/api/gen/go/servora/conf/v1"
	"github.com/Servora-Kit/servora/pkg/logger"
)

// NewConnOptional opens a ClickHouse connection from the Data config.
// Returns nil when ClickHouse is not configured or connection fails; errors are
// logged internally so callers only need a nil check — consistent with the
// Optional-init pattern used by pkg/broker/kafka.NewBrokerOptional and
// pkg/openfga.NewClientOptional.
//
// The caller is responsible for closing the connection via conn.Close().
func NewConnOptional(cfg *conf.Data, l logger.Logger) driver.Conn {
	log := logger.For(l, "clickhouse/db/pkg")

	if cfg == nil || cfg.Clickhouse == nil || len(cfg.Clickhouse.Addrs) == 0 {
		log.Info("ClickHouse not configured")
		return nil
	}

	chCfg := cfg.Clickhouse

	dialTimeout := 10 * time.Second
	if chCfg.DialTimeout != nil {
		dialTimeout = chCfg.DialTimeout.AsDuration()
	}
	readTimeout := 30 * time.Second
	if chCfg.ReadTimeout != nil {
		readTimeout = chCfg.ReadTimeout.AsDuration()
	}
	maxOpenConns := 10
	if chCfg.MaxOpenConns > 0 {
		maxOpenConns = int(chCfg.MaxOpenConns)
	}
	maxIdleConns := 5
	if chCfg.MaxIdleConns > 0 {
		maxIdleConns = int(chCfg.MaxIdleConns)
	}
	connMaxLifetime := 5 * time.Minute
	if chCfg.ConnMaxLifetime != nil {
		connMaxLifetime = chCfg.ConnMaxLifetime.AsDuration()
	}

	opts := &clickhouse.Options{
		Addr: chCfg.Addrs,
		Auth: clickhouse.Auth{
			Database: chCfg.Database,
			Username: chCfg.Username,
			Password: chCfg.Password,
		},
		DialTimeout:      dialTimeout,
		ReadTimeout:      readTimeout,
		MaxOpenConns:     maxOpenConns,
		MaxIdleConns:     maxIdleConns,
		ConnMaxLifetime:  connMaxLifetime,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	}

	if chCfg.Tls {
		opts.TLS = &tls.Config{InsecureSkipVerify: chCfg.TlsSkipVerify} //nolint:gosec
	}

	switch chCfg.Compress {
	case "lz4":
		opts.Compression = &clickhouse.Compression{Method: clickhouse.CompressionLZ4}
	case "zstd":
		opts.Compression = &clickhouse.Compression{Method: clickhouse.CompressionZSTD}
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		log.Warnf("failed to open ClickHouse connection: %v", err)
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	if err := conn.Ping(ctx); err != nil {
		log.Warnf("failed to ping ClickHouse: %v", err)
		_ = conn.Close()
		return nil
	}

	log.Info("ClickHouse connected")
	return conn
}

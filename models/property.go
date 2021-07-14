package models

import (
	"context"
	"github.com/MixinNetwork/supergroup/config"
	"time"

	"github.com/MixinNetwork/supergroup/session"
)

const properties_DDL = `
CREATE TABLE IF NOT EXISTS properties (
	key         VARCHAR(512) PRIMARY KEY,
	value       VARCHAR(8192) NOT NULL,
	updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

const (
	MainnetSnapshotsCheckpoint = "service-mainnet-snapshots-checkpoint"
)

type Property struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

func ReadProperty(ctx context.Context, key string) (string, error) {
	var val string
	query := "SELECT value FROM properties WHERE key=$1"
	err := session.Database(ctx).QueryRow(ctx, query, key).Scan(&val)
	return val, err
}

func WriteProperty(ctx context.Context, key, value string) error {
	query := "INSERT INTO properties (key,value,updated_at) VALUES($1,$2,$3) ON CONFLICT (key) DO UPDATE SET (value,updated_at)=(EXCLUDED.value, EXCLUDED.updated_at)"
	return session.Database(ctx).ConnExec(ctx, query, key, value, time.Now())
}

func CleanModelCache() {
	cacheAssets = make(map[string]Asset)
	cacheClient = make(map[string]Client)
	cacheManagerMap = make(map[string][]string)
	cacheBlockClientUserIDMap = make(map[string]map[string]bool)
	cacheAllClient = make([]ClientInfo, 0)
}

func cleanCache() {
	for {
		time.Sleep(config.CacheTime)
		CleanModelCache()
	}
}

func init() {
	go cleanCache()
}

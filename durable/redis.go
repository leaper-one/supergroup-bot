package durable

import (
	"context"
	"fmt"
	"os"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/go-redis/redis/v8"
)

type Redis struct {
	*redis.Client
}

func GetRedisConversationStatus(clientID string) string {
	// 0 正常模式 1 禁言模式 2 图文直播模式
	return fmt.Sprintf("client-conversation-%s", clientID)
}

func GetRedisNewMemberNotice(clientID string) string {
	// 1 开启 0 关闭
	return fmt.Sprintf("client-new-member-%s", clientID)
}

func NewRedis(ctx context.Context) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Config.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	err := rdb.Set(ctx, "test", "ok", 0).Err()
	if err != nil {
		os.Exit(1)
	}
	val, err := rdb.Get(ctx, "test").Result()
	if err != nil {
		os.Exit(1)
	}
	if val != "ok" {
		os.Exit(1)
	}
	return &Redis{rdb}
}

func (r *Redis) QGet(ctx context.Context, key string) string {
	val, err := r.Get(ctx, key).Result()
	if err == redis.Nil {
		return ""
	}
	return val
}

func (r *Redis) QSet(ctx context.Context, key, val string) error {
	return r.Set(ctx, key, val, redis.KeepTTL).Err()
}

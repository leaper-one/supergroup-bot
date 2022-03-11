package durable

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/go-redis/redis/v8"
)

type Redis struct {
	*redis.Client
}

func NewRedis(ctx context.Context) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Config.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	err := rdb.Set(ctx, "test", "ok", -1).Err()
	if err != nil {
		log.Println("redis error1...", err)
		os.Exit(1)
	}
	val, err := rdb.Get(ctx, "test").Result()
	if err != nil {
		log.Println("redis error2...", err)
		os.Exit(1)
	}
	if val != "ok" {
		log.Println("redis error3...", err)
		os.Exit(1)
	}
	return &Redis{rdb}
}

func (r *Redis) QPublish(ctx context.Context, channel, clientID string) error {
	return r.Publish(ctx, channel, clientID).Err()
}

func (r *Redis) QSubscribe(ctx context.Context, channel string) *redis.PubSub {
	return r.Subscribe(ctx, channel)
}

func (r *Redis) StructScan(ctx context.Context, key string, res interface{}) error {
	test, err := r.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(test), res)
}

func (r *Redis) StructSet(ctx context.Context, key string, req interface{}) error {
	tByte, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, string(tByte), 15*time.Minute).Err()
}

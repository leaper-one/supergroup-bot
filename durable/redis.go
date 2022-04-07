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

func (r *Redis) QKeys(ctx context.Context, p string) ([]string, error) {
	cursor := uint64(0)
	var keys []string
	var err error
	isStart := false
	_res := make(map[string]bool)
	for {
		if isStart && cursor == 0 {
			break
		}
		if !isStart {
			isStart = true
		}
		keys, cursor, err = r.Scan(ctx, cursor, p, 5000).Result()
		if err != nil {
			return nil, err
		}
		for _, v := range keys {
			_res[v] = true
		}
	}

	res := make([]string, 0, len(_res))
	for k := range _res {
		res = append(res, k)
	}
	return res, nil
}

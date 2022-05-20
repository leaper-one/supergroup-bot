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
	R *redis.Client
	W *redis.Client
}

func NewRedis(ctx context.Context) *Redis {
	w := redis.NewClient(&redis.Options{
		Addr:     config.Config.RedisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	var r *redis.Client
	if config.Config.RedisAddrReplica != "" {
		r = redis.NewClient(&redis.Options{
			Addr:     config.Config.RedisAddrReplica,
			Password: "", // no password set
			DB:       0,  // use default DB
		})
	} else {
		r = w
	}
	err := w.Set(ctx, "test", "ok", -1).Err()
	if err != nil {
		log.Println("redis error1...", err)
		os.Exit(1)
	}
	val, err := r.Get(ctx, "test").Result()
	if err != nil {
		log.Println("redis error2...", err)
		os.Exit(1)
	}
	if val != "ok" {
		log.Println("redis error3...", err)
		os.Exit(1)
	}
	return &Redis{W: w, R: r}
}

func (r *Redis) StructScan(ctx context.Context, key string, res interface{}) error {
	test, err := r.R.Get(ctx, key).Result()
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
	return r.W.Set(ctx, key, string(tByte), 15*time.Minute).Err()
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
		keys, cursor, err = r.R.Scan(ctx, cursor, p, 1000).Result()
		if err != nil {
			return nil, err
		}
		for _, v := range keys {
			_res[v] = true
		}
		time.Sleep(time.Millisecond * 10)
	}

	res := make([]string, 0, len(_res))
	for k := range _res {
		res = append(res, k)
	}
	return res, nil
}

func (r *Redis) QDel(ctx context.Context, keys ...string) error {
	return r.W.Unlink(ctx, keys...).Err()
}

func (r *Redis) QGet(ctx context.Context, k string) *redis.StringCmd {
	return r.R.Get(ctx, k)
}

func (r *Redis) SyncGet(ctx context.Context, k string) *redis.StringCmd {
	return r.W.Get(ctx, k)
}

func (r *Redis) QSet(ctx context.Context, k string, v interface{}, expiration time.Duration) error {
	return r.W.Set(ctx, k, v, expiration).Err()
}

func (r *Redis) QPipelined(ctx context.Context, f func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return r.W.Pipelined(ctx, f)
}

func (r *Redis) QSMembers(ctx context.Context, p string) ([]string, error) {
	return r.R.SMembers(ctx, p).Result()
}

func (r *Redis) QZRange(ctx context.Context, p string, start, stop int64) ([]string, error) {
	return r.W.ZRange(ctx, p, start, stop).Result()
}

func (r *Redis) QZRangeByScore(ctx context.Context, p string, opt *redis.ZRangeBy) ([]string, error) {
	return r.W.ZRangeByScore(ctx, p, opt).Result()
}

func (r *Redis) QIncr(ctx context.Context, p string, d time.Duration) (int64, error) {
	res, err := r.W.Incr(ctx, p).Result()
	if err != nil {
		return 0, err
	}
	if d != 0 {
		r.W.PExpire(ctx, p, d)
	}
	return res, nil
}

func (r *Redis) QPublish(ctx context.Context, p string, v string) error {
	return r.W.Publish(ctx, p, v).Err()
}

func (r *Redis) QSubscribe(ctx context.Context, p string) *redis.PubSub {
	return r.W.Subscribe(ctx, p)
}

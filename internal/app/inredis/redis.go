package inredis

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/ids79/anti-bruteforce/internal/app"
	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
)

type backets struct {
	logg        logger.Logg
	freq        map[app.BacketType]int
	expireLimit time.Duration
	rdb         *redis.Client
	mu          sync.Mutex
}

func NewBackets(logg logger.Logg, conf *config.Config) app.WorkWithBackets {
	freq := make(map[app.BacketType]int)
	freq[app.LOGIN] = conf.Limits.TryForLogin
	freq[app.PASSWORD] = conf.Limits.TryForPass
	freq[app.IP] = conf.Limits.TreyForIP
	b := &backets{}
	b.expireLimit = time.Second * time.Duration(conf.ExpireLimit)
	b.freq = freq
	b.logg = logg
	b.rdb = redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password,
		DB:       0,
	})
	err := b.rdb.Ping().Err()
	if err != nil {
		logg.Error(err)
		return nil
	}
	return b
}

func (b *backets) Connect() error {
	err := b.rdb.Ping().Err()
	if err != nil {
		return err
	}
	return nil
}

func (b *backets) AccessVerification(key string, t app.BacketType) (bool, error) {
	if key == "" {
		return false, errors.New(app.BadRequestStr)
	}
	key = string(t) + key
	tm := time.Now().String()
	keyTime := "{" + key + "}" + tm
	b.mu.Lock()
	err := b.rdb.Set(keyTime, "1", b.expireLimit).Err()
	if err != nil {
		return false, err
	}
	keySlot, err := b.rdb.ClusterKeySlot(key).Result()
	if err != nil {
		return false, err
	}
	count, err := b.rdb.ClusterCountKeysInSlot(int(keySlot)).Result()
	if err != nil {
		return false, err
	}
	b.mu.Unlock()
	return int(count) < b.freq[t], nil
}

func (b *backets) ResetBacket(ctx context.Context, key string, t app.BacketType) error {
	select {
	case <-ctx.Done():
		return app.ErrContextWasExpire
	default:
		key = string(t) + key
		keySlot, err := b.rdb.ClusterKeySlot(key).Result()
		if err != nil {
			return err
		}
		keys, err := b.rdb.ClusterGetKeysInSlot(int(keySlot), b.freq[t]).Result()
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			return app.ErrBacketNotFound
		}
		b.rdb.Del(keys...).Err()
		if err != nil {
			return err
		}
		return nil
	}
}

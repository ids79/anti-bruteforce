package inmem

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ids79/anti-bruteforce/internal/app"
	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
)

const (
	LOGIN    app.BacketType = "LOGIN"
	PASSWORD app.BacketType = "PASS"
	IP       app.BacketType = "IP"
)

type backet struct {
	freq     int
	attempts []time.Time
}

type backets struct {
	logg        logger.Logg
	freq        map[app.BacketType]int
	expireLimit time.Duration
	mu          sync.Mutex
	backets     map[string]*backet
}

func NewBackets(ctx context.Context, logg logger.Logg, conf *config.Config) app.WorkWithBackets {
	freq := make(map[app.BacketType]int)
	freq[LOGIN] = conf.Limits.TryForLogin
	freq[PASSWORD] = conf.Limits.TryForPass
	freq[IP] = conf.Limits.TreyForIP
	b := &backets{}
	b.expireLimit = time.Second * time.Duration(conf.ExpireLimit)
	b.freq = freq
	b.logg = logg
	b.backets = make(map[string]*backet)
	go b.DelAttemptsAndBackets(ctx, time.Millisecond*time.Duration(conf.TickInterval))
	return b
}

func (b *backets) AccessVerification(key string, t app.BacketType) (bool, error) {
	if key == "" {
		return false, errors.New(app.BadRequestStr)
	}
	key = string(t) + key
	b.mu.Lock()
	defer b.mu.Unlock()
	bac, ok := b.backets[key]
	if !ok {
		att := []time.Time{time.Now()}
		bac = &backet{freq: b.freq[t], attempts: att}
		b.backets[key] = bac
	} else {
		bac.attempts = append(bac.attempts, time.Now())
	}
	return len(bac.attempts) < bac.freq, nil
}

func (b *backets) ResetBacket(ctx context.Context, key string, t app.BacketType) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	select {
	case <-ctx.Done():
		return app.ErrContextWasExpire
	default:
		key = string(t) + key
		if val, ok := b.backets[key]; ok {
			val.attempts = nil
			return nil
		}
		return app.ErrBacketNotFound
	}
}

func (b *backets) delAtt(v *backet) {
	l := len(v.attempts)
	for i := l - 1; i >= 0; i-- {
		if time.Since(v.attempts[i]) >= b.expireLimit {
			if i == l-1 {
				v.attempts = nil
			} else {
				v.attempts = v.attempts[i+1:]
			}
			break
		}
	}
}

func (b *backets) DelAttemptsAndBackets(ctx context.Context, tic time.Duration) {
	ticker := time.NewTicker(tic)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.mu.Lock()
			for k, v := range b.backets {
				b.delAtt(v)
				if len(v.attempts) == 0 {
					delete(b.backets, k)
				}
			}
			b.mu.Unlock()
		}
	}
}

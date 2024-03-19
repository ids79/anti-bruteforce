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
	mu       sync.RWMutex
	attempts []time.Time
}

func (b *backet) successfulAttempt() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attempts = append(b.attempts, time.Now())
	return len(b.attempts) < b.freq
}

func (b *backet) resetBacket() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.attempts = nil
}

type backets struct {
	logg        logger.Logg
	freq        map[app.BacketType]int
	expireLimit time.Duration
	backets     map[string]*backet
	mu          sync.RWMutex
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
	go b.delAttemptsAndBackets(ctx, time.Millisecond*time.Duration(conf.TickInterval))
	return b
}

func (b *backets) getValue(key string) (*backet, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	val, ok := b.backets[key]
	return val, ok
}

func (b *backets) delValue(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.backets, key)
}

func (b *backets) AccessVerification(key string, t app.BacketType) (bool, error) {
	if key == "" {
		return false, errors.New(app.BadRequestStr)
	}
	key = string(t) + key
	b.mu.Lock()
	bac, ok := b.backets[key]
	if !ok {
		att := []time.Time{time.Now()}
		b.backets[key] = &backet{freq: b.freq[t], attempts: att}
		b.mu.Unlock()
		return true, nil
	}
	b.mu.Unlock()
	return bac.successfulAttempt(), nil
}

func (b *backets) ResetBacket(ctx context.Context, key string, t app.BacketType) error {
	select {
	case <-ctx.Done():
		return app.ErrContextWasExpire
	default:
		key = string(t) + key
		if val, ok := b.getValue(key); ok {
			val.resetBacket()
			return nil
		}
		return app.ErrBacketNotFound
	}
}

func (b *backets) delAtt(v *backet) {
	v.mu.Lock()
	defer v.mu.Unlock()
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

func (b *backets) delAttemptsAndBackets(ctx context.Context, tic time.Duration) {
	ticker := time.NewTicker(tic)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.mu.RLock()
			bacCopy := make(map[string]*backet)
			for k, v := range b.backets {
				bacCopy[k] = v
			}
			b.mu.RUnlock()
			for k, v := range bacCopy {
				b.delAtt(v)
				if len(v.attempts) == 0 {
					b.delValue(k)
				}
			}
		}
	}
}

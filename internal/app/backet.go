package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ids79/anti-bruteforcer/internal/config"
	"github.com/ids79/anti-bruteforcer/internal/logger"
)

var ErrBacketNotFound = errors.New("the backet is not found")

type backetType string

const (
	LOGIN    backetType = "LOGIN"
	PASSWORD backetType = "PASS"
	IP       backetType = "IP"
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

type WorkWithBackets interface {
	AccessVerification(key string, t backetType) (bool, error)
	ResetBacket(key string, t backetType) error
}

type Backets struct {
	logg        logger.Logg
	freq        map[backetType]int
	expireLimit int
	mu          sync.RWMutex
	backets     map[string]*backet
}

func NewBackets(ctx context.Context, logg logger.Logg, conf *config.Config) *Backets {
	freq := make(map[backetType]int)
	freq[LOGIN] = conf.Limits.N
	freq[PASSWORD] = conf.Limits.M
	freq[IP] = conf.Limits.K
	b := &Backets{}
	b.expireLimit = conf.ExpireLimit
	b.freq = freq
	b.backets = make(map[string]*backet)
	b.logg = logg
	go b.delAttemptsAndBackets(ctx, conf.TickInterval)
	return b
}

func (b *Backets) getValue(key string) (*backet, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	val, ok := b.backets[key]
	return val, ok
}

func (b *Backets) setValue(key string, val *backet) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.backets[key] = val
}

func (b *Backets) delValue(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.backets, key)
}

func (b *Backets) AccessVerification(key string, t backetType) (bool, error) {
	if key == "" {
		b.logg.Error(t, ": parametr was not faund")
		return false, fmt.Errorf("%w: parametr %s or mask not found", ErrBadRequest, t)
	}
	key = string(t) + key
	bac, ok := b.getValue(key)
	if !ok {
		att := []time.Time{time.Now()}
		b.setValue(key, &backet{freq: b.freq[t], attempts: att})
		return true, nil
	}
	return bac.successfulAttempt(), nil
}

func (b *Backets) ResetBacket(key string, t backetType) error {
	key = string(t) + key
	if val, ok := b.getValue(key); ok {
		val.resetBacket()
		return nil
	}
	return ErrBacketNotFound
}

func (b *Backets) delAtt(v *backet) {
	v.mu.Lock()
	defer v.mu.Unlock()
	l := len(v.attempts)
	for i := l - 1; i >= 0; i-- {
		if time.Since(v.attempts[i]) >= time.Second*time.Duration(b.expireLimit) {
			if i == l-1 {
				v.attempts = nil
			} else {
				v.attempts = v.attempts[i+1:]
			}
			break
		}
	}
}

func (b *Backets) delAttemptsAndBackets(ctx context.Context, ticDur int) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(ticDur))
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

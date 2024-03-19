package app

import (
	"context"
	"errors"
)

var ErrBacketNotFound = errors.New("the backet was not found")

var ErrContextWasExpire = errors.New("the context was expire")

type BacketType string

const (
	LOGIN    BacketType = "LOGIN"
	PASSWORD BacketType = "PASS"
	IP       BacketType = "IP"
)

type WorkWithBackets interface {
	AccessVerification(key string, t BacketType) (bool, error)
	ResetBacket(ctx context.Context, key string, t BacketType) error
}

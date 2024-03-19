package app

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
	"github.com/ids79/anti-bruteforce/internal/storage"
)

var BadRequestStr = "bad request"

var ErrIncorrectIP = errors.New("incorrect IP address")

type WorkWithIPList interface {
	AddWhiteList(ctx context.Context, ipStr string) error
	DelWhiteList(ctx context.Context, ipStr string) error
	AddBlackList(ctx context.Context, ipStr string) error
	DelBlackList(ctx context.Context, ipStr string) error
	IsInWhiteList(ctx context.Context, ipStr string) (bool, error)
	IsInBlackList(ctx context.Context, ipStr string) (bool, error)
	GetList(ctx context.Context, t string) []string
}

type IPList struct {
	logg    logger.Logg
	storage storage.Storage
	conf    *config.Config
}

func NewIPList(storage storage.Storage, logg logger.Logg, conf *config.Config) *IPList {
	return &IPList{
		logg:    logg,
		storage: storage,
		conf:    conf,
	}
}

type IPItem struct {
	IP     string
	Mask   string
	IPfrom string
	IPto   string
}

func (ip IPItem) String() string {
	return fmt.Sprintf("ip: %s, mask: %s, ipfrom: %s, ipto: %s", ip.IP, ip.Mask, ip.IPfrom, ip.IPto)
}

func checkParam(ipStr string) error {
	if ipStr == "" {
		return errors.New("ip parametr was not faund")
	}
	return nil
}

func (l *IPList) AddWhiteList(ctx context.Context, ipStr string) error {
	if err := checkParam(ipStr); err != nil {
		return fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	ipNet, err := netip.ParsePrefix(ipStr)
	if err != nil {
		return fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	if err := l.storage.AddWhiteList(ctx, ipNet.Masked()); err != nil {
		return err
	}
	l.logg.Info("IP ", ipStr, " was added in the white list")
	return nil
}

func (l *IPList) DelWhiteList(ctx context.Context, ipStr string) error {
	if err := checkParam(ipStr); err != nil {
		return fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	ipNet, err := netip.ParsePrefix(ipStr)
	if err != nil {
		return fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	if err := l.storage.DelWhiteList(ctx, ipNet.Masked()); err != nil {
		return err
	}
	l.logg.Info("IP ", ipStr, " was deleted from the white list")
	return nil
}

func (l *IPList) AddBlackList(ctx context.Context, ipStr string) error {
	if err := checkParam(ipStr); err != nil {
		return fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	ipNet, err := netip.ParsePrefix(ipStr)
	if err != nil {
		return fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	if err := l.storage.AddBlackList(ctx, ipNet.Masked()); err != nil {
		return err
	}
	l.logg.Info("IP ", ipStr, " was added in the black list")
	return nil
}

func (l *IPList) DelBlackList(ctx context.Context, ipStr string) error {
	if err := checkParam(ipStr); err != nil {
		return fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	ipNet, err := netip.ParsePrefix(ipStr)
	if err != nil {
		return fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	if err := l.storage.DelBlackList(ctx, ipNet.Masked()); err != nil {
		return err
	}
	l.logg.Info("IP ", ipStr, " was deleted from the black list")
	return nil
}

func (l *IPList) IsInWhiteList(ctx context.Context, ipStr string) (bool, error) {
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return false, fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	exist, err := l.storage.IsInWhiteList(ctx, ip)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (l *IPList) IsInBlackList(ctx context.Context, ipStr string) (bool, error) {
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return false, fmt.Errorf("%s: %w", BadRequestStr, err)
	}
	exist, err := l.storage.IsInBlackList(ctx, ip)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (l *IPList) GetList(ctx context.Context, t string) []string {
	var list []string
	var err error
	if t == "w" {
		list, err = l.storage.GetWhiteList(ctx)
	} else if t == "b" {
		list, err = l.storage.GetBlackList(ctx)
	}
	if err != nil {
		l.logg.Error("GetList: ", err)
		return nil
	}
	return list
}

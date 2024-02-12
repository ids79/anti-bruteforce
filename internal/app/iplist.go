package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
	"github.com/ids79/anti-bruteforce/internal/storage"
)

var ErrBadRequest = errors.New("entered invalid parameter")

var ErrIncorrectIP = errors.New("incorrect IP address")

type WorkWithIPList interface {
	AddWhiteList(ipStr string, maskStr string) error
	DelWhiteList(ipStr string) error
	AddBlackList(ipStr string, maskStr string) error
	DelBlackList(ipStr string) error
	IsInWhiteList(ipStr string) (bool, error)
	IsInBlackList(ipStr string) (bool, error)
	GetList(t string) []IPItem
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

func checkParam(ipStr, maskStr string, logg logger.Logg) bool {
	if ipStr == "" {
		logg.Error("ip parametr was not faund")
		return false
	}
	if maskStr == "" {
		logg.Error("mask parametr was not faund")
		return false
	}
	return true
}

func IPtoInt(val string) (int, error) {
	m := strings.Split(val, ".")
	if len(m) != 4 {
		return 0, ErrIncorrectIP
	}
	ip1, err := strconv.Atoi(m[0])
	if err != nil {
		return 0, ErrIncorrectIP
	}
	ip2, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, ErrIncorrectIP
	}
	ip3, err := strconv.Atoi(m[2])
	if err != nil {
		return 0, ErrIncorrectIP
	}
	ip4, err := strconv.Atoi(m[3])
	if err != nil {
		return 0, ErrIncorrectIP
	}
	return ip1<<24 + ip2<<16 + ip3<<8 + ip4, nil
}

func IPtoStr(val int) string {
	bild := strings.Builder{}
	bild.WriteString(strconv.Itoa(val >> 24))
	bild.WriteString(".")
	bild.WriteString(strconv.Itoa(val & 0b00000000111111110000000000000000 >> 16))
	bild.WriteString(".")
	bild.WriteString(strconv.Itoa(val & 0b00000000000000001111111100000000 >> 8))
	bild.WriteString(".")
	bild.WriteString(strconv.Itoa(val & 0b00000000000000000000000011111111))
	return bild.String()
}

func GetIPRange(ipStr, mask string) (r storage.IPItem, err error) {
	r.IP, err = IPtoInt(ipStr)
	if err != nil {
		return
	}
	r.Mask, err = strconv.Atoi(mask)
	if err != nil {
		return
	}
	s1 := strings.Repeat("1", r.Mask) + strings.Repeat("0", 32-r.Mask)
	s2 := strings.Repeat("0", r.Mask) + strings.Repeat("1", 32-r.Mask)
	m1, _ := strconv.ParseInt(s1, 2, 64)
	m2, _ := strconv.ParseInt(s2, 2, 64)
	r.IPfrom = r.IP & int(m1)
	r.IPto = r.IP | int(m2)
	return
}

func (l *IPList) AddWhiteList(ipStr string, maskStr string) error {
	if !checkParam(ipStr, maskStr, l.logg) {
		return fmt.Errorf("%w: parametr ip or mask not found", ErrBadRequest)
	}
	r, err := GetIPRange(ipStr, maskStr)
	if err != nil {
		l.logg.Error("AddWhiteList: ", err)
		return fmt.Errorf("%w: error get range", ErrBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	if err := l.storage.AddWhiteList(ctx, r); err != nil {
		l.logg.Error("AddWhiteList: ", err)
		return err
	}
	l.logg.Info("IP ", r.IP, " was added in the white list")
	return nil
}

func (l *IPList) DelWhiteList(ipStr string) error {
	if !checkParam(ipStr, "1", l.logg) {
		return fmt.Errorf("%w: parametr ip not found", ErrBadRequest)
	}
	ip, err := IPtoInt(ipStr)
	if err != nil {
		l.logg.Error("AddWhiteList: ", err)
		return fmt.Errorf("%w: error get ip from string", ErrBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	if err := l.storage.DelWhiteList(ctx, ip); err != nil {
		l.logg.Error("DelWhiteList: ", err)
		return err
	}
	l.logg.Info("IP ", ip, " was deleted from the white list")
	return nil
}

func (l *IPList) AddBlackList(ipStr string, maskStr string) error {
	if !checkParam(ipStr, maskStr, l.logg) {
		return fmt.Errorf("%w: parametr ip or mask not found", ErrBadRequest)
	}
	r, err := GetIPRange(ipStr, maskStr)
	if err != nil {
		l.logg.Error("AddWhiteList: ", err)
		return fmt.Errorf("%w: error get range", ErrBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	if err := l.storage.AddBlackList(ctx, r); err != nil {
		l.logg.Error("AddBlackList: ", err)
		return err
	}
	l.logg.Info("IP ", r.IP, " was added in the black list")
	return nil
}

func (l *IPList) DelBlackList(ipStr string) error {
	if !checkParam(ipStr, "1", l.logg) {
		return fmt.Errorf("%w: parametr ip not found", ErrBadRequest)
	}
	ip, err := IPtoInt(ipStr)
	if err != nil {
		l.logg.Error("AddWhiteList: ", err)
		return fmt.Errorf("%w: error get ip from string", ErrBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	if err := l.storage.DelBlackList(ctx, ip); err != nil {
		l.logg.Error("DelBlackList: ", err)
		return err
	}
	l.logg.Info("IP ", ip, " was deleted from the black list")
	return nil
}

func (l *IPList) IsInWhiteList(ipStr string) (bool, error) {
	ip, err := IPtoInt(ipStr)
	if err != nil {
		l.logg.Error("AddWhiteList: ", err)
		return false, fmt.Errorf("%w: error get ip from string", ErrBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	exist, err := l.storage.IsInWhiteList(ctx, ip)
	if err != nil {
		l.logg.Error("IsInWhiteList: ", err)
		return false, err
	}
	return exist, nil
}

func (l *IPList) IsInBlackList(ipStr string) (bool, error) {
	ip, err := IPtoInt(ipStr)
	if err != nil {
		l.logg.Error("AddWhiteList: ", err)
		return false, fmt.Errorf("%w: error get ip from string", ErrBadRequest)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	exist, err := l.storage.IsInBlackList(ctx, ip)
	if err != nil {
		l.logg.Error("IsInBlackList: ", err)
		return false, err
	}
	return exist, nil
}

func listFormBaseToApp(list []storage.IPItem) []IPItem {
	l := make([]IPItem, len(list))
	for i, it := range list {
		l[i] = IPItem{
			IP:     IPtoStr(it.IP),
			Mask:   strconv.Itoa(it.Mask),
			IPfrom: IPtoStr(it.IPfrom),
			IPto:   IPtoStr(it.IPto),
		}
	}
	return l
}

func (l *IPList) GetList(t string) []IPItem {
	var list []storage.IPItem
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	if t == "w" {
		list, err = l.storage.GetWhiteList(ctx)
	} else if t == "b" {
		list, err = l.storage.GetBlackList(ctx)
	}
	if err != nil {
		l.logg.Error("GetList: ", err)
		return nil
	}
	return listFormBaseToApp(list)
}

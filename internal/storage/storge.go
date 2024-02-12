package storage

import (
	"context"
	"embed"
	"errors"
	"time"

	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

var ErrIPAlreadyExistInWhiteRange = errors.New("this IP is already exist in white range")

var ErrIPAlreadyExistInBlackRange = errors.New("this IP is already exist in black range")

type Storage interface {
	AddWhiteList(ctx context.Context, item IPItem) error
	DelWhiteList(ctx context.Context, ip int) error
	AddBlackList(ctx context.Context, item IPItem) error
	DelBlackList(ctx context.Context, ip int) error
	IsInWhiteList(ctx context.Context, ip int) (bool, error)
	IsInBlackList(ctx context.Context, ip int) (bool, error)
	GetBlackList(ctx context.Context) ([]IPItem, error)
	GetWhiteList(ctx context.Context) ([]IPItem, error)
	Close(logg logger.Logg) error
}

type storage struct {
	connStr string
	conn    *sqlx.DB
}

type IPItem struct {
	IP     int
	Mask   int
	IPfrom int
	IPto   int
}

func New(logg logger.Logg, config config.Config) (Storage, error) {
	stor := &storage{
		connStr: config.Database.ConnectString,
	}
	err := stor.connect(logg)
	if err != nil {
		return nil, err
	}
	err = stor.migration(logg)
	if err != nil {
		return nil, err
	}
	return stor, nil
}

func (s *storage) connect(logg logger.Logg) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()
	s.conn, err = sqlx.ConnectContext(ctx, "pgx", s.connStr)
	if err != nil {
		logg.Error("cannot connect to base psql: ", err)
		return err
	}
	return s.conn.Ping()
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (s *storage) migration(logg logger.Logg) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		logg.Error("Data migration failed with an error: ", err)
		return err
	}
	// if err := goose.Down(s.conn.DB, "migrations"); err != nil {
	// 	s.logg.Error("Data migration failed with an error: ", err)
	// 	return err
	// }
	if err := goose.Up(s.conn.DB, "migrations"); err != nil {
		logg.Error("Data migration failed with an error: ", err)
		return err
	}
	logg.Info("Data migration was successful")
	return nil
}

func (s *storage) Close(logg logger.Logg) error {
	if err := s.conn.DB.Close(); err != nil {
		logg.Error(err)
		return err
	}
	logg.Info("connect to storage is closed")
	return nil
}

func (s *storage) AddWhiteList(ctx context.Context, item IPItem) error {
	query := `select * from blacklist where ipfrom <= $1 and ipto >= $1`
	rows, err := s.conn.QueryContext(ctx, query, item.IP)
	if err != nil {
		return err
	}
	if rows.Next() {
		return ErrIPAlreadyExistInBlackRange
	}
	err = s.DelWhiteList(ctx, item.IP)
	if err != nil {
		return err
	}
	query = `insert into whitelist (ip, mask, ipfrom, ipto) values(:ip, :mask, :ipfrom, :ipto)`
	_, err = s.conn.NamedQueryContext(ctx, query, item)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) DelWhiteList(ctx context.Context, ip int) error {
	query := `delete from whitelist where ip = $1`
	_, err := s.conn.ExecContext(ctx, query, ip)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) AddBlackList(ctx context.Context, item IPItem) error {
	query := `select * from whitelist where ipfrom <= $1 and ipto >= $1`
	rows, err := s.conn.QueryContext(ctx, query, item.IP)
	if err != nil {
		return err
	}
	if rows.Next() {
		return ErrIPAlreadyExistInWhiteRange
	}
	err = s.DelBlackList(ctx, item.IP)
	if err != nil {
		return err
	}
	query = `insert into blacklist (ip, mask, ipfrom, ipto) values(:ip, :mask, :ipfrom, :ipto)`
	_, err = s.conn.NamedQueryContext(ctx, query, item)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) DelBlackList(ctx context.Context, ip int) error {
	query := `delete from blacklist where ip = $1`
	_, err := s.conn.ExecContext(ctx, query, ip)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) IsInWhiteList(ctx context.Context, ip int) (bool, error) {
	query := `select * from whitelist where ipfrom <= $1 and ipto >= $1`
	rows, err := s.conn.QueryContext(ctx, query, ip)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

func (s *storage) IsInBlackList(ctx context.Context, ip int) (bool, error) {
	query := `select * from blacklist where ipfrom <= $1 and ipto >= $1`
	rows, err := s.conn.QueryContext(ctx, query, ip)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

func getRows(rows *sqlx.Rows) ([]IPItem, error) {
	list := make([]IPItem, 0)
	for rows.Next() {
		var ipem IPItem
		err := rows.StructScan(&ipem)
		if err != nil {
			return nil, err
		}
		list = append(list, ipem)
	}
	return list, nil
}

func (s *storage) GetWhiteList(ctx context.Context) ([]IPItem, error) {
	query := `select * from whitelist`
	selection, err := s.conn.NamedQueryContext(ctx, query, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	rows, err := getRows(selection)
	return rows, err
}

func (s *storage) GetBlackList(ctx context.Context) ([]IPItem, error) {
	query := `select * from blacklist`
	selection, err := s.conn.NamedQueryContext(ctx, query, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	rows, err := getRows(selection)
	return rows, err
}

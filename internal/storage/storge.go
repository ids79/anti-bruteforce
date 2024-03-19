package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"net/netip"
	"time"

	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

var ErrIPAlreadyExistInWhiteRange = errors.New("this IP is already exist in white range")

var ErrIPAlreadyExistInBlackRange = errors.New("this IP is already exist in black range")

type Storage interface {
	AddWhiteList(ctx context.Context, ip netip.Prefix) error
	DelWhiteList(ctx context.Context, ip netip.Prefix) error
	AddBlackList(ctx context.Context, ip netip.Prefix) error
	DelBlackList(ctx context.Context, ip netip.Prefix) error
	IsInWhiteList(ctx context.Context, ip netip.Addr) (bool, error)
	IsInBlackList(ctx context.Context, ip netip.Addr) (bool, error)
	GetBlackList(ctx context.Context) ([]string, error)
	GetWhiteList(ctx context.Context) ([]string, error)
	Close(logg logger.Logg)
}

type storage struct {
	connStr string
	conn    *sqlx.DB
}

func New(config *config.Config) (Storage, error) {
	stor := &storage{
		connStr: config.Database.ConnectString,
	}
	err := stor.connect()
	if err != nil {
		return nil, err
	}
	err = stor.migration()
	if err != nil {
		return nil, err
	}
	return stor, nil
}

func (s *storage) connect() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()
	s.conn, err = sqlx.ConnectContext(ctx, "pgx", s.connStr)
	if err != nil {
		return err
	}
	return s.conn.Ping()
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (s *storage) migration() error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	// if err := goose.Down(s.conn.DB, "migrations"); err != nil {
	// 	return err
	// }
	if err := goose.Up(s.conn.DB, "migrations"); err != nil {
		return err
	}
	return nil
}

func (s *storage) Close(logg logger.Logg) {
	if err := s.conn.DB.Close(); err != nil {
		logg.Error(err)
		return
	}
	logg.Info("connect to storage is closed")
}

func (s *storage) AddWhiteList(ctx context.Context, ip netip.Prefix) error {
	query := `select * from blacklist where ip  >>= $1`
	if s.conn.QueryRowContext(ctx, query, ip).Scan() != sql.ErrNoRows {
		return ErrIPAlreadyExistInBlackRange
	}
	err := s.DelWhiteList(ctx, ip)
	if err != nil {
		return err
	}
	query = `insert into whitelist (ip) values($1)`
	_, err = s.conn.ExecContext(ctx, query, ip)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) DelWhiteList(ctx context.Context, ip netip.Prefix) error {
	query := `delete from whitelist where ip = $1`
	_, err := s.conn.ExecContext(ctx, query, ip)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) AddBlackList(ctx context.Context, ip netip.Prefix) error {
	query := `select * from whitelist where ip  >>= $1`
	if s.conn.QueryRowContext(ctx, query, ip).Scan() != sql.ErrNoRows {
		return ErrIPAlreadyExistInWhiteRange
	}
	err := s.DelBlackList(ctx, ip)
	if err != nil {
		return err
	}
	query = `insert into blacklist (ip) values($1)`
	_, err = s.conn.ExecContext(ctx, query, ip)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) DelBlackList(ctx context.Context, ip netip.Prefix) error {
	query := `delete from blacklist where ip = $1`
	_, err := s.conn.ExecContext(ctx, query, ip)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) IsInWhiteList(ctx context.Context, ip netip.Addr) (bool, error) {
	query := `select * from whitelist where ip >>= $1`
	rows, err := s.conn.QueryContext(ctx, query, ip)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

func (s *storage) IsInBlackList(ctx context.Context, ip netip.Addr) (bool, error) {
	query := `select * from blacklist where ip >>= $1`
	rows, err := s.conn.QueryContext(ctx, query, ip)
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

func getRows(rows *sqlx.Rows) ([]string, error) {
	list := make([]string, 0)
	for rows.Next() {
		var ip string
		err := rows.Scan(&ip)
		if err != nil {
			return nil, err
		}
		list = append(list, ip)
	}
	return list, nil
}

func (s *storage) GetWhiteList(ctx context.Context) ([]string, error) {
	query := `select ip from whitelist`
	selection, err := s.conn.NamedQueryContext(ctx, query, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	defer selection.Close()
	rows, err := getRows(selection)
	return rows, err
}

func (s *storage) GetBlackList(ctx context.Context) ([]string, error) {
	query := `select ip from blacklist`
	selection, err := s.conn.NamedQueryContext(ctx, query, map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	defer selection.Close()
	rows, err := getRows(selection)
	return rows, err
}

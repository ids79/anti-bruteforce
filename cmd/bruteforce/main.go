package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/ids79/anti-bruteforce/internal/app"
	"github.com/ids79/anti-bruteforce/internal/app/inmem"
	"github.com/ids79/anti-bruteforce/internal/app/inredis"
	"github.com/ids79/anti-bruteforce/internal/config"
	"github.com/ids79/anti-bruteforce/internal/logger"
	internalgrpc "github.com/ids79/anti-bruteforce/internal/server/grpc"
	internalhttp "github.com/ids79/anti-bruteforce/internal/server/http"
	"github.com/ids79/anti-bruteforce/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
}

func main() {
	flag.Parse()
	if configFile == "" {
		configFile, _ = os.LookupEnv("CONFIG_FILE")
	}
	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	wg := sync.WaitGroup{}
	config, err := config.NewConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%W", err)
		return
	}
	logg := logger.New(config.Logger, "Anti-bruteforce:")
	storage, err := storage.New(config)
	if err != nil {
		logg.Error("psql initial error: ", err)
		return
	}
	var backets app.WorkWithBackets
	if strings.Compare(config.ExpireBase, "in_mem") == 0 {
		backets = inmem.NewBackets(ctx, logg, config)
	} else if strings.Compare(config.ExpireBase, "in_radis") == 0 {
		backets = inredis.NewBackets(logg, config)
		if backets == nil {
			return
		}
	}
	iplist := app.NewIPList(storage, logg, config)
	serverHTTP := internalhttp.NewServer(backets, iplist, logg, config)
	serverGRPC := internalgrpc.NewServer(backets, iplist, logg, config)
	logg.Info("anti-bruteforce is running...")

	wg.Add(1)
	go func() {
		defer wg.Done()
		serverHTTP.Start()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverGRPC.Start()
	}()

	<-ctx.Done()
	serverHTTP.Close()
	serverGRPC.Close()
	storage.Close(logg)
	wg.Wait()
}

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ids79/anti-bruteforcer/internal/app"
	"github.com/ids79/anti-bruteforcer/internal/config"
	"github.com/ids79/anti-bruteforcer/internal/logger"
	internalgrpc "github.com/ids79/anti-bruteforcer/internal/server/grpc"
	internalhttp "github.com/ids79/anti-bruteforcer/internal/server/http"
	"github.com/ids79/anti-bruteforcer/internal/storage"
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
	config := config.NewConfig(configFile)
	logg := logger.New(config.Logger, "Anti-bruteforce:")
	storage, err := storage.New(logg, config)
	if err != nil {
		return
	}
	backets := app.NewBackets(ctx, logg, &config)
	iplist := app.NewIPList(storage, logg, &config)
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

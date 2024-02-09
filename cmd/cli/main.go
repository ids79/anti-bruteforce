package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/ids79/anti-bruteforcer/internal/cli"
	"github.com/ids79/anti-bruteforcer/internal/config"
	"github.com/ids79/anti-bruteforcer/internal/logger"
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
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	config := config.NewConfig(configFile)
	logg := logger.New(config.Logger, "Anti-bruteforce:")

	cli := cli.NewCLI(logg, &config)
	go cli.Start(ctx)

	<-ctx.Done()
}

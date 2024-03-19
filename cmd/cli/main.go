package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ids79/anti-bruteforce/internal/cli"
)

var address string

func init() {
	flag.StringVar(&address, "grpc", "", "Address to grpc server")
}

func main() {
	flag.Parse()
	if address == "" {
		address, _ = os.LookupEnv("GRPC_ADDRESS")
	}
	if address == "" {
		fmt.Println("The address for connecting to the GPS service is not specified")
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()
	cli := cli.New(address)
	go cli.Start(ctx)

	<-ctx.Done()
}

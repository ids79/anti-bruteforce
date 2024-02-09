package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ids79/anti-bruteforcer/internal/app"
	"github.com/ids79/anti-bruteforcer/internal/config"
	"github.com/ids79/anti-bruteforcer/internal/logger"
	"github.com/ids79/anti-bruteforcer/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type CLI struct {
	logg   logger.Logg
	client pb.CLiApiClient
}

func NewCLI(logg logger.Logg, conf *config.Config) *CLI {
	conn, err := grpc.Dial(conf.GRPCServer.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logg.Error(err)
	}
	return &CLI{
		logg:   logg,
		client: pb.NewCLiApiClient(conn),
	}
}

func (c *CLI) Start(ctx context.Context) {
	time.Sleep(time.Second)
	fmt.Println("Welcom to CLI interface!")
	inScanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Enter your command:")
		select {
		case <-ctx.Done():
			return
		default:
			if !inScanner.Scan() {
				if inScanner.Err() != nil {
					c.logg.Error(inScanner.Err())
					break
				}
			}
			str := inScanner.Text()
			if len(str) == 0 {
				time.Sleep(time.Microsecond * 300)
				break
			}
			command := str[:4]
			var params []string
			if len(str) > 4 {
				params = strings.Split(strings.TrimSpace(str[5:]), " ")
			}
			c.performCommand(command, params)
		}
	}
}

func printList(list []app.IPItem) {
	if len(list) > 0 {
		for _, item := range list {
			fmt.Println(item)
		}
	} else {
		fmt.Println(("white list is empty"))
	}
}

func checkParams(params []string, num int) bool {
	if len(params) < num {
		fmt.Println("not enough parameters")
		return false
	}
	return true
}

func IPFormPBtoApp(list *pb.List) []app.IPItem {
	l := make([]app.IPItem, len(list.Items))
	for i, it := range list.Items {
		l[i] = app.IPItem{
			IP:     it.IP,
			Mask:   it.Mask,
			IPfrom: it.IPfrom,
			IPto:   it.IPto,
		}
	}
	return l
}

func setMetadata(ctx context.Context, method string) context.Context {
	md := metadata.New(nil)
	md.Set("method", method)
	fmt.Println(method)
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *CLI) performCommand(command string, params []string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	switch command {
	case "addw":
		if !checkParams(params, 2) {
			break
		}
		ctx = setMetadata(ctx, "add white")
		res, err := c.client.AddWhite(ctx, &pb.IPMask{IP: params[0], Mask: params[1]})
		fmt.Println(res, err)
	case "addb":
		if !checkParams(params, 2) {
			break
		}
		ctx = setMetadata(ctx, "add black")
		res, err := c.client.AddBlack(ctx, &pb.IPMask{IP: params[0], Mask: params[1]})
		fmt.Println(res, err)
	case "delw":
		if !checkParams(params, 1) {
			break
		}
		ctx = setMetadata(ctx, "del white")
		res, err := c.client.DelWhite(ctx, &pb.IP{IP: params[0]})
		fmt.Println(res, err)
	case "delb":
		if !checkParams(params, 1) {
			break
		}
		ctx = setMetadata(ctx, "del black")
		res, err := c.client.DelBlack(ctx, &pb.IP{IP: params[0]})
		fmt.Println(res, err)
	case "resi":
		if !checkParams(params, 1) {
			break
		}
		ctx = setMetadata(ctx, "reset backet")
		res, err := c.client.ResetBacket(ctx, &pb.Backet{Backet: params[0], Type: "IP"})
		fmt.Println(res, err)
	case "resl":
		if !checkParams(params, 1) {
			break
		}
		ctx = setMetadata(ctx, "reset backet")
		res, err := c.client.ResetBacket(ctx, &pb.Backet{Backet: params[0], Type: "LOGIN"})
		fmt.Println(res, err)
	case "resp":
		if !checkParams(params, 1) {
			break
		}
		ctx = setMetadata(ctx, "reset backet")
		res, err := c.client.ResetBacket(ctx, &pb.Backet{Backet: params[0], Type: "PASS"})
		fmt.Println(res, err)
	case "getw":
		ctx = setMetadata(ctx, "get white list")
		list, err := c.client.GetList(ctx, &pb.TypeList{Type: "w"})
		if err != nil {
			c.logg.Error(err)
		}
		printList(IPFormPBtoApp(list))
	case "getb":
		ctx = setMetadata(ctx, "get black list")
		list, err := c.client.GetList(ctx, &pb.TypeList{Type: "b"})
		if err != nil {
			c.logg.Error(err)
		}
		printList(IPFormPBtoApp(list))
	default:
		fmt.Println("Entered command is not correct")
	}
}

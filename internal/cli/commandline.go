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

func (c *CLI) performCommand(command string, params []string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	switch command {
	case "addw":
		if len(params) < 2 {
			fmt.Println("not enough parameters")
			break
		}
		res, err := c.client.AddWhite(ctx, &pb.IPMask{IP: params[0], Mask: params[1]})
		if err != nil {
			c.logg.Error(err)
		}
		fmt.Println(res)
	// case "addb":
	// 	if len(params) < 2 {
	// 		fmt.Println("not enough parameters")
	// 		break
	// 	}
	// 	err := c.iplist.AddBlackList(params[0], params[1])
	// 	if errors.Is(err, app.ErrBadRequest) {
	// 		fmt.Println(err)
	// 	}
	case "delw":
		if len(params) != 1 {
			fmt.Println("not enough parameters")
			break
		}
		res, err := c.client.DelWhite(ctx, &pb.IP{IP: params[0]})
		if err != nil {
			c.logg.Error(err)
		}
		fmt.Println(res)
	// case "delb":
	// 	if len(params) != 1 {
	// 		fmt.Println("not enough parameters")
	// 		break
	// 	}
	// 	err := c.iplist.DelBlackList(params[0])
	// 	if errors.Is(err, app.ErrBadRequest) {
	// 		fmt.Println(err)
	// 	}
	// case "resi":
	// 	if c.backets.ResetBacket(params[0], app.IP) != nil {
	// 		fmt.Println("IP backet was no found in backets")
	// 		break
	// 	}
	// 	fmt.Println(("IP backet was reset"))
	// case "resl":
	// 	if c.backets.ResetBacket(params[0], app.LOGIN) != nil {
	// 		fmt.Println("login backet was no found in backets")
	// 		break
	// 	}
	// 	fmt.Println(("login backet was reset"))
	// case "resp":
	// 	if c.backets.ResetBacket(params[0], app.PASSWORD) != nil {
	// 		fmt.Println("password backet was no found in backets")
	// 		break
	// 	}
	// 	fmt.Println(("password backet was reset"))
	// case "getw":
	// 	list := c.iplist.GetList("w")
	// 	printList(list)
	// case "getb":
	// 	list := c.iplist.GetList("b")
	// 	printList(list)
	default:
		fmt.Println("Entered command is not correct")
	}
}

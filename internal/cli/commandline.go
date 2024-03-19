package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ids79/anti-bruteforce/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type CLI struct {
	client pb.CLiApiClient
}

func New(address string) *CLI {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
	}
	return &CLI{
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
					fmt.Println(inScanner.Err())
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

func printList(list []string) {
	if len(list) == 0 {
		fmt.Println("list is empty")
		return
	}
	for _, item := range list {
		fmt.Println(item)
	}
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
		if len(params) != 1 {
			fmt.Println("not enough parameters")
			break
		}
		ctx = setMetadata(ctx, "add white")
		res, err := c.client.AddWhite(ctx, &pb.IP{IP: params[0]})
		fmt.Println(res, err)
	case "addb":
		if len(params) != 1 {
			fmt.Println("not enough parameters")
			break
		}
		ctx = setMetadata(ctx, "add black")
		res, err := c.client.AddBlack(ctx, &pb.IP{IP: params[0]})
		fmt.Println(res, err)
	case "delw":
		if len(params) != 1 {
			fmt.Println("not enough parameters")
			break
		}
		ctx = setMetadata(ctx, "del white")
		res, err := c.client.DelWhite(ctx, &pb.IP{IP: params[0]})
		fmt.Println(res, err)
	case "delb":
		if len(params) != 1 {
			fmt.Println("not enough parameters")
			break
		}
		ctx = setMetadata(ctx, "del black")
		res, err := c.client.DelBlack(ctx, &pb.IP{IP: params[0]})
		fmt.Println(res, err)
	case "resi":
		if len(params) != 1 {
			fmt.Println("not enough parameters")
			break
		}
		ctx = setMetadata(ctx, "reset backet")
		res, err := c.client.ResetBacket(ctx, &pb.Backet{Backet: params[0], Type: "IP"})
		fmt.Println(res, err)
	case "resl":
		if len(params) != 1 {
			fmt.Println("not enough parameters")
			break
		}
		ctx = setMetadata(ctx, "reset backet")
		res, err := c.client.ResetBacket(ctx, &pb.Backet{Backet: params[0], Type: "LOGIN"})
		fmt.Println(res, err)
	case "resp":
		if len(params) != 1 {
			fmt.Println("not enough parameters")
			break
		}
		ctx = setMetadata(ctx, "reset backet")
		res, err := c.client.ResetBacket(ctx, &pb.Backet{Backet: params[0], Type: "PASS"})
		fmt.Println(res, err)
	case "getw":
		ctx = setMetadata(ctx, "get white list")
		list, err := c.client.GetList(ctx, &pb.TypeList{Type: "w"})
		if err != nil {
			fmt.Println(err)
		}
		printList(list.Items)
	case "getb":
		ctx = setMetadata(ctx, "get black list")
		list, err := c.client.GetList(ctx, &pb.TypeList{Type: "b"})
		if err != nil {
			fmt.Println(err)
		}
		printList(list.Items)
	default:
		fmt.Println("Entered command is not correct")
	}
}

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/cli/client"
	"github.com/CodethinkLabs/wago/pkg/cli/common"
	"github.com/CodethinkLabs/wago/pkg/proto"
	"google.golang.org/grpc"
	"time"
)

func main() {
	cluster := flag.String("cluster", "localhost:8080", "wago cluster to connect to")
	flag.Parse()

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Could not connect to server at %s... exiting\n", *cluster)
		panic(err)
	}
	defer conn.Close()

	protoClient := proto.NewWalletServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
	defer cancel()

	executor, completer := cli.CreateCLI(
		common.NewCommand,
		common.DeleteCommand,
		common.AuthCommand,
		client.BankCommand(ctx, protoClient),
		client.SendCommand(ctx, protoClient),
		client.CreateCommand(ctx, protoClient),
	)
	cli.StartCLI(executor, completer)
}

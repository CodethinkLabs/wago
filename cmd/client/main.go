// Client runs a non-participating raft client that can send
// receive currency. Being non-participating, this node does
// not take part in the application and replication of commits.
// It is only able to submit transactions to be committed to
// the cluster on their behalf.
//
// Program flags:
//	--cluster: 	The IP address of any node in the cluster you
//				wish to send transactions to.
//
// The client will crash on terminals without a tty.
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/cli/client"
	"github.com/CodethinkLabs/wago/pkg/cli/common"
	"github.com/CodethinkLabs/wago/pkg/proto"
	"google.golang.org/grpc"
)

func main() {
	cluster := flag.String("cluster", "localhost:8080", "wago cluster to connect to")
	flag.Parse()

	conn, err := grpc.Dial(*cluster, grpc.WithInsecure())
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

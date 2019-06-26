// package main is the entry point into the server
// command line interface
//
// Program flags:
//  --cluster: A comma separated list of peer IP addresses
//  --id:      This node's index in the list of peers
//  --join:    Whether this node is joining an existing cluster
//
// The cluster string should be identical between all nodes.
// Because the ID has to be unique between nodes, we can use that
// to assign the addresses. ID 1 takes the first IP address and so on.
package main

import (
	"flag"
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/cli"
	wagoRaft "github.com/CodethinkLabs/wago/pkg/raft"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	etcdRaft "go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	cluster := flag.String("cluster", "http://127.0.0.1:9020", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()

	proposeC := make(chan string)               // for state machine proposals
	confChangeC := make(chan raftpb.ConfChange) // for config proposals (peer layout)
	defer close(proposeC)
	defer close(confChangeC)

	var store *wallet.Store
	getSnapshot := func() ([]byte, error) { return store.GetSnapshot() }

	// channels for all the validated commits, errors, and an indicator when snapshots are ready
	// this starts many goroutines+-
	commitC, errorC, snapshotterReady, statusGetter := wagoRaft.NewRaftNode(*id, strings.Split(*cluster, ","), *join, getSnapshot, proposeC, confChangeC)

	// initialize the chat store with all the channels
	// this starts a goroutine
	wg := sync.WaitGroup{}
	wg.Add(1)
	store = wallet.NewStore(<-snapshotterReady, proposeC, commitC, errorC, wg)

	if hasTTY() {
		// if we have access to a tty, start the CLI
		// run in a goroutine so that if raft closes, the CLI exits
		go func() {
			executor, completer := cli.CreateCLI(
				server.BankCommand(store),
				server.SendCommand(store),
				server.CreateCommand(store),
				common.NewCommand,
				common.DeleteCommand,
				common.AuthCommand,
				server.NodeCommand(confChangeC, statusGetter),
				server.StatusCommand(statusGetter),
			)
			cli.StartCLI(executor, completer)
		}()
	} else {
		// else just start in "headless mode"
		fmt.Println("Starting in headless mode.")
	}

	wg.Wait()
	fmt.Println("Raft closed. Goodbye.")
}

func init() {
	disableLogging()
}

func hasTTY() bool {
	in, err := syscall.Open("/dev/tty", syscall.O_RDONLY, 0)
	if err != nil {
		return false
	} else {
		syscall.Close(in)
		return true
	}
}

func disableLogging() {
	discard := log.New(ioutil.Discard, "", 0)
	etcdRaft.SetLogger(&etcdRaft.DefaultLogger{Logger: discard})
	wagoRaft.Log = discard
	// raft zap logger
	prodZapLog, err := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.ErrorLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	if err != nil {
		panic(err)
	}
	wagoRaft.ZapLog = prodZapLog
}

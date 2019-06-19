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
	"github.com/CodethinkLabs/wago/pkg/cli"
	wagoRaft "github.com/CodethinkLabs/wago/pkg/raft"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"github.com/c-bata/go-prompt"
	etcdRaft "go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"io/ioutil"
	"log"
	"strings"

	"go.uber.org/zap"
)

var store *wallet.WalletStore

// disable logging to stdout
func init() {
	discard := log.New(ioutil.Discard, "", 0)
	etcdRaft.SetLogger(&etcdRaft.DefaultLogger{Logger: discard})
	wagoRaft.Log = discard

	// raft zap logger
	prodZapLog, err := zap.Config{
		Level: zap.NewAtomicLevelAt(zap.ErrorLevel),
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

func main() {
	cluster := flag.String("cluster", "http://127.0.0.1:9020", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()

	proposeC := make(chan string)               // for state machine proposals
	confChangeC := make(chan raftpb.ConfChange) // for config proposals (peer layout)
	defer close(proposeC)
	defer close(confChangeC)

	// channels for all the validated commits, errors, and an indicator when snapshots are ready
	commitC, errorC, snapshotterReady := wagoRaft.NewRaftNode(*id, strings.Split(*cluster, ","), *join, store.GetSnapshot, proposeC, confChangeC)

	// initialize the chat store with all the channels
	store = wallet.NewWalletStore(<-snapshotterReady, proposeC, commitC, errorC)

	p := prompt.New(Executor, Completer)
	p.Run()
}

// a wrapper around the cli executor to provide it the store
func Executor(in string) {
	cli.StoreExecutor(in, store)
}

// a wrapper around the cli completer to provide it the store
func Completer(in prompt.Document) []prompt.Suggest {
	return cli.StoreCompleter(in, store)
}

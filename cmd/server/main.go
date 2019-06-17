package main

import (
	"flag"
	"github.com/c-bata/go-prompt"
	"go.etcd.io/etcd/raft/raftpb"
	"strings"
	"wago/pkg/cli"
	"wago/pkg/raft"
	"wago/pkg/wallet"
)

var store *wallet.WalletStore

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
	commitC, errorC, snapshotterReady := raft.NewRaftNode(*id, strings.Split(*cluster, ","), *join, store.GetSnapshot, proposeC, confChangeC)

	// initialize the chat store with all the channels
	store = wallet.NewWalletStore(<-snapshotterReady, proposeC, commitC, errorC)

	p := prompt.New(Executor, Completer)
	p.Run()
}

func Executor(in string) {
	cli.EExecutor(in, store)
}

func Completer(in prompt.Document) []prompt.Suggest {
	return cli.CCompleter(in, store)
}

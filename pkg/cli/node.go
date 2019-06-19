package cli

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"go.etcd.io/etcd/raft/raftpb"
	"strconv"
	"strings"
)

var subCommands Commands

// executes the node command, allowing members
// of the cluster to add and remove nodes
func NodeCommand(confChangeC chan<- raftpb.ConfChange) Command {
	nodeCreateExecutor := func(args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("command takes exactly ${ID} ${URL}")
		}

		nodeId, err := strconv.ParseUint(args[1], 0, 64)
		if err != nil {
			return fmt.Errorf("node id not a valid integer")
		}
		url := args[2]

		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  nodeId,
			Context: []byte(url),
		}
		confChangeC <- cc
		return nil
	}

	nodeDeleteExecutor := func(i []string) error {
		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeRemoveNode,
			NodeID:  1/*todo*/,
		}
		confChangeC <- cc
		return nil
	}

	nodeDeleteCompleter := func(in prompt.Document) []prompt.Suggest {
		// todo autocomplete a delete
		return []prompt.Suggest{{Text: "1"}}
	}
	subCommands = Commands{
		createCommand("create", "Add a node to the cluster", nodeCreateExecutor, nil),
		createCommand("delete", "Delete a node from the cluster", nodeDeleteExecutor, nodeDeleteCompleter),
	}

	return createCommand("node", "Configure the cluster layout", nodeExecutor, nodeCompleter)
}

func nodeExecutor(args []string) error {

	if len(args) > 1 {
		subCommand, err := subCommands.Match(args[1])
		if err == nil {
			return subCommand.executor(args[1:])
		}
	}
	subCommands.GenerateHelp("Available subcommands:")
	return nil
}

func nodeCompleter(in prompt.Document) []prompt.Suggest {
	currentCommand := in.TextBeforeCursor()
	args := strings.Split(currentCommand, " ")

	var subCommand Command
	if len(args) > 1 {
		var err error
		subCommand, err = subCommands.Match(args[1])
		if err != nil {
			return subCommands.GenerateSuggestions()
		}
	}
	if len(args) > 2 && subCommand.completer != nil {
		return subCommand.completer(in)
	}

	return []prompt.Suggest{}
}

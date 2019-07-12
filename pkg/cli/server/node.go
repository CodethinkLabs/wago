package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/raft"
	"github.com/c-bata/go-prompt"
	"go.etcd.io/etcd/raft/raftpb"
)

// NodeCommand creates the node command, allowing members
// of the cluster to add and remove nodes
//
// syntax: node create|delete ${ID} [URL]
func NodeCommand(confChangeC chan<- raftpb.ConfChange, statusGetter func() (raft.Status, error)) cli.Command {
	nodeCreateExecutor := func(args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("command takes exactly ${ID} ${URL}")
		}

		nodeID, err := strconv.ParseUint(args[1], 0, 64)
		if err != nil {
			return fmt.Errorf("node id not a valid integer")
		}
		url := args[2]

		cc := raftpb.ConfChange{
			Type:    raftpb.ConfChangeAddNode,
			NodeID:  nodeID,
			Context: []byte(url),
		}
		confChangeC <- cc
		return nil
	}

	nodeDeleteExecutor := func(args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("command takes exactly ${ID}")
		}

		nodeID, err := strconv.ParseUint(args[1], 0, 64)
		if err != nil {
			return fmt.Errorf("node id not a valid integer")
		}

		cc := raftpb.ConfChange{
			Type:   raftpb.ConfChangeRemoveNode,
			NodeID: nodeID,
		}
		confChangeC <- cc
		return nil
	}

	nodeDeleteCompleter := func(in prompt.Document) []prompt.Suggest {
		var suggestions []prompt.Suggest

		status, err := statusGetter()
		if err == nil {
			for _, nodeID := range status.Nodes {
				suggestions = append(suggestions, prompt.Suggest{Text: fmt.Sprint(nodeID)})
			}
		}

		return suggestions
	}

	subCommands := cli.Commands{
		cli.CreateCommand("create", "Add a node to the cluster", nodeCreateExecutor, nil),
		cli.CreateCommand("delete", "Delete a node from the cluster", nodeDeleteExecutor, nodeDeleteCompleter),
	}

	nodeExecutor := func(args []string) error {
		if len(args) > 1 {
			subCommand, err := subCommands.Match(args[1])
			if err == nil {
				return subCommand.Executor(args[1:])
			}
		}
		subCommands.GenerateHelp("Available subcommands:")
		return nil
	}

	nodeCompleter := func(in prompt.Document) []prompt.Suggest {
		currentCommand := in.TextBeforeCursor()
		args := strings.Split(currentCommand, " ")

		var subCommand cli.Command
		if len(args) > 1 {
			var err error
			subCommand, err = subCommands.Match(args[1])
			if err != nil {
				return subCommands.GenerateSuggestions()
			}
		}
		if len(args) > 2 && subCommand.Completer != nil {
			return subCommand.Completer(in)
		}

		return []prompt.Suggest{}
	}

	return cli.CreateCommand("node", "Configure the cluster layout", nodeExecutor, nodeCompleter)
}

package server

import (
	"fmt"

	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/raft"
)

// StatusCommand creates the status command which allows the
// user to access information about the status of the cluster.
//
// syntax: status
func StatusCommand(getStatus func() (raft.Status, error)) cli.Command {
	statusExecutor := func(strings []string) error {
		status, err := getStatus()
		if err != nil {
			return fmt.Errorf("could not get status, wait for cluster to initialize")
		}
		fmt.Println("Cluster status:")
		fmt.Printf("  - My ID: %v\n", status.ID)
		fmt.Printf("  - Node IDs: %v\n", status.Nodes)
		fmt.Printf("  - Active Nodes (%d): %v\n", len(status.ActiveNodes), status.ActiveNodes)

		return nil
	}

	return cli.CreateCommand("status", "Get the status of the cluster", statusExecutor, nil)
}

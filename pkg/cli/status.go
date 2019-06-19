package cli

import (
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/raft"
)

func StatusCommand(getStatus func() (raft.RaftStatus, error)) Command {
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

	return createCommand("status", "Get the status of the cluster", statusExecutor, nil)
}

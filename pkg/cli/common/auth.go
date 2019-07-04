package common

import (
	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/wallet"
)

// AuthCommand executes the auth command, allowing the
// user to set a password for the current session for
// encryption and decryption of the wallet file
//
// syntax: auth ${OPTIONAL_PASS}
var AuthCommand = cli.CreateCommand("auth", "Set the password for the session", authExecutor, nil)

func authExecutor(args []string) error {
	if len(args) != 2 {
		wallet.Authenticate(nil)
	} else {
		wallet.Authenticate(&args[1])
	}
	return nil
}

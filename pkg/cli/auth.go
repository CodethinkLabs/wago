package cli

import "github.com/CodethinkLabs/wago/pkg/wallet"

// executes the auth command, allowing the user to
// set a password for the current session for
// encryption and decryption of the wallet file
// syntax: auth ${OPTIONAL_PASS}
var AuthCommand = createCommand("auth", authExecutor, nil)

func authExecutor(args []string, store *wallet.WalletStore) error {
	if len(args) != 2 {
		wallet.Authenticate(nil)
	} else {
		wallet.Authenticate(&args[1])
	}
	return nil
}

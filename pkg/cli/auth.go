package cli

import "github.com/CodethinkLabs/wago/pkg/wallet"

func authCommand(args []string, store *wallet.WalletStore) error {
	if len(args) != 2 {
		wallet.Authenticate(nil)
	} else {
		wallet.Authenticate(&args[1])
	}
	return nil
}

var AuthCommand = createCommand("auth", authCommand, nil)
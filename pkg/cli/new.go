package cli

import (
	"github.com/CodethinkLabs/wago/pkg/wallet"
)

func newCommand(args []string, store *wallet.WalletStore) error {
	privateKey, publicKey := wallet.GenerateKeys(args[1:])
	walletFile := wallet.ReadWallet()
	walletFile = wallet.AddKeys(walletFile, publicKey, privateKey)
	wallet.WriteWallet(walletFile)
	return nil
}

var NewCommand = createCommand("new", newCommand, nil)


package cli

import (
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"golang.org/x/crypto/ed25519"
	"strings"
)

// executes the new command, allowing the user to
// generate new keypairs for use in the system
// syntax: new ${OPTIONAL MULTI-WORD SEED}
var NewCommand = createCommand("new", newCommand, nil)

func newCommand(args []string, store *wallet.WalletStore) error {
	var privateKey ed25519.PrivateKey
	var publicKey ed25519.PublicKey
	var err error

	if len(args) > 1 {
		seed := strings.Join(args[1:], " ")
		privateKey, publicKey, err = wallet.GenerateKeys(&seed)
	} else {
		privateKey, publicKey, err = wallet.GenerateKeys(nil)
	}
	if err != nil {
		return err
	}

	walletFile := wallet.ReadWallet()
	walletFile = wallet.AddKeys(walletFile, publicKey, privateKey)
	wallet.WriteWallet(walletFile)
	return nil
}

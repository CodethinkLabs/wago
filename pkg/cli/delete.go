package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"github.com/c-bata/go-prompt"
	"strings"
)

func deleteCommand(args []string, store *wallet.WalletStore) error {
	if len(args) < 2 {
		return fmt.Errorf("must provide an address")
	}
	key := args[1]
	walletFile := wallet.ReadWallet()
	i := 0
	for _, keyPair := range walletFile {
		if !strings.HasPrefix(hex.EncodeToString(keyPair.PublicKey), key) {
			walletFile[i] = keyPair
			i++
		} else {
			fmt.Printf("Removed key %x\n", keyPair.PublicKey[:6])
		}
	}
	if len(walletFile) == i {
		return fmt.Errorf("no matching keys in wallet")
	} else {
		walletFile = walletFile[:i]
	}
	wallet.WriteWallet(walletFile)
	return nil
}

func deleteCompleter(in prompt.Document, store *wallet.WalletStore) []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0)
	walletFile := wallet.ReadWallet()
	for _, keyPair := range walletFile {
		suggestions = append(suggestions, prompt.Suggest{Text: hex.EncodeToString(keyPair.PublicKey)[:12]})
	}

	return prompt.FilterFuzzy(suggestions, in.GetWordBeforeCursor(), true)
}

var DeleteCommand = createCommand("delete", deleteCommand, deleteCompleter)

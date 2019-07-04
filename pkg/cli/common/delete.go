package common

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"github.com/c-bata/go-prompt"
)

// DeleteCommand executes the delete command, allowing the
// user to remove keys from their personal wallet
//
// syntax: delete ${PUBLIC_KEY}
var DeleteCommand = cli.CreateCommand("delete", "Deletes a key from the local wallet", deleteExecutor, deleteCompleter)

func deleteExecutor(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("must provide exactly one address")
	}
	walletFile, err := wallet.ReadWallet()
	if err != nil {
		return err
	}

	key := args[1]
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
	}

	walletFile = walletFile[:i]
	wallet.WriteWallet(walletFile)
	return nil
}

func deleteCompleter(in prompt.Document) []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0)
	walletFile, err := wallet.ReadWallet()
	if err != nil {
		// if unauthed, dont suggest anything
		return []prompt.Suggest{}
	}
	for _, keyPair := range walletFile {
		suggestions = append(suggestions, prompt.Suggest{Text: hex.EncodeToString(keyPair.PublicKey)[:12]})
	}

	return prompt.FilterFuzzy(suggestions, in.GetWordBeforeCursor(), true)
}

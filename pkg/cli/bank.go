package cli

import (
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/util"
	"github.com/CodethinkLabs/wago/pkg/wallet"
)

// executes the bank command, returning to the user the
// current balance of all their local wallets
// syntax: bank [full]
func BankCommand(store *wallet.Store) Command {
	bankExecutor := func(args []string) error {
		walletFile, err := wallet.ReadWallet()
		if err != nil {
			return err
		}
		if len(walletFile) == 0 {
			fmt.Println("No credentials in wallet.")
			return nil
		}
		for _, key := range walletFile {
			pubKey := key.PublicKey
			if len(args) == 1 || args[1] != "full" {
				pubKey = pubKey[:6]
			}
			fmt.Printf("Public key %x:", pubKey)
			if currencies, ok := store.WalletStore[util.ToBytes(key.PublicKey)]; ok {
				fmt.Print("\n")
				for currency, amount := range currencies {
					fmt.Printf("  - %s: %s\n", currency, amount)
				}
			} else {
				fmt.Print(" no currency\n")
			}
		}
		return nil
	}

	return createCommand("bank", "Display the current balance in each of the local wallets", bankExecutor, nil)
}

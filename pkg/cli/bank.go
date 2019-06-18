package cli

import (
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/util"
	"github.com/CodethinkLabs/wago/pkg/wallet"
)

func BankCommand(args []string, store *wallet.WalletStore) {
	walletFile := wallet.ReadWallet()
	if len(walletFile) == 0 {
		println("No key pair in wallet.")
		return
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
				fmt.Printf("  - %s: %d.%d\n", currency, amount.Value, amount.Decimal)
			}
		} else {
			fmt.Print(" no currency\n")
		}
	}
}

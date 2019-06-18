package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"strings"
)

func DeleteCommand(args []string) {
	if len(args) < 2 {
		panic("must provide address")
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
		println("No keyPair match.")
	} else {
		walletFile = walletFile[:i]
	}
	wallet.WriteWallet(walletFile)
}

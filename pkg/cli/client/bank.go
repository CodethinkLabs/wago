package client

import (
	"context"
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/proto"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"golang.org/x/crypto/ed25519"
)

// executes the bank command, returning to the user the
// current balance of all their local wallets
// syntax: bank [full]
func BankCommand(ctx context.Context, client proto.WalletServiceClient) cli.Command {
	bankExecutor := func(args []string) error {
		walletFile, err := wallet.ReadWallet()
		if err != nil {
			return err
		}
		if len(walletFile) == 0 {
			fmt.Println("No credentials in wallet.")
			return nil
		}
		response, err := client.GetBalance(ctx, &proto.BalanceRequest{PublicKeys: convertKeys(walletFile.GetKeys())})

		if err != nil {
			panic(err)
		}

		for pubKey, w := range response.Wallets {
			if len(args) == 1 || args[1] != "full" {
				pubKey = pubKey[:6]
			}
			fmt.Printf("Public key %s:", pubKey)
			if len(w.Currencies) > 0 {
				fmt.Print("\n")
				for currency, amount := range w.Currencies {
					fmt.Printf("  - %s: %s\n", currency, wallet.DecimalAmount{Value: amount.Value, Decimal: int8(amount.Decimal)})
				}
			} else {
				fmt.Print(" no currency\n")
			}
		}
		return nil
	}
	return cli.CreateCommand("bank", "Display the current balance in each of the local wallets", bankExecutor, nil)
}

func convertKeys(keys []ed25519.PublicKey) [][]byte {
	converted := make([][]byte, 0)
	for _, key := range keys {
		converted = append(converted, []byte(key))
	}
	return converted
}

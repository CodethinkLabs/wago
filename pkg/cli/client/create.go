package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/proto"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/crypto/ed25519"
)

type commandContext struct {
	Address  ed25519.PublicKey
	Amount   wallet.DecimalAmount
	Currency wallet.Currency
	Password string
}

// CreateCommand creates the create command, generating currency in
// the provided wallet address. Requires a server-set password.
//
// syntax: create ${ADDRESS} ${AMOUNT} ${CURRENCY} ${PASSWORD}
func CreateCommand(context context.Context, client proto.WalletServiceClient) cli.Command {
	createExecutor := func(args []string) error {
		walletFile, err := wallet.ReadWallet()
		if err != nil {
			return err
		}

		ctx := commandContext{}

		switch len(args) {
		default:
			fallthrough
		case 5: // password
			ctx.Password = args[4]
			fallthrough
		case 4: // currency
			ctx.Currency = wallet.Currency(args[3])
			fallthrough
		case 3: // amount todo(arlyon) decimal numbers a well
			num, _ := strconv.Atoi(args[2])
			ctx.Amount = wallet.DecimalAmount{Value: int64(num)}
			fallthrough
		case 2: // dst address
			if match, ok := walletFile.PrefixSearch(args[1]); ok {
				ctx.Address = match
			}
		case 1:
		case 0:
		}

		createClient, err := client.CreateCurrency(context, &proto.Create{
			Update: &proto.WalletUpdate{
				Dest:     ctx.Address,
				Amount:   &proto.DecimalAmount{Value: ctx.Amount.Value, Decimal: int64(ctx.Amount.Decimal)},
				Currency: string(ctx.Currency),
			},
			Timestamp: ptypes.TimestampNow(),
			Password:  ctx.Password,
		})

		for {
			update, err := createClient.Recv()
			if err != nil {
				return err
			} else if update.Status == proto.TransactionUpdate_COMMITTED {
				return nil
			} else if update.Status == proto.TransactionUpdate_INVALIDATED {
				return fmt.Errorf(update.Message)
			}
		}
	}

	return cli.CreateCommand("create", "Generate currency for a given wallet", createExecutor, nil)
}

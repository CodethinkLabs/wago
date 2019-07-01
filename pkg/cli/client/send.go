package client

import (
	"context"
	"encoding/hex"
	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/proto"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"github.com/c-bata/go-prompt"
	"github.com/golang/protobuf/ptypes"
	"golang.org/x/crypto/ed25519"
	"strconv"
	"strings"
)

// a context object that can be populated
// by commands to provide details of an input
type sendContext struct {
	SrcPublic  ed25519.PublicKey
	SrcPrivate ed25519.PrivateKey
	DstPublic  ed25519.PublicKey
	Value      int64
	Decimal    int8
	Currency   wallet.Currency
}

// executes the send command, allowing the user to
// send currency from one of their wallets to another
// syntax: send ${SRC} ${DST} ${CURRENCY} ${AMOUNT}
func SendCommand(ctx context.Context, client proto.WalletServiceClient) cli.Command {
	sendExecutor := func(args []string) error {
		walletFile, err := wallet.ReadWallet()
		if err != nil {
			return err
		}
		sendCtx := sendContext{}

		switch len(args) {
		default:
			fallthrough
		case 5: // currency
			sendCtx.Currency = wallet.Currency(args[4])
			fallthrough
		case 4: // amount todo(arlyon) decimal numbers a well
			num, _ := strconv.Atoi(args[3])
			sendCtx.Value = int64(num)
			fallthrough
		case 3: // dst address
			if hexBytes, err := hex.DecodeString(args[2]); len(hexBytes) == ed25519.PublicKeySize && err == nil {
				sendCtx.DstPublic = hexBytes
			}
			fallthrough
		case 2: // src address
			if match, ok := walletFile.PrefixSearch(args[1]); ok {
				sendCtx.SrcPublic = match
				sendCtx.SrcPrivate = walletFile.Lookup(match)
			}
		case 1:
		case 0:
		}

		return createTransaction(sendCtx, ctx, client)
	}

	sendCompleter := func(in prompt.Document) []prompt.Suggest {
		walletFile, err := wallet.ReadWallet()
		if err != nil {
			return []prompt.Suggest{}
		}
		suggestions := make([]prompt.Suggest, 0)
		currentCommand := in.TextBeforeCursor()
		args := strings.Split(currentCommand, " ")

		switch len(args) {
		case 2:
			// src
			for _, keyPair := range walletFile {
				suggestions = append(suggestions, prompt.Suggest{Text: hex.EncodeToString(keyPair.PublicKey)[:12]})
			}
		case 3:
			// dst
		case 5:
			// currency
		}

		return prompt.FilterFuzzy(suggestions, in.GetWordBeforeCursor(), true)
	}

	return cli.CreateCommand("send", "Send currency from one of your local wallets", sendExecutor, sendCompleter)
}

func createTransaction(sendCtx sendContext, ctx context.Context, client proto.WalletServiceClient) error {
	trans, err := wallet.NewTransaction(sendCtx.SrcPublic, sendCtx.DstPublic, sendCtx.Currency, wallet.DecimalAmount{Value: sendCtx.Value, Decimal: sendCtx.Decimal}, false)
	if err != nil {
		return err
	}

	err = trans.Sign(sendCtx.SrcPrivate)
	if err != nil {
		return err
	}

	protoTransaction := &proto.Transaction{
		Update: &proto.WalletUpdate{
			Src: sendCtx.SrcPublic,
			Dest: sendCtx.DstPublic,
			Amount: &proto.DecimalAmount{
				Value: sendCtx.Value,
				Decimal: int64(sendCtx.Decimal),
			},
			Currency: string(sendCtx.Currency),
		},
		Timestamp: ptypes.TimestampNow(),
		Sig: trans.Sig[:],
	}

	transactionClient, err := client.SubmitTransaction(ctx, protoTransaction)
	if err != nil {
		return err
	}

	for {
		update, err := transactionClient.Recv()
		if err != nil {
			return err
		} else if update.Status == proto.TransactionUpdate_COMMITTED {
			return nil
		}
	}
}

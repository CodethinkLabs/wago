package server

import (
	"bytes"
	"encoding/hex"
	"github.com/CodethinkLabs/wago/pkg/cli"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"github.com/c-bata/go-prompt"
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
	Amount     wallet.DecimalAmount
	Currency   wallet.Currency
}

// executes the send command, allowing the user to
// send currency from one of their wallets to another
// syntax: send ${SRC} ${DST} ${AMOUNT} ${CURRENCY}
func SendCommand(store *wallet.Store) cli.Command {
	sendExecutor := func(args []string) error {
		walletFile, err := wallet.ReadWallet()
		if err != nil {
			return err
		}

		ctx := sendContext{}

		switch len(args) {
		default:
			fallthrough
		case 5: // currency
			ctx.Currency = wallet.Currency(args[4])
			fallthrough
		case 4: // amount todo(arlyon) decimal numbers a well
			num, _ := strconv.Atoi(args[3])
			ctx.Amount = wallet.DecimalAmount{Value: int64(num)}
			fallthrough
		case 3: // dst address
			if match, ok := store.PrefixSearch(args[2]); ok {
				ctx.DstPublic = match
			} else if hexBytes, err := hex.DecodeString(args[2]); len(hexBytes) == ed25519.PublicKeySize && err == nil {
				ctx.DstPublic = hexBytes
			}
			fallthrough
		case 2: // src address
			if match, ok := walletFile.PrefixSearch(args[1]); ok {
				ctx.SrcPublic = match
				ctx.SrcPrivate = walletFile.Lookup(match)
			}
		case 1:
		case 0:
		}

		return createTransaction(*store, ctx, false)
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
			for publicKey := range store.WalletStore {
				suggestions = append(suggestions, prompt.Suggest{Text: hex.EncodeToString(publicKey[:])[:12]})
			}
		case 5:
			// currency
			sourceKey, _ := hex.DecodeString(args[1])
			for key := range store.WalletStore {
				if !bytes.HasPrefix(key[:], sourceKey) {
					continue
				}
				for curr := range store.WalletStore[key] {
					suggestions = append(suggestions, prompt.Suggest{Text: string(curr)})
				}
			}
		}

		return prompt.FilterFuzzy(suggestions, in.GetWordBeforeCursor(), true)
	}

	return cli.CreateCommand("send", "Send currency from one of your local wallets", sendExecutor, sendCompleter)
}

// executes the create command, allowing the user to
// create currency and deposit into the provided wallet
// syntax: create ${KEY} ${AMOUNT} ${CURRENCY}
func CreateCommand(store *wallet.Store) cli.Command {
	createExecutor := func(args []string) error {
		walletFile, err := wallet.ReadWallet()
		if err != nil {
			return err
		}
		ctx := sendContext{}

		switch len(args) {
		default:
			fallthrough
		case 4: // currency
			ctx.Currency = wallet.Currency(args[3])
			fallthrough
		case 3: // amount todo(arlyon) decimal numbers
			num, _ := strconv.Atoi(args[2])
			ctx.Amount.Value= int64(num)
			fallthrough
		case 2:
			if match, ok := walletFile.PrefixSearch(args[1]); ok {
				ctx.DstPublic = match
			} else if hexBytes, err := hex.DecodeString(args[1]); len(hexBytes) == ed25519.PublicKeySize && err == nil {
				ctx.DstPublic = hexBytes
			}
		case 1:
		case 0:
		}

		return createTransaction(*store, ctx, true)
	}

	createCompleter := func(in prompt.Document) []prompt.Suggest {
		suggestions := make([]prompt.Suggest, 0)

		walletFile, err := wallet.ReadWallet()
		if err != nil {
			return suggestions
		}
		currentCommand := in.TextBeforeCursor()
		args := strings.Split(currentCommand, " ")

		switch len(args) {
		case 2:
			// src
			for _, keyPair := range walletFile {
				suggestions = append(suggestions, prompt.Suggest{Text: hex.EncodeToString(keyPair.PublicKey)[:12]})
			}
		}

		return prompt.FilterFuzzy(suggestions, in.GetWordBeforeCursor(), true)
	}

	return cli.CreateCommand("create", "Create currency", createExecutor, createCompleter)
}

// creates a transaction and proposes it to the cluster
// param create: If this flag is specified, the transaction is a "create" command
// 				 meaning that the currency will be generated from nothing.
//			     Create commands with a source public or private key will error.
func createTransaction(store wallet.Store, ctx sendContext, create bool) error {
	trans, err := wallet.NewTransaction(ctx.SrcPublic, ctx.DstPublic, ctx.Currency, ctx.Amount, create)
	if err != nil {
		return err
	}

	err = trans.Sign(ctx.SrcPrivate)
	if err != nil {
		return err
	}

	err = store.Propose(trans)
	return err // or nil
}

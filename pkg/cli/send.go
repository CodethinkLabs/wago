package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"
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
	Value      int64
	Decimal    int8
	Currency   wallet.Currency
}

// executes the send command, allowing the user to
// send currency from one of their wallets to another
// syntax: send ${SRC} ${DST} ${CURRENCY} ${AMOUNT}
var SendCommand = createCommand("send", sendExecutor, sendCompleter)

// executes the create command, allowing the user to
// create currency and deposit into the provided wallet
// syntax: create ${KEY} ${AMOUNT} ${CURRENCY}
var CreateCommand = createCommand("create", createExecutor, createCompleter)

func sendExecutor(args []string, store *wallet.WalletStore) error {
	walletFile := wallet.ReadWallet()
	ctx := sendContext{}

	switch len(args) {
	default:
		fallthrough
	case 5: // currency
		ctx.Currency = wallet.Currency(args[4])
		fallthrough
	case 4: // amount todo decimal numbers a well
		num, _ := strconv.Atoi(args[3])
		ctx.Value = int64(num)
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

func sendCompleter(in prompt.Document, store *wallet.WalletStore) []prompt.Suggest {
	walletFile := wallet.ReadWallet()
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
			// todo can suggest wrong currencies if keys clash
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

func createExecutor(args []string, store *wallet.WalletStore) error {
	walletFile := wallet.ReadWallet()
	ctx := sendContext{}

	switch len(args) {
	default:
		fallthrough
	case 4: // currency
		ctx.Currency = wallet.Currency(args[3])
		fallthrough
	case 3: // amount todo decimal numbers
		num, _ := strconv.Atoi(args[2])
		ctx.Value = int64(num)
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

func createCompleter(in prompt.Document, store *wallet.WalletStore) []prompt.Suggest {
	walletFile := wallet.ReadWallet()
	suggestions := make([]prompt.Suggest, 0)
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

// creates a transaction and proposes it to the cluster
// param create: If this flag is specified, the transaction is a "create" command
// 				 meaning that the currency will be generated from nothing.
//			     Create commands with a source public or private key will error.
func createTransaction(store wallet.WalletStore, ctx sendContext, create bool) error {
	trans := wallet.NewTransaction(ctx.SrcPublic, ctx.DstPublic, ctx.Currency, wallet.DecimalAmount{Value: ctx.Value, Decimal: ctx.Decimal}, create)

	if !create && ctx.SrcPublic == nil {
		return fmt.Errorf("invalid source address provided")
	} else if !create && len(ctx.SrcPrivate) != ed25519.PrivateKeySize {
		return fmt.Errorf("private key for address %x is the wrong length", ctx.SrcPublic)
	} else if create && len(ctx.SrcPrivate) != 0 && len(ctx.SrcPublic) != 0 {
		// we panic here because the invariant is the fault of the programmer
		panic(fmt.Errorf("source private or public key passed to create command"))
	}

	trans.Sign(ctx.SrcPrivate)
	err := store.Propose(trans)
	return err // or nil
}

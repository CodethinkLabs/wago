package cli

import (
	"bytes"
	"encoding/hex"
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

func createTransaction(store wallet.WalletStore, ctx sendContext, create bool) {
	trans := wallet.NewTransaction(ctx.SrcPublic, ctx.DstPublic, ctx.Currency, wallet.DecimalAmount{Value: ctx.Value, Decimal: ctx.Decimal}, create)
	trans.Sign(ctx.SrcPrivate)

	err := store.Propose(trans)
	if err != nil {
		panic(err)
	}
}

func sendExecutor(args []string, store *wallet.WalletStore) error {
	walletFile := wallet.ReadWallet()
	ctx := sendContext{}

	if len(args) > 1 && args[0] == "send" {
		if match, ok := walletFile.PrefixSearch(args[1]); ok {
			ctx.SrcPublic = match
			ctx.SrcPrivate = walletFile.Lookup(match)
		}
	}
	if len(args) > 2 && args[0] == "send" {
		if match, ok := store.PrefixSearch(args[2]); ok {
			ctx.DstPublic = match
		} else if hexBytes, err := hex.DecodeString(args[2]); len(hexBytes) == ed25519.PublicKeySize && err == nil {
			ctx.DstPublic = hexBytes
		}
	}
	if len(args) > 3 && args[0] == "send" {
		// todo decimal numbers a well
		num, _ := strconv.Atoi(args[3])
		ctx.Value = int64(num)
	}
	if len(args) > 4 && args[0] == "send" {
		ctx.Currency = wallet.Currency(args[4])
	}

	createTransaction(*store, ctx, false)
	return nil
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

func createExecutor(args []string, store *wallet.WalletStore) error {
	walletFile := wallet.ReadWallet()
	ctx := sendContext{}

	if len(args) > 1 && args[0] == "create" {
		if match, ok := walletFile.PrefixSearch(args[1]); ok {
			ctx.DstPublic = match
		} else if hexBytes, err := hex.DecodeString(args[1]); len(hexBytes) == ed25519.PublicKeySize && err == nil {
			ctx.DstPublic = hexBytes
		}
	}

	if len(args) > 2 && args[0] == "create" {
		// todo decimal numbers a well
		num, _ := strconv.Atoi(args[2])
		ctx.Value = int64(num)
	}

	if len(args) > 3 && args[0] == "create" {
		ctx.Currency = wallet.Currency(args[3])
	}

	createTransaction(*store, ctx, true)
	return nil
}

var SendCommand = createCommand("send", sendExecutor, sendCompleter)
var CreateCommand = createCommand("create", createExecutor, createCompleter)

package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"github.com/c-bata/go-prompt"
	"golang.org/x/crypto/ed25519"
	"os"
	"strconv"
	"strings"
)

type commandContext struct {
	SrcPublic  ed25519.PublicKey
	SrcPrivate ed25519.PrivateKey
	DstPublic  ed25519.PublicKey
	Value      int64
	Decimal    int8
	Currency   wallet.Currency
}

// todo make file private
var CommandContext commandContext

func EExecutor(in string, store *wallet.WalletStore) {
	CommandContext = commandContext{}

	in = strings.TrimSpace(in)
	args := strings.Split(in, " ")
	switch args[0] {
	case "exit":
		os.Exit(0)
	case "authenticate":
		if len(args) != 2 {
			wallet.Authenticate(nil)
		} else {
			wallet.Authenticate(&args[1])
		}
	case "bank":
		BankCommand(args, store)
	case "new":
		privateKey, publicKey := wallet.GenerateKeys(args[1:])
		walletFile := wallet.ReadWallet()
		walletFile = wallet.AddKeys(walletFile, publicKey, privateKey)
		wallet.WriteWallet(walletFile)
	case "create":
		walletFile := wallet.ReadWallet()
		ctx := commandContext{}

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
		SendTransactionCommand(*store, ctx, true)
	case "health":
		// todo info about cluster
	case "send":
		walletFile := wallet.ReadWallet()

		if len(args) > 1 && args[0] == "send" {
			if match, ok := walletFile.PrefixSearch(args[1]); ok {
				CommandContext.SrcPublic = match
				CommandContext.SrcPrivate = walletFile.Lookup(match)
			}
		}
		if len(args) > 2 && args[0] == "send" {
			if match, ok := store.PrefixSearch(args[2]); ok {
				CommandContext.DstPublic = match
			} else if hexBytes, err := hex.DecodeString(args[2]); len(hexBytes) == ed25519.PublicKeySize && err == nil {
				CommandContext.DstPublic = hexBytes
			}
		}
		if len(args) > 3 && args[0] == "send" {
			// todo decimal numbers a well
			num, _ := strconv.Atoi(args[3])
			CommandContext.Value = int64(num)
		}
		if len(args) > 4 && args[0] == "send" {
			CommandContext.Currency = wallet.Currency(args[4])
		}
		SendTransactionCommand(*store, CommandContext, false)
	case "delete":
		DeleteCommand(args)
	case "":
		// ignore
	default:
		fmt.Println("Sorry, I don't understand.")
	}
}

func CCompleter(in prompt.Document, store *wallet.WalletStore) []prompt.Suggest {
	commands := []prompt.Suggest{
		{Text: "new", Description: "Adds a new key to the wallet"},
		{Text: "delete", Description: "Deletes a key from the wallet"},
		{Text: "bank", Description: "Gets the currency values for all of your wallets"},
		{Text: "send", Description: "Sends a transaction to the cluster"},
		{Text: "create", Description: "Creates new currency"},
		{Text: "authenticate", Description: "Authenticates the current session"},
		{Text: "health", Description: "Gets information about the current cluster"},
		{Text: "exit", Description: "Quits the application"},
	}

	currentCommand := in.TextBeforeCursor()
	args := strings.Split(currentCommand, " ")

	if strings.HasPrefix(currentCommand, "send") {
		walletFile := wallet.ReadWallet()
		commands = make([]prompt.Suggest, 0)

		switch len(args) {
		case 2:
			// src
			for _, keyPair := range walletFile {
				commands = append(commands, prompt.Suggest{Text: hex.EncodeToString(keyPair.PublicKey)[:12]})
			}
		case 3:
			// dst
			for publicKey := range store.WalletStore {
				commands = append(commands, prompt.Suggest{Text: hex.EncodeToString(publicKey[:])[:12]})
			}
		case 5:
			// currency
			sourceKey, _ := hex.DecodeString(args[1])
			for key := range store.WalletStore {
				// todo insecure if keys clash
				if !bytes.HasPrefix(key[:], sourceKey) {
					continue
				}
				for curr := range store.WalletStore[key] {
					commands = append(commands, prompt.Suggest{Text: string(curr)})
				}
			}
		}

		return prompt.FilterFuzzy(commands, in.GetWordBeforeCursor(), true)
	}

	if strings.HasPrefix(currentCommand, "delete") {
		commands = make([]prompt.Suggest, 0)
		walletFile := wallet.ReadWallet()
		for _, keyPair := range walletFile {
			commands = append(commands, prompt.Suggest{Text: hex.EncodeToString(keyPair.PublicKey)[:12]})
		}

		return prompt.FilterFuzzy(commands, in.GetWordBeforeCursor(), true)
	}

	if strings.HasPrefix(currentCommand, "create") {
		walletFile := wallet.ReadWallet()
		commands = make([]prompt.Suggest, 0)

		switch len(args) {
		case 2:
			// src
			for _, keyPair := range walletFile {
				commands = append(commands, prompt.Suggest{Text: hex.EncodeToString(keyPair.PublicKey)[:12]})
			}
		}

		return prompt.FilterFuzzy(commands, in.GetWordBeforeCursor(), true)
	}

	return prompt.FilterHasPrefix(commands, in.GetWordBeforeCursor(), true)
}

// package cli provides the command line interface
// for the application. There are a few commands
// that are available for use, which are implemented
// in separate files
package cli

import (
	"fmt"
	"github.com/CodethinkLabs/wago/pkg/wallet"
	"github.com/c-bata/go-prompt"
	"os"
	"strings"
)

// an executable command
type Command struct {
	name      string
	executor  func([]string, *wallet.WalletStore) error
	completer func(prompt.Document, *wallet.WalletStore) []prompt.Suggest
}

type Commands []Command

// creates a new command with a given name
// completer may be null
func createCommand(name string, executor func([]string, *wallet.WalletStore) error, completer func(prompt.Document, *wallet.WalletStore) []prompt.Suggest) Command {
	return Command{
		name, executor, completer,
	}
}

var commandList = Commands{
	BankCommand,
	DeleteCommand,
	SendCommand,
	CreateCommand,
	NewCommand,
	AuthCommand,
	createCommand("exit", func(args []string, w *wallet.WalletStore) error { os.Exit(0); return nil }, nil),
}

// gets a range of suggestions for each command in the list
func (c Commands) GenerateSuggestions() []prompt.Suggest {
	var suggestions []prompt.Suggest
	for _, command := range c {
		suggestions = append(suggestions, prompt.Suggest{Text: command.name /* todo desc */ })
	}
	return suggestions
}

// iterates through each registered command until it
// finds a match and runs its executor
func StoreExecutor(in string, store *wallet.WalletStore) {
	in = strings.TrimSpace(in)
	args := strings.Split(in, " ")
	for _, c := range commandList {
		if args[0] == c.name {
			err := c.executor(args, store)
			if err != nil {
				fmt.Printf("Error with previous command: %s\n", err)
			} else {
				return
			}
		}
	}

	switch args[0] {
	case "":
		// ignore
	default:
		fmt.Println("Sorry, I don't understand.")
	}
}

// iterates through each registered command until it
// finds a match and runs its completer
func StoreCompleter(in prompt.Document, store *wallet.WalletStore) []prompt.Suggest {
	currentCommand := in.TextBeforeCursor()
	args := strings.Split(currentCommand, " ")

	for _, c := range commandList {
		if args[0] != c.name {
			continue
		}
		if c.completer == nil {
			return []prompt.Suggest{}
		} else {
			return c.completer(in, store)
		}
	}

	suggestions := commandList.GenerateSuggestions()
	return prompt.FilterHasPrefix(suggestions, in.GetWordBeforeCursor(), true)
}

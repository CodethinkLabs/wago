// package cli provides the command line interface
// for the application. There are a few commands
// that are available for use, which are implemented
// in separate files
package cli

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"os"
	"strings"
)

// provided a list of commands, generates a completer
// and executor to run those commands
func CreateCLI(commands ...Command) (func(in string), func(in prompt.Document) []prompt.Suggest) {

	var commandList Commands = commands

	// iterates through each registered command until it
	// finds a match and runs its executor
	executor := func(in string) {
		in = strings.TrimSpace(in)
		args := strings.Split(in, " ")
		command, err := commandList.Match(args[0])

		if err == nil {
			err := command.executor(args)
			if err != nil {
				fmt.Printf("Error with previous command: %s\n", err)
			}
			return
		}

		switch args[0] {
		case "":
		case "help":
			commandList.GenerateHelp("Available commands:")
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("Sorry, I don't understand.")
		}
	}

	// iterates through each registered command until it
	// finds a match and runs its completer
	completer := func(in prompt.Document) []prompt.Suggest {
		currentCommand := in.TextBeforeCursor()
		args := strings.Split(currentCommand, " ")
		match, err := commandList.Match(args[0])
		if err != nil {
			suggestions := commandList.GenerateSuggestions()
			return prompt.FilterHasPrefix(suggestions, in.GetWordBeforeCursor(), true)
		}
		if match.completer != nil {
			return match.completer(in)
		} else {
			return []prompt.Suggest{}
		}
	}

	return executor, completer
}

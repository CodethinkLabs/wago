// Package cli provides the command line interface
// for the application. There are a few commands
// that are available for use, which are implemented
// in separate files
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
)

// CreateCLI when provided a list of commands, generates a
// Completer and Executor to run those commands
func CreateCLI(commands ...Command) (func(in string), func(in prompt.Document) []prompt.Suggest) {

	var commandList Commands = commands

	// iterates through each registered command until it
	// finds a match and runs its Executor
	executor := func(in string) {
		in = strings.TrimSpace(in)
		args := strings.Split(in, " ")
		command, err := commandList.Match(args[0])

		if err == nil {
			err := command.Executor(args)
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
	// finds a match and runs its Completer
	completer := func(in prompt.Document) []prompt.Suggest {
		currentCommand := in.TextBeforeCursor()
		args := strings.Split(currentCommand, " ")
		match, err := commandList.Match(args[0])
		if err != nil {
			suggestions := commandList.GenerateSuggestions()
			return prompt.FilterHasPrefix(suggestions, in.GetWordBeforeCursor(), true)
		}
		if match.Completer != nil {
			return match.Completer(in)
		}

		return []prompt.Suggest{}
	}

	return executor, completer
}

// StartCLI runs the command line with the
// given executor and completer functions
func StartCLI(executor func(in string), completer func(in prompt.Document) []prompt.Suggest) {
	fmt.Println("Welcome to wago.")
	p := prompt.New(
		executor, completer,
		prompt.OptionTitle("wago Wallet"),
		prompt.OptionPrefixTextColor(prompt.White),
		prompt.OptionSuggestionBGColor(prompt.Purple),
		prompt.OptionDescriptionBGColor(prompt.White),
		prompt.OptionDescriptionTextColor(prompt.Purple),
		prompt.OptionSelectedSuggestionTextColor(prompt.White),
		prompt.OptionSelectedSuggestionBGColor(prompt.DarkBlue),
		prompt.OptionSelectedDescriptionBGColor(prompt.Blue),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionPrefix("$ "),
	)
	p.Run()
}

package cli

import (
	"fmt"

	"github.com/c-bata/go-prompt"
)

// Command encapsulates an executable command
type Command struct {
	name        string
	description string
	Executor    func([]string) error
	Completer   func(prompt.Document) []prompt.Suggest
}

// Commands is a list of executable commands
type Commands []Command

// CreateCommand creates a new command with a given name
//
// completer may be nil
func CreateCommand(name string, description string, executor func([]string) error, completer func(prompt.Document) []prompt.Suggest) Command {
	return Command{name, description, executor, completer}
}

// Match attempts to find a command that matches
// the provided name, returning an error if it
// does not exist
func (c Commands) Match(name string) (Command, error) {
	for _, c := range c {
		if name == c.name {
			return c, nil
		}
	}
	return Command{}, fmt.Errorf("no matching command")
}

// GenerateSuggestions creates a Suggest object
// for each Command in the Commands list
func (c Commands) GenerateSuggestions() []prompt.Suggest {
	var suggestions []prompt.Suggest
	for _, command := range c {
		suggestions = append(suggestions, prompt.Suggest{Text: command.name, Description: command.description})
	}
	return suggestions
}

// GenerateHelp creates a help command for a
// given Commands list
func (c Commands) GenerateHelp(explanationText string) {
	println(explanationText)
	for _, sug := range c.GenerateSuggestions() {
		fmt.Printf(" - %s: %s\n", sug.Text, sug.Description)
	}
}

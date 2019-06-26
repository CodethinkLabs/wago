package cli

import (
	"fmt"
	"github.com/c-bata/go-prompt"
)

// an executable command
type Command struct {
	name        string
	description string
	Executor    func([]string) error
	Completer   func(prompt.Document) []prompt.Suggest
}

type Commands []Command

// creates a new command with a given name
// Completer may be null
func CreateCommand(name string, description string, executor func([]string) error, completer func(prompt.Document) []prompt.Suggest) Command {
	return Command{name, description, executor, completer,}
}

func (c Commands) Match(name string) (Command, error) {
	for _, c := range c {
		if name == c.name {
			return c, nil
		}
	}
	return Command{}, fmt.Errorf("no matching command")
}

// gets a range of suggestions for each command in the list
func (c Commands) GenerateSuggestions() []prompt.Suggest {
	var suggestions []prompt.Suggest
	for _, command := range c {
		suggestions = append(suggestions, prompt.Suggest{Text: command.name, Description: command.description})
	}
	return suggestions
}

func (c Commands) GenerateHelp(explanationText string) {
	println(explanationText)
	for _, sug := range c.GenerateSuggestions() {
		fmt.Printf(" - %s: %s\n", sug.Text, sug.Description)
	}
}

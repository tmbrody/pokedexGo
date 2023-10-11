package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/peterh/liner"
	"github.com/tmbrody/pokedexGo/pokeapi"
	"github.com/tmbrody/pokedexGo/pokecache"
)

var cache *pokecache.Cache
var pokedex map[string]Pokemon
var commands map[string]cliCommand

type Pokemon struct {
	Name           string    `json:"name"`
	BaseExperience int       `json:"base_experience"`
	Height         int       `json:"height"`
	Weight         int       `json:"weight"`
	Stats          []APIStat `json:"stats"`
	Types          []APIType `json:"types"`
}

type APIStat struct {
	BaseStat int  `json:"base_stat"`
	Stat     Stat `json:"stat"`
}

type Stat struct {
	Name string `json:"name"`
}

type APIType struct {
	Type Type `json:"type"`
}

type Type struct {
	Name string `json:"name"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(config *pokeapi.Config, args []string) error
}

func (p *Pokemon) FormatStats() string {
	var statsText string
	for _, apiStat := range p.Stats {
		statsText += fmt.Sprintf("  -%s: %d\n", apiStat.Stat.Name, apiStat.BaseStat)
	}
	return statsText
}

func (p *Pokemon) FormatTypes() string {
	var typesText string
	for _, apiType := range p.Types {
		typesText += fmt.Sprintf("  - %s\n", apiType.Type.Name)
	}
	return typesText
}

func main() {
	pokedex = make(map[string]Pokemon)
	config := &pokeapi.Config{}
	cache = pokecache.NewCache(5 * time.Minute)

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exits the Pokedex",
			callback: func(config *pokeapi.Config, args []string) error {
				return commandExit(line, config, args)
			},
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "explore",
			description: "Explores a location area and displays the Pokemon found",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Tries catching a Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Displays various Pokemon stats",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Displays all caught Pokemon",
			callback:    commandPokedex,
		},
	}

	for {
		cmdText, err := line.Prompt("Pokedex > ")
		if err != nil {
			break
		}

		cmdText = strings.TrimSpace(cmdText)
		if cmdText == "" {
			continue
		}

		line.AppendHistory(cmdText)

		parts := strings.Fields(cmdText)

		if len(parts) == 0 {
			continue
		}

		cmdName := parts[0]
		cmdArgs := parts[1:]

		if cmd, ok := commands[cmdName]; ok {
			err := cmd.callback(config, cmdArgs)
			if err != nil {
				fmt.Printf("\nError executing command: %v\n\n", err)
			}
		} else {
			fmt.Println("\nUnknown command. Type 'help' for a list of commands.")
			fmt.Println()
		}
	}

	cache.Close()
}

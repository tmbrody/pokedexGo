package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tmbrody/pokeapi"
	"github.com/tmbrody/pokecache"
)

var cache *pokecache.Cache

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

var pokedex map[string]Pokemon

type cliCommand struct {
	name        string
	description string
	callback    func(config *pokeapi.Config, args []string) error
}

var commands map[string]cliCommand

func commandHelp(config *pokeapi.Config, args []string) error {
	fmt.Println("\nWelcome to PokedexGo\n\nAvailable commands:")
	fmt.Println()

	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	fmt.Println()

	return nil
}

func commandExit(config *pokeapi.Config, args []string) error {
	fmt.Println("\nExiting the Pokedex")
	os.Exit(0)
	return nil
}

func commandMap(config *pokeapi.Config, args []string) error {
	cacheKey := fmt.Sprintf("locationAreas_%d", pokeapi.GetCurrentOffset(config))

	var locationAreas []pokeapi.LocationArea
	var err error

	if val, found := cache.Get(cacheKey); found {
		err = json.Unmarshal(val, &locationAreas)
		if err != nil {
			return err
		}
	} else {
		locationAreas, err = pokeapi.FetchLocationAreas(config)
		if err != nil {
			return err
		}

		serializedLocationAreas, marshalErr := json.Marshal(locationAreas)
		if marshalErr != nil {
			return marshalErr
		}
		cache.Add(cacheKey, serializedLocationAreas)
	}

	fmt.Println("\nLocation Areas:")

	for _, area := range locationAreas {
		fmt.Println(area.Name)
	}

	pokeapi.UpdateOffsets(config, true)

	return nil
}

func commandMapBack(config *pokeapi.Config, args []string) error {
	pokeapi.UpdateOffsets(config, false)

	if config.Next == nil {
		return fmt.Errorf("already on the first page")
	}

	return commandMap(config, nil)
}

func commandExplore(config *pokeapi.Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: explore <area_name>")
	}

	areaName := args[0]
	cacheKey := fmt.Sprintf("explore_%s", areaName)

	var pokemon []string
	var err error

	if val, found := cache.Get(cacheKey); found {
		err = json.Unmarshal(val, &pokemon)
		if err != nil {
			return err
		}
	} else {
		httpClient := &http.Client{}
		pokemon, err = pokeapi.ExploreLocationArea(httpClient, areaName)
		if err != nil {
			return err
		}

		serializedPokemon, marshalErr := json.Marshal(pokemon)
		if marshalErr != nil {
			return marshalErr
		}
		cache.Add(cacheKey, serializedPokemon)
	}

	fmt.Printf("\nExploring %s...\n", areaName)
	fmt.Println("Found Pokemon:")

	for _, p := range pokemon {
		fmt.Printf(" - %s\n", p)
	}

	return nil
}

func commandCatch(config *pokeapi.Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: catch <pokemon_name>")
	}

	pokemonName := args[0]

	if _, found := pokedex[pokemonName]; found {
		fmt.Printf("You already caught %s!\n", pokemonName)
		return nil
	}

	httpClient := &http.Client{}
	pokeapiPokemon, err := pokeapi.GetPokemonByName(httpClient, pokemonName)
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemonName)

	catchRate := float64(pokeapiPokemon.BaseExperience) / 255.0

	source := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(source)
	randNum := randGen.Float64()

	if randNum <= catchRate {
		var stats []APIStat
		for _, apiStat := range pokeapiPokemon.Stats {
			stats = append(stats, APIStat{
				Stat: Stat{
					Name: apiStat.Stat.Name,
				},
				BaseStat: apiStat.BaseStat,
			})
		}

		var types []APIType
		for _, apiType := range pokeapiPokemon.Types {
			types = append(types, APIType{
				Type: Type{
					Name: apiType.Type.Name,
				},
			})
		}

		pokemon := Pokemon{
			Name:           pokeapiPokemon.Name,
			BaseExperience: pokeapiPokemon.BaseExperience,
			Height:         pokeapiPokemon.Height,
			Weight:         pokeapiPokemon.Weight,
			Stats:          stats,
			Types:          types,
		}
		fmt.Printf("%s was caught!\n", pokemonName)
		fmt.Println("You may now inspect it with the inspect command.")
		pokedex[pokemonName] = pokemon
	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
	}

	return nil
}

func commandInspect(config *pokeapi.Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: inspect <Pokemon_name>")
	}

	pokemonName := args[0]

	if pokemon, found := pokedex[pokemonName]; !found {
		fmt.Println("You haven't caught that Pokemon yet")
		return nil
	} else {
		fmt.Printf("Name: %s\n", pokemon.Name)
		fmt.Printf("Height: %d\n", pokemon.Height)
		fmt.Printf("Weight: %d\n", pokemon.Weight)

		fmt.Println("Stats:")
		fmt.Printf("%s", pokemon.FormatStats())

		fmt.Println("Types:")
		fmt.Printf("%s", pokemon.FormatTypes())
	}

	return nil
}

func commandPokedex(config *pokeapi.Config, args []string) error {
	if len(pokedex) == 0 {
		fmt.Println("You haven't caught any Pokemon yet")
	} else {
		for _, pokemon := range pokedex {
			fmt.Printf(" - %s\n", pokemon.Name)
		}
	}
	return nil
}

func main() {
	pokedex = make(map[string]Pokemon)

	config := &pokeapi.Config{}

	cache = pokecache.NewCache(5 * time.Minute)

	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exits the Pokedex",
			callback:    commandExit,
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

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("Pokedex > ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(input)

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

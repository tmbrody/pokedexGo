package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/peterh/liner"
	"github.com/tmbrody/pokedexGo/pokeapi"
)

func commandHelp(config *pokeapi.Config, args []string) error {
	fmt.Println("\nWelcome to PokedexGo\n\nAvailable commands:")
	fmt.Println()

	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	fmt.Println()

	return nil
}

func commandExit(line *liner.State, config *pokeapi.Config, args []string) error {
	fmt.Println("\nExiting the Pokedex")
	line.Close()
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

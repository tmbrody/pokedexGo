package pokeapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

type LocationAreaResponse struct {
	PokemonEncounters []struct {
		Pokemon Pokemon `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

var currentOffset int

type Config struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
}

type LocationArea struct {
	Name string `json:"name"`
}

func GetPokemonByName(client *http.Client, pokemonName string) (Pokemon, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", pokemonName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Pokemon{}, err
	}

	req.Header.Add("User-Agent", "PokedexGo")

	resp, err := client.Do(req)
	if err != nil {
		return Pokemon{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Pokemon{}, fmt.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	var pokemon Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&pokemon); err != nil {
		return Pokemon{}, err
	}

	return pokemon, nil
}

func ExploreLocationArea(client *http.Client, areaId string) ([]string, error) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s/", areaId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "PokedexGo")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	var locationAreaResp LocationAreaResponse
	if err := json.NewDecoder(resp.Body).Decode(&locationAreaResp); err != nil {
		return nil, err
	}

	pokemonNames := make([]string, len(locationAreaResp.PokemonEncounters))
	for i, encounter := range locationAreaResp.PokemonEncounters {
		pokemonNames[i] = encounter.Pokemon.Name
	}

	return pokemonNames, nil
}

func FetchLocationAreas(config *Config) ([]LocationArea, error) {
	if config.Next == nil {
		config.Next = stringPointer("https://pokeapi.co/api/v2/location-area?limit=20&offset=0")
	}

	url := config.Next

	if url == nil {
		return nil, fmt.Errorf("no more pages available")
	}

	resp, err := http.Get(*url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var locationAreas struct {
		Results []LocationArea `json:"results"`
	}
	err = json.NewDecoder(resp.Body).Decode(&locationAreas)
	if err != nil {
		return nil, err
	}

	return locationAreas.Results, nil
}

func UpdateOffsets(config *Config, forward bool) {
	if forward {
		currentOffset += 20
	} else {
		currentOffset -= 40
	}

	if currentOffset < 0 {
		currentOffset = 20
		config.Next = nil
	} else {
		configNextString := fmt.Sprintf("https://pokeapi.co/api/v2/location-area?limit=20&offset=%d", currentOffset)
		config.Next = stringPointer(configNextString)
	}
}

func GetCurrentOffset(config *Config) int {
	return currentOffset
}

func stringPointer(s string) *string {
	return &s
}

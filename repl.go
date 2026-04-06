package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"

	"github.com/mathias-schnell/godex/internal/pokecache"
)

// a struct that represents a location-area from the PokeAPI
type location_area struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

// a struct that represents the encounter information for a location-area from the PokeAPI
type encounter_info struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

// a struct that represents the information for a Pokemon from the PokeAPI
type pokemon_info struct {
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
}

// a struct that holds the cache and api URL information for the REPL
type config struct {
	pokedex  map[string]pokemon_info
	cache    *pokecache.Cache
	apiURL   string
	Next     string
	Previous string
}

// a struct that represents a command in the REPL
type cliCommand struct {
	name        string
	description string
	callback    func(c *config, args []string) error
}

// a map that holds the valid commands for the REPL
var validCommands = make(map[string]cliCommand)

// a function that initializes the validCommands map with the available commands
func init() {
	validCommands["catch"] = cliCommand{
		name:        "catch",
		description: "Attempt to catch a Pokemon",
		callback:    commandCatch,
	}
	validCommands["exit"] = cliCommand{
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	}
	validCommands["explore"] = cliCommand{
		name:        "explore",
		description: "Explore the area with the given name and display a list of Pokemon that can be encountered there",
		callback:    commandExplore,
	}
	validCommands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	}
	validCommands["inspect"] = cliCommand{
		name:        "inspect",
		description: "Inspect the details of a Pokemon that you've caught",
		callback:    commandInspect,
	}
	validCommands["map"] = cliCommand{
		name:        "map",
		description: "Displays the current 20 areas",
		callback:    commandMap,
	}
	validCommands["mapb"] = cliCommand{
		name:        "mapb",
		description: "Displays the previous 20 areas",
		callback:    commandMapBack,
	}
	validCommands["pokedex"] = cliCommand{
		name:        "pokedex",
		description: "Displays the list of Pokemon you've caught",
		callback:    commandPokedex,
	}
}

// a function that takes a string input and returns a slice of strings with the cleaned input
func cleanInput(test string) []string {
	return strings.Fields(strings.ToLower(test))
}

// a function that handles the "catch" command
func commandCatch(c *config, args []string) error {
	// Check if the user provided the name of a Pokemon to catch
	if len(args) < 1 {
		return fmt.Errorf("Please provide the name of a Pokemon to catch")
	}

	// Print a message that a Pokeball is being thrown at the specified Pokemon
	fmt.Println("Throwing a Pokeball at " + args[0] + "...")

	// If all is well, construct the URL for the pokemon endpoint
	// Setup the data struct and make a GET request to the PokeAPI using the constructed URL
	pokemonName := args[0]
	url := fmt.Sprintf(c.apiURL+"pokemon/%s/", pokemonName)
	pokemon_info := pokemon_info{}
	if err := getPokeapiData(url, &pokemon_info, c); err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	catch_attempt := rand.Float64() * 100.0
	catch_threshold := math.Pow(math.Log(float64(pokemon_info.BaseExperience)), 2) * 2.0
	if catch_attempt > catch_threshold {
		fmt.Printf("%s was caught!\n", pokemon_info.Name)
		fmt.Printf("You may now inspect %s with the 'inspect' command.\n", pokemon_info.Name)
		c.pokedex[pokemon_info.Name] = pokemon_info
	} else {
		fmt.Printf("%s escaped!\n", pokemon_info.Name)
	}

	return nil
}

// a function that handles the "exit" command
func commandExit(c *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	defer os.Exit(0)
	return nil
}

// a function that handles the "explore" command
func commandExplore(c *config, args []string) error {
	// Check if the user provided an area name as an argument
	if len(args) < 1 {
		return fmt.Errorf("Please provide an area name to explore")
	}

	// If all is well, construct the URL for the location-area endpoint
	// Setup the data struct and make a GET request to the PokeAPI using the constructed URL
	areaName := args[0]
	url := fmt.Sprintf(c.apiURL+"location-area/%s/", areaName)
	encounter_info := encounter_info{}
	if err := getPokeapiData(url, &encounter_info, c); err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// Start text output with a message indicating the area being explored
	fmt.Printf("Exploring %s...\n", args[0])
	fmt.Printf("Found Pokemon:\n")

	// Loop through the PokemonEncounters in the encounter_info struct
	// Print the name of each Pokemon that can be encountered in this specified area
	for _, encounter := range encounter_info.PokemonEncounters {
		fmt.Println(" - " + encounter.Pokemon.Name)
	}

	return nil
}

// a function that handles the "help" command
func commandHelp(c *config, args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	for _, cmd := range validCommands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

// a function that handles the "inspect" command
func commandInspect(c *config, args []string) error {
	// Check if the user provided the name of a Pokemon to inspect
	if len(args) < 1 {
		return fmt.Errorf("Please provide the name of a Pokemon to inspect")
	}

	pokemonName := args[0]
	if pokemon, ok := c.pokedex[pokemonName]; ok {
		fmt.Printf("Name: %s\n", pokemon.Name)
		fmt.Printf("Height: %d\n", pokemon.Height)
		fmt.Printf("Weight: %d\n", pokemon.Weight)
		fmt.Printf("Stats:\n")
		for _, s := range pokemon.Stats {
			fmt.Printf(" -%s: %d\n", s.Stat.Name, s.BaseStat)
		}
		fmt.Printf("Types:\n")
		for _, t := range pokemon.Types {
			fmt.Printf(" - %s\n", t.Type.Name)
		}
	} else {
		fmt.Println("You have not caught that Pokemon")
	}

	return nil
}

// a function that handles the "map" command
func commandMap(c *config, args []string) error {
	// Check if the Next URL is empty
	// If it is, we're just starting the REPL and need to initialize it
	if c.Next == "" {
		c.Next = c.apiURL + "location-area/"
	}

	// Setup the data struct and make a GET request to the PokeAPI using the Next URL
	location_areas := location_area{}
	if err := getPokeapiData(c.Next, &location_areas, c); err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// Update the config with the next and previous URLs for pagination
	c.Next = location_areas.Next
	if previous, ok := location_areas.Previous.(string); ok {
		c.Previous = previous
	} else {
		c.Previous = ""
	}

	// Print the names of the location_areas in the current page
	for _, loc := range location_areas.Results {
		fmt.Println(loc.Name)
	}

	return nil
}

// a function that handles the "mapb" command
func commandMapBack(c *config, args []string) error {
	// Check if the Previous URL is empty, which means we're on the first page
	if c.Previous == "" {
		fmt.Println("You're on the first page")
		return nil
	}

	// Setup the data struct and make a GET request to the PokeAPI using the Previous URL
	location_areas := location_area{}
	if err := getPokeapiData(c.Previous, &location_areas, c); err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// Update the config with the next and previous URLs for pagination
	c.Next = location_areas.Next
	if previous, ok := location_areas.Previous.(string); ok {
		c.Previous = previous
	} else {
		c.Previous = ""
	}

	// Print the names of the location_areas in the current page
	for _, loc := range location_areas.Results {
		fmt.Println(loc.Name)
	}

	return nil
}

func getPokeapiData(url string, poke_struct any, c *config) error {
	// Check if the data at this URL is in the cache
	if val, ok := c.cache.Get(url); ok {
		err := json.Unmarshal(val, &poke_struct)
		if err != nil {
			return err
		}
	} else {
		// Make a GET request to the PokeAPI using the constructed URL
		res, err := http.Get(url)
		if err != nil {
			return err
		}

		// Read the response body and unmarshal it into the Pokeapi data struct
		// Add the response body to the cache before unmarshalling it
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return err
		}
		c.cache.Add(url, body)

		//Unmarshal the response body into the Pokeapi data struct
		err = json.Unmarshal(body, &poke_struct)
		if err != nil {
			return err
		}
	}
	return nil
}

// a function that handles the "pokedex" command
func commandPokedex(c *config, args []string) error {
	if len(c.pokedex) == 0 {
		fmt.Println("You haven't caught any Pokemon yet!")
		return nil
	}

	fmt.Println("Your Pokedex:")
	for name := range c.pokedex {
		fmt.Printf(" - %s\n", name)
	}
	return nil
}

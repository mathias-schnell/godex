package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mathias-schnell/godex/internal/pokecache"
)

// a struct that represents the response from the PokeAPI
type location_area struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type encounter_info struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

// a struct that holds the next and previous URLs for pagination
type config struct {
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
	validCommands["help"] = cliCommand{
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
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
}

// a function that takes a string input and returns a slice of strings with the cleaned input
func cleanInput(test string) []string {
	return strings.Fields(strings.ToLower(test))
}

// a function that handles the "exit" command
func commandExit(c *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	defer os.Exit(0)
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

// a function that handles the "map" command
func commandMap(c *config, args []string) error {
	// Check if the Next URL is empty
	// If it is, we're just starting the REPL and need to initialize it
	if c.Next == "" {
		c.Next = c.apiURL + "location-area/"
	}

	// Check if the Next URL is in the cache
	// if it is, use the cached response instead of making a new GET request
	if val, ok := c.cache.Get(c.Next); ok {
		location_areas := location_area{}
		err := json.Unmarshal(val, &location_areas)
		if err != nil {
			return err
		}

		for _, loc := range location_areas.Results {
			fmt.Println(loc.Name)
		}
		return nil
	}

	// Make a GET request to the PokeAPI using the Next URL from the config
	res, err := http.Get(c.Next)
	if err != nil {
		return err
	}

	// Read the response body and unmarshal it into the location_area struct
	// Add the response body to the cache before unmarshalling it
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	c.cache.Add(c.Next, body)

	//	Unmarshal the response body into the location_area struct
	location_areas := location_area{}
	err = json.Unmarshal(body, &location_areas)
	if err != nil {
		return err
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
		fmt.Println("you're on the first page")
		return nil
	}

	// Check if the Previous URL is in the cache
	// if it is, use the cached response instead of making a new GET request
	if val, ok := c.cache.Get(c.Previous); ok {
		location_areas := location_area{}
		err := json.Unmarshal(val, &location_areas)
		if err != nil {
			return err
		}

		for _, loc := range location_areas.Results {
			fmt.Println(loc.Name)
		}
		return nil
	}

	// Make a GET request to the Previous URL to get the previous page of location_areas
	res, err := http.Get(c.Previous)
	if err != nil {
		return err
	}

	// Read the response body and unmarshal it into the location_area struct
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}

	//	Unmarshal the response body into the location_area struct
	location_areas := location_area{}
	err = json.Unmarshal(body, &location_areas)
	if err != nil {
		return err
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

// a function that handles the "explore" command
func commandExplore(c *config, args []string) error {
	// Check if the user provided an area name as an argument
	if len(args) < 1 {
		return fmt.Errorf("please provide an area name to explore")
	}

	// If all is well, construct the URL for the location-area endpoint
	areaName := args[0]
	url := fmt.Sprintf(c.apiURL+"location-area/%s/", areaName)

	// Check if the data at this URL is in the cache
	if val, ok := c.cache.Get(url); ok {
		encounter_info := encounter_info{}
		err := json.Unmarshal(val, &encounter_info)
		if err != nil {
			return err
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

	// Make a GET request to the PokeAPI using the constructed URL
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	// Read the response body and unmarshal it into the encounter_info struct
	// Add the response body to the cache before unmarshalling it
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}
	c.cache.Add(url, body)

	//Unmarshal the response body into the encounter_info struct
	encounter_info := encounter_info{}
	err = json.Unmarshal(body, &encounter_info)
	if err != nil {
		return err
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

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// pokeapi is a struct that represents the response from the PokeAPI
type pokeapi struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

// config is a struct that holds the next and previous URLs for pagination
type config struct {
	Next     string
	Previous string
}

// cliCommand is a struct that represents a command in the REPL
type cliCommand struct {
	name        string
	description string
	callback    func(c *config) error
}

// validCommands is a map that holds the valid commands for the REPL
var validCommands = make(map[string]cliCommand)

// init is a function that initializes the validCommands map with the available commands
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

// cleanInput is a function that takes a string input
// and returns a slice of strings with the cleaned input
func cleanInput(test string) []string {
	return strings.Fields(strings.ToLower(test))
}

// commandExit is a function that handles the "exit" command
func commandExit(c *config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	defer os.Exit(0)
	return nil
}

// commandHelp is a function that handles the "help" command
func commandHelp(c *config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Printf("Usage:\n\n")
	for _, cmd := range validCommands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

// commandMap is a function that handles the "map" command
func commandMap(c *config) error {
	// Make a GET request to the PokeAPI using the Next URL from the config
	res, err := http.Get(c.Next)
	if err != nil {
		return err
	}

	// Read the response body and unmarshal it into the pokeapi struct
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}

	//	Unmarshal the response body into the pokeapi struct
	location_areas := pokeapi{}
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

// commandMapBack is a function that handles the "mapb" command
func commandMapBack(c *config) error {
	// Check if the Previous URL is empty, which means we're on the first page
	if c.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}

	// Make a GET request to the Previous URL to get the previous page of location_areas
	res, err := http.Get(c.Previous)
	if err != nil {
		return err
	}

	// Read the response body and unmarshal it into the pokeapi struct
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return err
	}

	//	Unmarshal the response body into the pokeapi struct
	location_areas := pokeapi{}
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

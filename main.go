package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/mathias-schnell/godex/internal/pokecache"
)

func main() {
	c := &config{
		pokedex:  make(map[string]pokemon_info),
		cache:    pokecache.NewCache(5 * time.Minute),
		apiURL:   "https://pokeapi.co/api/v2/",
		Next:     "",
		Previous: "",
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		text := cleanInput(scanner.Text())
		if len(text) == 0 {
			continue
		}
		command := text[0]
		args := text[1:]

		cmd, ok := validCommands[command]
		if !ok {
			fmt.Println("Unknown command")
			continue
		}

		err := cmd.callback(c, args)
		if err != nil {
			fmt.Println("Error executing command:", err)
		}
	}
}

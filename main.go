package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	c := &config{
		Next:     "https://pokeapi.co/api/v2/location-area/",
		Previous: "",
	}
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()
		text := cleanInput(scanner.Text())
		switch text[0] {
		case "exit":
			err := validCommands["exit"].callback(c)
			if err != nil {
				fmt.Println("Error executing command:", err)
			}
		case "help":
			err := validCommands["help"].callback(c)
			if err != nil {
				fmt.Println("Error executing command:", err)
			}
		case "map":
			err := validCommands["map"].callback(c)
			if err != nil {
				fmt.Println("Error executing command:", err)
			}
		case "mapb":
			err := validCommands["mapb"].callback(c)
			if err != nil {
				fmt.Println("Error executing command:", err)
			}
		default:
			fmt.Printf("Unknown command")
		}
	}
}

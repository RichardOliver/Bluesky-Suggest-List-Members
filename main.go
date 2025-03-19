package main

import (
	"flag"
	"fmt"
)

func main() {
	// Read CLI arguments
	username := flag.String("username", "", "Bluesky handle")
	password := flag.String("password", "", "Bluesky App Password")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("❌ Error: --username and --password are required.")
		return
	}

	fmt.Printf("Fetching lists for user: %s\n", *username)
}

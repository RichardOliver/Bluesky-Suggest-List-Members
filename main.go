package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	var verbose bool
	var jsonOutput bool
	var version bool
	username := flag.String("username", "", "Bluesky handle")
	listName := flag.String("list", "", "List name (optional)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")
	flag.BoolVar(&jsonOutput, "json", false, "Enable JSON output")
	flag.BoolVar(&jsonOutput, "j", false, "Enable JSON output (shorthand)")
	flag.BoolVar(&version, "version", false, "Print version information")
	flag.Parse()

	if version {
		fmt.Println("followsTools v0.1.0")
		return
	}

	if *username == "" {
		fmt.Println("❌ Error: --username is required.")
		fmt.Println("Usage:")
		fmt.Printf("%s [flags] [arguments]\n", os.Args[0]) // os.Args[0] is the program name.
		fmt.Println("Flags:")
		flag.PrintDefaults()
		return
	}

	decentralizedIdentifier, err := resolveHandle(*username)
	if err != nil {
		log.Fatal("❌ Authentication failed:", err)
	}

	lists, err := getLists(decentralizedIdentifier)
	if err != nil {
		log.Fatal("❌ Failed:", err)
	}

	if *listName == "" {
		fmt.Println("📋", *username, "'s lists:")
		for _, v := range lists {
			fmt.Println("-", v.Name)
		}
		return
	}

	var listUri string
	for _, v := range lists {
		if strings.EqualFold(v.Name, *listName) {
			listUri = v.Uri
			break
		}
	}

	if listUri == "" {
		log.Fatal("❌ Error: List not found:", *listName)
	}

	members, err := getListMembers(listUri)
	if err != nil {
		log.Fatal("❌ Failed to retrieve list members:", err)
	}

	fmt.Println("👥 Users in list:", *listName)

	followCount := make(map[Follower]int)

	for _, handle := range members {
		follows, err := getFollows(handle)
		if err != nil {
			log.Fatal("❌ Failed to retrieve follows:", err)
		}

		for _, follow := range follows {
			if followCount[follow] == 0 {
				followCount[follow] = 1
			} else {
				followCount[follow]++
			}
		}
	}

	usersFollows, err := getFollows(*username)
	if err != nil {
		log.Fatal("❌ Failed to retrieve user's follows:", err)
	}
	// TODO: improve the performace of filterFollows
	// TODO: add a flag for verbose output
	// TODO: add a flag to parallelize the requests
	followCount = filterFollows(followCount, 1, members, usersFollows)
	sortedList := sortFollowCount(followCount)
	if jsonOutput {
		outputJSON(sortedList)
		return
	} else {
		outputText(sortedList)
		return
	}
}

type UserWithCount struct {
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}

func outputJSON(sortedList KeyValueList) {
	// Convert to a slice of UserWithCount
	var userList []UserWithCount
	for _, kv := range sortedList {
		userList = append(userList, UserWithCount{
			Handle:      kv.Key.Handle,
			DisplayName: kv.Key.DisplayName,
			Description: kv.Key.Description,
			Count:       kv.Value,
		})
	}

	// Output as JSON
	jsonData, err := json.MarshalIndent(userList, "", "  ")
	if err != nil {
		log.Fatal("❌ Failed to marshal JSON:", err)
	}
	fmt.Println(string(jsonData))
}

func outputText(sortedList KeyValueList) {
	// Output as text
	for _, kv := range sortedList {
		if kv.Value > 1 {
			fmt.Printf("(%d) %q - %q\n", kv.Value, kv.Key.DisplayName, kv.Key.Handle)
			fmt.Println("      ", kv.Key.Description)
			fmt.Println()
		}
	}
}

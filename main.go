package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	username := flag.String("username", "", "Bluesky handle")
	listName := flag.String("list", "", "List name (optional)")
	flag.Parse()

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
			// Skip if the follow is already in the list (pehaps refactor this to a function)
			//if follow.Handle == "bsky.app" || stringInSliceIgnoreCase(follow.Handle, members) {
			//	continue
			//}

			if followCount[follow] == 0 {
				followCount[follow] = 1
			} else {
				followCount[follow]++
			}
		}
	}

	for follower, count := range followCount {
		if count > 1 {
			fmt.Printf("(%d) %q - %q\n", count, follower.DisplayName, follower.Handle)
			fmt.Println("      ", follower.Description)
			fmt.Println()
		}
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const baseURL = "https://public.api.bsky.app/xrpc/"
const resolveHandleEndpoint = "com.atproto.identity.resolveHandle"

type resolveHandleResponse struct {
	Did string `json:"did"`
}

func resolveHandle(username string) (string, error) {

	parsedURL, err := url.Parse(baseURL + resolveHandleEndpoint)
	if err != nil {
		return "", err
	}

	queryParams := url.Values{}
	queryParams.Add("handle", username)
	parsedURL.RawQuery = queryParams.Encode()
	fullUrl := parsedURL.String()

	var results resolveHandleResponse

	err = doRequest(http.MethodGet, fullUrl, nil, &results)
	if err != nil {
		return "", err
	}

	return results.Did, nil
}

const listsEndpoint = "app.bsky.graph.getLists"

type ListsResponse struct {
	Cursor string     `json:"cursor"`
	Lists  []ListItem `json:"lists"`
}

type ListItem struct {
	Name string `json:"name"`
	Uri  string `json:"uri"`
}

func getLists(username string) ([]ListItem, error) {
	parsedURL, err := url.Parse(baseURL + listsEndpoint)
	if err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	queryParams.Add("actor", username)
	parsedURL.RawQuery = queryParams.Encode()
	fullUrl := parsedURL.String()

	var results ListsResponse
	// TODO: Refactor this to loop with the cursor
	err = doRequest(http.MethodGet, fullUrl, nil, &results)
	if err != nil {
		return nil, err
	}

	return results.Lists, nil
}

const listMembersEndpoint = "app.bsky.graph.getList"

type ListMembersResponse struct {
	Items []ListMember `json:"items"`
}

type ListMember struct {
	Subject struct {
		Handle string `json:"handle"`
	} `json:"subject"`
}

func getListMembers(listUri string) ([]string, error) {
	parsedURL, err := url.Parse(baseURL + listMembersEndpoint)
	if err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	queryParams.Add("list", listUri)
	parsedURL.RawQuery = queryParams.Encode()
	fullUrl := parsedURL.String()

	var results ListMembersResponse
	// TODO: Refactor this to loop with the cursor
	err = doRequest(http.MethodGet, fullUrl, nil, &results)
	if err != nil {
		return nil, err
	}

	var handles []string
	for _, item := range results.Items {
		handles = append(handles, item.Subject.Handle)
	}

	return handles, nil
}

const followsEndpoint = "app.bsky.graph.getFollows"

type FollowsResponse struct {
	Follows []Follower `json:"follows"`
	Cursor  string     `json:"cursor"`
}

type Follower struct {
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

func getFollows(username string) ([]Follower, error) {
	var allFollows []Follower
	cursor := ""

	parsedURL, err := url.Parse(baseURL + followsEndpoint)
	if err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	queryParams.Add("actor", username)

	for {
		if cursor != "" {
			if queryParams.Get("cursor") != "" {
				queryParams.Set("cursor", cursor)
			} else {
				queryParams.Add("cursor", cursor)
			}
		}

		parsedURL.RawQuery = queryParams.Encode()
		fullUrl := parsedURL.String()

		var results FollowsResponse
		// TODO: Refactor this to loop with the cursor
		err = doRequest(http.MethodGet, fullUrl, nil, &results)
		if err != nil {
			return nil, err
		}

		allFollows = append(allFollows, results.Follows...)

		if results.Cursor == "" {
			break
		}
		cursor = results.Cursor

	}

	return allFollows, nil
}

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

package main

import (
	"encoding/json"
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

	queryParams := url.Values{}
	queryParams.Add("handle", username)

	parsedURL, err := url.Parse(baseURL + resolveHandleEndpoint)
	if err != nil {
		return "", err
	}

	parsedURL.RawQuery = queryParams.Encode()
	fullURL := parsedURL.String()

	client := &http.Client{}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch user %s: status %d", username, response.StatusCode)
	}

	var results resolveHandleResponse
	if err := json.NewDecoder(response.Body).Decode(&results); err != nil {
		return "", err
	}

	return results.Did, nil
}

const listsEndpoint = "app.bsky.graph.getLists"

type ListsResponse struct {
	Lists []ListItem `json:"lists"`
}

// Struct to parse authentication response
type ListItem struct {
	Name string `json:"name"`
	Uri  string `json:"uri"`
}

func getLists(username string) ([]ListItem, error) {

	queryParams := url.Values{}
	queryParams.Add("actor", username)

	parsedURL, err := url.Parse(baseURL + listsEndpoint)
	if err != nil {
		return nil, err
	}
	parsedURL.RawQuery = queryParams.Encode()
	fullURL := parsedURL.String()

	client := &http.Client{}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status code:", response.StatusCode)
	}
	var results ListsResponse
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&results)
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
	queryParams := url.Values{}
	queryParams.Add("list", listUri)

	parsedURL, err := url.Parse(baseURL + listMembersEndpoint)
	if err != nil {
		return nil, err
	}
	parsedURL.RawQuery = queryParams.Encode()
	fullURL := parsedURL.String()

	client := &http.Client{}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Println("Request failed with status code:", response.StatusCode)
	}

	var results ListMembersResponse
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&results)
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

	client := &http.Client{}

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
		} else if queryParams.Get("cursor") != "" {
			queryParams.Del("cursor")
		}

		parsedURL.RawQuery = queryParams.Encode()
		fullURL := parsedURL.String()

		req, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("Accept", "application/json")

		response, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("request failed with status code: %d", response.StatusCode)
		}

		var results FollowsResponse
		decoder := json.NewDecoder(response.Body)
		err = decoder.Decode(&results)
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

func stringInSliceIgnoreCase(a string, list []string) bool {
	for _, b := range list {
		if strings.EqualFold(a, b) {
			return true
		}
	}
	return false
}

func incrementOrAdd(followerCount *[]followerWithCount, follower Follower) {
	for i, pair := range *followerCount {
		if pair.Follower == follower {
			(*followerCount)[i].Count++
			return
		}
	}
	*followerCount = append(*followerCount, followerWithCount{Follower: follower, Count: 1})
}

type followerWithCount struct {
	Follower Follower
	Count    int
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
			if follow.Handle == "bsky.app" || stringInSliceIgnoreCase(follow.Handle, members) {
				continue
			}

			if followCount[follow] == 0 {
				followCount[follow] = 1
			} else {
				followCount[follow]++
			}
		}
	}

	fmt.Println("Gopher's Diner Breakfast Menu")
	for follower, count := range followCount {
		if count > 1 {
			fmt.Printf("(%d) %q - %q\n", count, follower.DisplayName, follower.Handle)
			fmt.Println("      ", follower.Description)
			fmt.Println()
		}
	}
}

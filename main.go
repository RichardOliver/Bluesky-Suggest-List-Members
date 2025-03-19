package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const baseURL = "https://bsky.social/xrpc/"
const authEndpoint = "com.atproto.server.createSession"

type AuthRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type AuthResponse struct {
	AccessJwt string `json:"accessJwt"`
	Handle    string `json:"handle"`
}

func authenticate(username, password string) (string, error) {
	data := AuthRequest{
		Identifier: username,
		Password:   password,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(baseURL+authEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to authenticate: %s", body)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", err
	}

	return authResp.AccessJwt, nil
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

func getLists(username string, authToken string) ([]ListItem, error) {

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

	req.Header.Add("Authorization", "Bearer "+authToken)
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

func getListMembers(listUri, authToken string) ([]string, error) {
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

	req.Header.Add("Authorization", "Bearer "+authToken)
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
	Follows []struct {
		Handle      string `json:"handle"`
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
	} `json:"follows"`
	Cursor string `json:"cursor"`
}

func getFollows(username, authToken string) ([]string, error) {
	var allFollows []string
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

		req.Header.Add("Authorization", "Bearer "+authToken)
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

		for _, user := range results.Follows {
			allFollows = append(allFollows, user.Handle)
		}

		if results.Cursor == "" {
			break
		}
		cursor = results.Cursor
	}

	return allFollows, nil
}

func main() {
	username := flag.String("username", "", "Bluesky handle")
	password := flag.String("password", "", "Bluesky App Password")
	listName := flag.String("list", "", "List name (optional)")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("❌ Error: --username and --password are required.")
		fmt.Println("Usage:")
		fmt.Printf("%s [flags] [arguments]\n", os.Args[0]) // os.Args[0] is the program name.
		fmt.Println("Flags:")
		flag.PrintDefaults()
		return
	}

	authToken, err := authenticate(*username, *password)
	if err != nil {
		log.Fatal("❌ Authentication failed:", err)
	}

	lists, err := getLists(*username, authToken)
	if err != nil {
		log.Fatal("Failed:", err)
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

	// Get members of the selected list
	members, err := getListMembers(listUri, authToken)
	if err != nil {
		log.Fatal("Failed to retrieve list members:", err)
	}

	// Print members
	fmt.Println("👥 Users in list:", *listName)
	for _, handle := range members {
		fmt.Println("-", handle, "follows:")

		follows, err := getFollows(handle, authToken)
		if err != nil {
			log.Fatal("Failed to retrieve follows:", err)
		}

		for _, follow := range follows {
			fmt.Println("  -", follow)
		}
	}
}

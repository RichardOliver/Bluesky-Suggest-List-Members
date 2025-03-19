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
)

const authURL = "https://bsky.social/xrpc/com.atproto.server.createSession"

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

	resp, err := http.Post(authURL, "application/json", bytes.NewBuffer(jsonData))
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

const listsURL = "https://bsky.social/xrpc/app.bsky.graph.getLists"

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

	parsedURL, err := url.Parse(listsURL)
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

func main() {
	username := flag.String("username", "", "Bluesky handle")
	password := flag.String("password", "", "Bluesky App Password")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("❌ Error: --username and --password are required.")
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

	for _, v := range lists {
		fmt.Println(v.Name)
	}
}

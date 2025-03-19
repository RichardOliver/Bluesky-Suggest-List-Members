package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
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

func main() {
	username := flag.String("username", "", "Bluesky handle")
	password := flag.String("password", "", "Bluesky App Password")
	flag.Parse()

	if *username == "" || *password == "" {
		fmt.Println("❌ Error: --username and --password are required.")
		return
	}

	token, err := authenticate(*username, *password)
	if err != nil {
		log.Fatal("❌ Authentication failed:", err)
	}

	fmt.Println("Your access token:", token)
}

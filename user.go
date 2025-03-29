package main

import (
	"net/http"
	"net/url"
)

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

const baseURL = "https://public.api.bsky.app/xrpc/"

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

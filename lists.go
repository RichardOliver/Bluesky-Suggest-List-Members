package main

import (
	"net/http"
	"net/url"
)

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

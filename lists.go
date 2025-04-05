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
	var allLists []ListItem
	cursor := ""

	parsedURL, err := url.Parse(baseURL + listsEndpoint)
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

		var results ListsResponse

		err = doRequest(http.MethodGet, fullUrl, nil, &results)
		if err != nil {
			return nil, err
		}

		allLists = append(allLists, results.Lists...)

		if results.Cursor == "" {
			break
		}
		cursor = results.Cursor
	}

	return allLists, nil
}

const listMembersEndpoint = "app.bsky.graph.getList"

type ListMembersResponse struct {
	Cursor string       `json:"cursor"`
	Items  []ListMember `json:"items"`
}

type ListMember struct {
	Subject struct {
		Handle string `json:"handle"`
	} `json:"subject"`
}

func getListMembers(listUri string) ([]string, error) {
	var allHandles []string
	cursor := ""

	parsedURL, err := url.Parse(baseURL + listMembersEndpoint)
	if err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	queryParams.Add("list", listUri)

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

		var results ListMembersResponse

		err = doRequest(http.MethodGet, fullUrl, nil, &results)
		if err != nil {
			return nil, err
		}

		for _, item := range results.Items {
			allHandles = append(allHandles, item.Subject.Handle)
		}

		if results.Cursor == "" {
			break
		}
		cursor = results.Cursor
	}

	return allHandles, nil
}

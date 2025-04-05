package main

import (
	"slices"
	"sort"
)

type KeyValue struct {
	Key   Follower
	Value int
}

type KeyValueList []KeyValue

// Implement sort.Interface for KeyValueList (Len, Swap, Less)
func (p KeyValueList) Len() int      { return len(p) }
func (p KeyValueList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Descending order by Value
func (p KeyValueList) Less(i, j int) bool {
	return p[i].Value > p[j].Value
}

func sortFollowCount(followCount map[Follower]int) []KeyValue {
	// Convert map to slice of KeyValue
	var kvList KeyValueList
	for k, v := range followCount {
		kvList = append(kvList, KeyValue{k, v})
	}

	// Sort the slice
	sort.Sort(kvList)

	return kvList
}

func filterFollows(followCount map[Follower]int, minCount int, existingListMembers []string, usersFollows []Follower) map[Follower]int {
	//TODO: refactor this to use a single outer loop
	followCount = filterMinCount(followCount, minCount)
	followCount = filterExistingListMembers(followCount, existingListMembers)
	followCount = filterUsersFollows(followCount, usersFollows)
	followCount = filterDefaultAccounts(followCount)
	return followCount
}

func filterMinCount(followCount map[Follower]int, minCount int) map[Follower]int {
	for k, v := range followCount {
		if v < minCount {
			delete(followCount, k)
		}
	}
	return followCount
}
func filterExistingListMembers(followCount map[Follower]int, existingListMembers []string) map[Follower]int {
	for k := range followCount {
		if slices.Contains(existingListMembers, k.Handle) {
			delete(followCount, k)
		}
	}
	return followCount
}
func filterUsersFollows(followCount map[Follower]int, usersFollows []Follower) map[Follower]int {
	for k := range followCount {
		for _, follower := range usersFollows {
			if k.Handle == follower.Handle {
				delete(followCount, k)
				continue
			}
		}
	}
	return followCount
}
func filterDefaultAccounts(followCount map[Follower]int) map[Follower]int {
	for k := range followCount {
		if k.Handle == "bsky.app" {
			delete(followCount, k)
		}
	}
	return followCount
}

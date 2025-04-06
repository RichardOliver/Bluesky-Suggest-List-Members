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

	for k, v := range followCount {
		if isDefaultAccount(k.Handle) ||
			isBelowMinCount(v, minCount) ||
			isExistingMemberOfList(k.Handle, existingListMembers) ||
			isAlreadyFollowedByUser(k.Handle, usersFollows) {

			delete(followCount, k)
		}
	}

	return followCount
}

func isDefaultAccount(handle string) bool {
	return handle == "bsky.app"
}

func isBelowMinCount(count, minCount int) bool {
	return count <= minCount
}

func isExistingMemberOfList(handle string, existingListMembers []string) bool {
	return slices.Contains(existingListMembers, handle)
}

func isAlreadyFollowedByUser(handle string, usersFollows []Follower) bool {
	for _, follower := range usersFollows {
		if handle == follower.Handle {
			return true
		}
	}
	return false
}

package main

import (
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

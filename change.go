package main

import (
	"fmt"
)

type ChangeType int

const (
	NewElement ChangeType = iota
	Changed
)

type Change struct {
	tp  ChangeType
	old []string
	new []string
}

func (ch Change) ToHTMLString(CourseName string) string {
	switch ch.tp {
	case NewElement:
		return fmt.Sprintf("%s\n🔆 <b>New:</b> %s %s", CourseName, ch.new[0], ch.new[4])
	case Changed:
		return fmt.Sprintf("%s\n❇️ <b>Changes</b> in %s\nOld: %s\nNew: %s", CourseName, ch.old[0], ch.old[4], ch.new[4])
	default:
		return "Unknown change type"
	}
}

func Compare(old, new [][]string) []Change {
	mp := map[string][]string{}
	for _, s := range old {
		mp[s[0]] = s
	}

	var changes []Change
	for _, s := range new {
		old, ok := mp[s[0]]
		if !ok {
			changes = append(changes, Change{
				tp:  NewElement,
				new: s,
			})
		} else {
			for i := range old {
				if old[i] != s[i] {
					changes = append(changes, Change{
						tp:  Changed,
						old: old,
						new: s,
					})
					break
				}
			}
		}
	}

	return changes
}

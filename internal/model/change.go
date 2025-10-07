package model

import (
	"fmt"
)

type ChangeType int

const (
	NewElement ChangeType = iota
	Changed
)

type Change struct {
	TP         ChangeType
	CourseName string
	Old        []string
	New        []string
}

func (ch Change) ToHTMLString() string {
	switch ch.TP {
	case NewElement:
		return fmt.Sprintf("%s\n🔆 <b>New:</b> %s %s", ch.CourseName, ch.New[0], ch.New[4])
	case Changed:
		return fmt.Sprintf("%s\n❇️ <b>Changes</b> in %s\nOld: %s\nNew: %s", ch.CourseName, ch.Old[0], ch.Old[4], ch.New[4])
	default:
		return "Unknown change type"
	}
}

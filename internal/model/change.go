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
	Old        *GradeRow
	New        *GradeRow
}

func (ch Change) ToHTMLString() string {
	switch ch.TP {
	case NewElement:
		return fmt.Sprintf("%s\n🔆 <b>New:</b> %s",
			ch.CourseName, ch.New.StringWithName())
	case Changed:
		return fmt.Sprintf("%s\n❇️ <b>Changes</b> in %s\nOld: %s\nNew: %s",
			ch.CourseName, ch.Old.AssName,
			ch.Old.StringWithoutName(), ch.New.StringWithoutName())
	default:
		return "Unknown change type"
	}
}

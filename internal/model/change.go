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
	var s string
	switch ch.TP {
	case NewElement:
		s = fmt.Sprintf("%s\n🔆 <i>New:</i> %s",
			ch.CourseName, ch.New.StringWithName())
	case Changed:
		s = fmt.Sprintf("%s\n❇️ <i>Changes</i> in %s\nOld: <s>%s</s>\nNew: %s",
			ch.CourseName, ch.Old.AssName,
			ch.Old.StringWithoutName(), ch.New.StringWithoutName())
	default:
		panic("unknown change type")
	}

	if ch.New.Feedback != "" {
		s += fmt.Sprintf("\n<i>Feedback:</i> %s", ch.New.Feedback)
	}

	return s
}

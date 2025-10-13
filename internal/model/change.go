package model

import (
	"fmt"
	"strings"
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
		return fmt.Sprintf("%s\n🔆 <b>New:</b> %s %s (%s)",
			ch.CourseName, mname(ch.New),
			perc(ch.New), score(ch.New))
	case Changed:
		return fmt.Sprintf("%s\n❇️ <b>Changes</b> in %s\nOld: %s (%s)\nNew: %s (%s)",
			ch.CourseName, mname(ch.Old),
			perc(ch.Old), score(ch.Old),
			perc(ch.New), score(ch.New))
	default:
		return "Unknown change type"
	}
}

func mname(row []string) string {
	return row[0]
}
func perc(row []string) string {
	return strings.ReplaceAll(row[4], " ", "")
}

func score(row []string) string {
	got := row[2]
	from := row[3]
	if splited := strings.Split(from, "–"); len(splited) == 2 {
		from = strings.TrimSpace(splited[1])
	}

	return got + "/" + from
}

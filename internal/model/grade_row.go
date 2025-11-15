package model

import (
	"fmt"
	"strings"
)

type GradeRow struct {
	AssName    string
	Percentage string
	Score      string
	Rang       string
	Feedback   string
	Raw        []string
}

func NewGradeRow(raw []string) *GradeRow {
	if len(raw) < 5 {
		panic("raw must have at least 5 elements")
	}
	return &GradeRow{
		AssName:    raw[0],
		Percentage: raw[4],
		Score:      raw[2],
		Rang:       raw[3],
		Feedback:   raw[5],
		Raw:        raw,
	}
}

func (gr *GradeRow) ToStringSlice() []string {
	return gr.Raw
}
func (gr *GradeRow) StringWithoutName() string {
	return fmt.Sprintf("%s (%s)", TrimWhiteSpace(gr.Percentage), gr.ScoreWithSlash())
}

func (gr *GradeRow) StringWithName() string {
	return fmt.Sprintf("%s %s (%s)", gr.AssName, TrimWhiteSpace(gr.Percentage), gr.ScoreWithSlash())
}

func (gr *GradeRow) ScoreWithSlash() string {
	got := gr.Score
	from := gr.Rang
	if splited := strings.Split(from, "–"); len(splited) == 2 {
		from = TrimWhiteSpace(splited[1])
	}

	return got + "/" + from
}

//go:inline
func TrimWhiteSpace(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

func (gr *GradeRow) IsEqual(other *GradeRow) bool {
	return gr.AssName == other.AssName && gr.Percentage == other.Percentage && gr.Score == other.Score && gr.Rang == other.Rang
}

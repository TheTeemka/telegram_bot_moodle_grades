package main

import (
	"encoding/csv"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Parser struct {
	LastTimeParsed time.Time
	fetcher        *MoodleFetcher
}

func NewParser(fetcher *MoodleFetcher) *Parser {
	return &Parser{
		fetcher: fetcher,
	}
}
func (p *Parser) parseAndCompare(output chan<- string) {
	err := p.fetcher.Login()
	if err != nil {
		panic(err)
	}
	slog.Info("Login successful")

	buf, err := p.fetcher.GetGradesPage()
	if err != nil {
		panic(err)
	}

	links, err := ExtractGradesLinks(buf)
	if err != nil {
		panic(err)
	}
	slog.Info("Successfully extracted links", "len", len(links))

	for _, link := range links {
		buf, err := p.fetcher.Fetch(link)
		if err != nil {
			slog.Error("Failed to fetch grade page", "link", link, "error", err)
		}

		courseName, newItems, err := ExtractItems(buf)
		if err != nil {
			slog.Error("Failed to extract items", "link", link, "error", err)
			continue
		}

		oldItems, err := readItems(courseName)
		exists := !errors.Is(err, os.ErrNotExist)
		if err != nil && exists {
			slog.Error("Failed to read old items", "course", courseName, "error", err)
			continue
		}

		if exists {
			changes := Compare(oldItems, newItems)
			for _, c := range changes {
				output <- c.ToHTMLString(courseName)
			}
		}

		writeItems(courseName, newItems)
	}
	p.LastTimeParsed = time.Now()

}

func (p *Parser) GetLastTimeParsed() time.Time {
	return p.LastTimeParsed
}
func buildFilePath(courseName string) string {
	return "csvs/" + sanitizeFilename(courseName) + "_grades.csv"
}

func sanitizeFilename(name string) string {
	// Replace common invalid filename characters with underscores
	invalidChars := []string{":", "/", "\\", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}
	// Trim whitespace and limit length to avoid issues
	name = strings.TrimSpace(name)
	if len(name) > 255 {
		name = name[:255]
	}
	return name
}

func writeItems(courseName string, items [][]string) error {
	file, err := os.OpenFile(buildFilePath(courseName), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	for _, item := range items {
		writer.Write(item)
	}
	writer.Flush()
	return file.Close()
}

func readItems(courseName string) ([][]string, error) {
	file, err := os.Open(buildFilePath(courseName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

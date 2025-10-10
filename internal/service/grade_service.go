package service

import (
	"encoding/csv"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/TheTeemka/telegram_bot_moodle_grades/internal/model"
)

var ErrInProgress = errors.New("❗️ already in progress")

type GradeService struct {
	LastTimeParsed time.Time
	fetcher        *MoodleFetcher
	isRunning      atomic.Bool
}

func NewGradeService(fetcher *MoodleFetcher) *GradeService {
	return &GradeService{
		fetcher: fetcher,
	}
}
func (p *GradeService) ParseAndCompare() ([]model.Change, error) {
	if !p.isRunning.CompareAndSwap(false, true) {
		slog.Debug("ParseAndCompare:already_running")
		return nil, ErrInProgress
	}
	defer p.isRunning.Store(false)

	buf, err := p.fetcher.GetGradesPage()
	if err != nil {
		return nil, err
	}

	links, err := extractGradesLinks(buf)
	if err != nil {
		return nil, err
	}
	slog.Debug("Successfully extracted links", "len", len(links))

	var TotalChanges []model.Change
	for _, link := range links {
		slog.Debug("Processing link", "link", link)
		buf, err := p.fetcher.Fetch(link)
		if err != nil {
			slog.Error("Failed to fetch grade page", "link", link, "error", err)
		}

		courseName, newItems, err := extractItems(buf)
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
			CourseChanges := Compare(courseName, oldItems, newItems)
			slog.Debug("Course changes found", "course", courseName, "count", len(CourseChanges))
			TotalChanges = append(TotalChanges, CourseChanges...)
		}

		err = writeItems(courseName, newItems)
		if err != nil {
			slog.Error("Failed to write new items", "course", courseName, "error", err)
		}
	}
	p.LastTimeParsed = time.Now()
	slog.Debug("ParseAndCompare:done", "total_changes", len(TotalChanges))
	return TotalChanges, nil
}

func (p *GradeService) GetLastTimeParsed() time.Time {
	slog.Debug("GetLastTimeParsed", "last", p.LastTimeParsed)
	return p.LastTimeParsed
}
func buildFilePath(courseName string) string {
	path := "csvs/" + sanitizeFilename(courseName) + "_grades.csv"
	return path
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
	slog.Debug("sanitizeFilename", "result", name)
	return name
}

func writeItems(courseName string, items [][]string) error {
	slog.Debug("writeItems", "course", courseName, "items", len(items))
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
	slog.Debug("readItems", "course", courseName)
	file, err := os.Open(buildFilePath(courseName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

func Compare(courseName string, old, new [][]string) []model.Change {
	slog.Debug("Compare:start", "course", courseName, "old", len(old), "new", len(new))
	mp := map[string][]string{}
	for _, s := range old {
		mp[s[0]] = s
	}

	var changes []model.Change
	for _, s := range new {
		old, ok := mp[s[0]]
		if !ok {
			changes = append(changes, model.Change{
				CourseName: courseName,
				TP:         model.NewElement,
				New:        s,
			})
		} else {
			for i := range old {
				if old[i] != s[i] {
					changes = append(changes, model.Change{
						CourseName: courseName,
						TP:         model.Changed,
						Old:        old,
						New:        s,
					})
					break
				}
			}
		}
	}

	return changes
}
